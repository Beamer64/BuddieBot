package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/commands/prefix"
	"github.com/Beamer64/BuddieBot/pkg/commands/slash"
	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/database"
	"github.com/Beamer64/BuddieBot/pkg/events"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/Beamer64/BuddieBot/pkg/lavalink_runner"
	"github.com/Beamer64/BuddieBot/pkg/voice_chat"
	"github.com/Beamer64/bb_data"
	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/snowflake/v2"
)

var (
	botSession *discordgo.Session
	linkClient disgolink.Client
	devRunner  *lavalink_runner.Runner
	db         *database.DB
)

func Init(cfg *config.Configs) error {
	token := ""
	botENV := ""
	if helper.IsLaunchedByDebugger() {
		token = cfg.Keys.TestBotToken
		botENV = "BB Test is ready for commands!"
	} else {
		token = cfg.Keys.ProdBotToken
		botENV = "BuddieBot is ready for commands!"
	}

	// Open DB before any process spawn — a bad path should fail fast.
	log.Printf("opening database at %s", cfg.Database.Path)
	dbConn, err := database.Open(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	db = dbConn
	cfg.DB = dbConn

	// Seed audio-enabled rows for the master and test guilds. Idempotent —
	// EnsureGuildExists doesn't touch existing rows, so admin disables
	// via /admin audio survive restarts. Backwards compat for the old
	// hardcoded IsAudioGuild check.
	seedCtx, seedCancel := context.WithTimeout(context.Background(), 5*time.Second)
	for _, gid := range []string{cfg.DiscordIDs.MasterGuildID, cfg.DiscordIDs.TestGuildID} {
		if gid == "" {
			continue
		}
		if err := dbConn.EnsureGuildExists(seedCtx, gid, true); err != nil {
			seedCancel()
			closeDB()
			return fmt.Errorf("seed guild %s: %w", gid, err)
		}
	}
	seedCancel()

	if helper.IsLaunchedByDebugger() {
		log.Println("spawning Lavalink (dev)…")
		lavalinkDir, ok := findLavalinkDir()
		if !ok {
			closeDB()
			return fmt.Errorf("Lavalink.jar not found in ./lavalink/, ../lavalink/, or ../../lavalink/ — see README dev setup")
		}
		readyURL := fmt.Sprintf("http://%s:%s/version", cfg.Lavalink.Host, cfg.Lavalink.Port)
		runner, err := lavalink_runner.Start(filepath.Join(lavalinkDir, "Lavalink.jar"), lavalinkDir, readyURL, cfg.Lavalink.Password, 90*time.Second)
		if err != nil {
			closeDB()
			return fmt.Errorf("dev lavalink: %w", err)
		}
		devRunner = runner
		log.Println("Lavalink ready")
	}

	discordgo.Logger = filteredDiscordLogger

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		stopDevRunner()
		closeDB()
		return fmt.Errorf("failed to create Discord session: %w", err)
	}
	botSession = session

	if err := bb_data.Load(); err != nil {
		stopDevRunner()
		closeDB()
		return fmt.Errorf("failed to load bb_data datasets: %w", err)
	}

	user, err := fetchSelfUserWithRetry(session)
	if err != nil {
		stopDevRunner()
		closeDB()
		return fmt.Errorf("failed to grab Discord session User: %w", err)
	}

	botUserID, err := snowflake.Parse(user.ID)
	if err != nil {
		stopDevRunner()
		closeDB()
		return fmt.Errorf("parse bot user id: %w", err)
	}

	linkClient = disgolink.New(botUserID)
	nodeCtx, nodeCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer nodeCancel()
	log.Printf("connecting to Lavalink at %s:%s", cfg.Lavalink.Host, cfg.Lavalink.Port)
	if _, err := linkClient.AddNode(
		nodeCtx, disgolink.NodeConfig{
			Name:     "main",
			Address:  cfg.Lavalink.Host + ":" + cfg.Lavalink.Port,
			Password: cfg.Lavalink.Password,
		},
	); err != nil {
		stopDevRunner()
		closeDB()
		return fmt.Errorf("connect to lavalink node: %w", err)
	}

	player := voice_chat.New(linkClient, session)
	cfg.Player = player
	linkClient.AddListeners(
		disgolink.NewListenerFunc(player.OnTrackEnd),
		disgolink.NewListenerFunc(player.OnTrackException),
	)

	registerEvents(session, cfg, user)
	registerVoiceForwarders(session, linkClient)

	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	if err = session.Open(); err != nil {
		stopDevRunner()
		closeDB()
		return fmt.Errorf("failed to open Discord session: %w", err)
	}

	if err := registerCommands(session); err != nil {
		return fmt.Errorf("failed to register commands: %w", err)
	}

	log.Println(botENV)
	return nil
}

// Shutdown closes the session, the disgolink client, the database, and
// (in dev) the Lavalink child process. Safe to call once at process exit.
func Shutdown() {
	if botSession != nil {
		_ = botSession.Close()
	}
	if linkClient != nil {
		linkClient.Close()
	}
	stopDevRunner()
	closeDB()
}

func stopDevRunner() {
	if devRunner != nil {
		devRunner.Stop()
		devRunner = nil
	}
}

func closeDB() {
	if db != nil {
		_ = db.Close()
		db = nil
	}
}

// findLavalinkDir searches the three candidate paths because cwd differs
// between `go run`, IDE debug launches, and built binaries.
func findLavalinkDir() (string, bool) {
	for _, dir := range []string{"./lavalink", "../lavalink", "../../lavalink"} {
		if _, err := os.Stat(filepath.Join(dir, "Lavalink.jar")); err == nil {
			return dir, true
		}
	}
	return "", false
}

// fetchSelfUserWithRetry retries User("@me") through Cloudflare 503s and
// post-restart 429s that would otherwise kill startup. Honors Retry-After
// when present; bails fast on 4xx auth errors.
func fetchSelfUserWithRetry(s *discordgo.Session) (*discordgo.User, error) {
	const maxAttempts = 3
	const maxWait = 30 * time.Second
	backoffs := []time.Duration{time.Second, 3 * time.Second}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		user, err := s.User("@me")
		if err == nil {
			return user, nil
		}
		lastErr = err

		if !shouldRetrySelfUser(err) {
			return nil, err
		}
		if attempt == maxAttempts {
			break
		}

		wait := backoffs[attempt-1]
		if hint := retryAfter(err); hint > 0 {
			wait = hint
		}
		if wait > maxWait {
			wait = maxWait
		}
		log.Printf("retrying User(@me) in %s (attempt %d/%d): %v", wait, attempt+1, maxAttempts, err)
		time.Sleep(wait)
	}
	return nil, lastErr
}

// shouldRetrySelfUser — 4xx auth/path errors won't improve; 5xx/429/network might.
func shouldRetrySelfUser(err error) bool {
	var restErr *discordgo.RESTError
	if !errors.As(err, &restErr) || restErr.Response == nil {
		return true
	}
	switch restErr.Response.StatusCode {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return false
	}
	return true
}

// retryAfter parses integer-seconds Retry-After; HTTP-date form is rare
// here and not supported.
func retryAfter(err error) time.Duration {
	var restErr *discordgo.RESTError
	if !errors.As(err, &restErr) || restErr.Response == nil {
		return 0
	}
	v := restErr.Response.Header.Get("Retry-After")
	if v == "" {
		return 0
	}
	secs, parseErr := strconv.Atoi(v)
	if parseErr != nil {
		return 0
	}
	return time.Duration(secs) * time.Second
}

func registerEvents(s *discordgo.Session, cfg *config.Configs, u *discordgo.User) {
	// Session
	s.AddHandler(events.NewReadyHandler(cfg).ReadyHandler)

	// Gateway state markers around any heartbeat-error spam — lets us see
	// at a glance whether a reconnect actually happened.
	s.AddHandler(
		func(_ *discordgo.Session, _ *discordgo.Connect) {
			log.Println("discordgo: gateway connected")
		},
	)
	s.AddHandler(
		func(_ *discordgo.Session, _ *discordgo.Disconnect) {
			log.Println("discordgo: gateway disconnected (auto-reconnect pending)")
		},
	)
	s.AddHandler(
		func(_ *discordgo.Session, _ *discordgo.Resumed) {
			log.Println("discordgo: gateway session resumed")
		},
	)

	// Guild
	guildHandler := events.NewGuildHandler(cfg)
	s.AddHandler(guildHandler.GuildCreateHandler)
	s.AddHandler(guildHandler.GuildDeleteHandler)
	s.AddHandler(guildHandler.GuildJoinHandler)
	s.AddHandler(guildHandler.GuildLeaveHandler)

	// Messages
	s.AddHandler(events.NewMessageCreateHandler(cfg, u).MessageCreateHandler)
	s.AddHandler(events.NewReactionHandler(cfg, u).ReactHandlerAdd)

	// Commands
	s.AddHandler(events.NewCommandHandler(cfg).CommandHandler)
}

// filteredDiscordLogger drops discordgo's "websocket: close sent" spam
// (~1×/s after a TCP abort) — the original error and the Disconnect/Resumed
// markers already cover the recovery window.
func filteredDiscordLogger(msgL, caller int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if strings.Contains(msg, "websocket: close sent") {
		return
	}
	pc, file, line, _ := runtime.Caller(caller)
	if idx := strings.LastIndexAny(file, "/\\"); idx >= 0 {
		file = file[idx+1:]
	}
	name := runtime.FuncForPC(pc).Name()
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[idx+1:]
	}
	log.Printf("[DG%d] %s:%d:%s() %s\n", msgL, file, line, name, msg)
}

// registerVoiceForwarders pipes discordgo voice gateway events to disgolink,
// which uses them to drive its own (DAVE-capable) voice WebSocket.
func registerVoiceForwarders(s *discordgo.Session, link disgolink.Client) {
	s.AddHandler(
		func(_ *discordgo.Session, e *discordgo.VoiceServerUpdate) {
			guildID, err := snowflake.Parse(e.GuildID)
			if err != nil {
				log.Printf("voice forwarder: parse guild id %q: %v", e.GuildID, err)
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			link.OnVoiceServerUpdate(ctx, guildID, e.Token, e.Endpoint)
		},
	)

	s.AddHandler(
		func(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
			if e.VoiceState == nil || s.State == nil || s.State.User == nil {
				return
			}
			// Only forward updates for the bot's own voice state.
			if e.UserID != s.State.User.ID {
				return
			}
			guildID, err := snowflake.Parse(e.GuildID)
			if err != nil {
				log.Printf("voice forwarder: parse guild id %q: %v", e.GuildID, err)
				return
			}
			var chID *snowflake.ID
			if e.ChannelID != "" {
				cID, err := snowflake.Parse(e.ChannelID)
				if err != nil {
					log.Printf("voice forwarder: parse channel id %q: %v", e.ChannelID, err)
					return
				}
				chID = &cID
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			link.OnVoiceStateUpdate(ctx, guildID, chID, e.SessionID)
		},
	)
}

func registerCommands(s *discordgo.Session) error {
	log.Println("Updating commands")

	// Bot session needs to settle before BulkOverwrite accepts the call.
	time.Sleep(3 * time.Second)
	commandsRegistered, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", slash.Commands)
	if err != nil {
		return err
	}

	topLevel := len(commandsRegistered)
	subCmds := 0
	cmdChoices := 0
	for _, cmd := range commandsRegistered {
		subCmds += countSubCommands(cmd.Options)
		cmdChoices += countCommandChoices(cmd.Options)
	}
	prefixCmds := len(prefix.Names)

	log.Printf("%d Top-level commands\n", topLevel)
	log.Printf("%d Command-option choices (e.g. /get type:joke)\n", cmdChoices)
	log.Printf("%d Sub-commands\n", subCmds)
	log.Printf("%d $-prefix commands\n", prefixCmds)
	log.Printf("%d Total features\n", topLevel+cmdChoices+subCmds+prefixCmds)
	return nil
}

// countSubCommands recurses into SubCommandGroups so nested leaves
// (e.g. effects under /image filter) are counted.
func countSubCommands(opts []*discordgo.ApplicationCommandOption) int {
	count := 0
	for _, opt := range opts {
		switch opt.Type {
		case discordgo.ApplicationCommandOptionSubCommand:
			count++
		case discordgo.ApplicationCommandOptionSubCommandGroup:
			count += countSubCommands(opt.Options)
		}
	}
	return count
}

// countCommandChoices counts choices on `type`-named options — those act
// as dispatchers (/get type:joke), distinct from parameter-only choices.
func countCommandChoices(opts []*discordgo.ApplicationCommandOption) int {
	count := 0
	for _, opt := range opts {
		if opt.Name == "type" && len(opt.Choices) > 0 {
			count += len(opt.Choices)
		}
	}
	return count
}
