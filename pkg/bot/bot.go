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

	if helper.IsLaunchedByDebugger() {
		log.Println("spawning Lavalink (dev)…")
		lavalinkDir, ok := findLavalinkDir()
		if !ok {
			return fmt.Errorf("Lavalink.jar not found in ./lavalink/, ../lavalink/, or ../../lavalink/ — see README dev setup")
		}
		readyURL := fmt.Sprintf("http://%s:%s/version", cfg.Lavalink.Host, cfg.Lavalink.Port)
		runner, err := lavalink_runner.Start(filepath.Join(lavalinkDir, "Lavalink.jar"), lavalinkDir, readyURL, cfg.Lavalink.Password, 90*time.Second)
		if err != nil {
			return fmt.Errorf("dev lavalink: %w", err)
		}
		devRunner = runner
		log.Println("Lavalink ready")
	}

	discordgo.Logger = filteredDiscordLogger

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		stopDevRunner()
		return fmt.Errorf("failed to create Discord session: %w", err)
	}
	botSession = session

	if err := bb_data.Load(); err != nil {
		stopDevRunner()
		return fmt.Errorf("failed to load bb_data datasets: %w", err)
	}

	user, err := fetchSelfUserWithRetry(session)
	if err != nil {
		stopDevRunner()
		return fmt.Errorf("failed to grab Discord session User: %w", err)
	}

	botUserID, err := snowflake.Parse(user.ID)
	if err != nil {
		stopDevRunner()
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
		return fmt.Errorf("failed to open Discord session: %w", err)
	}

	if err := registerCommands(session); err != nil {
		return fmt.Errorf("failed to register commands: %w", err)
	}

	log.Println(botENV)
	return nil
}

// Shutdown closes the Discord session, the Lavalink client, and (in dev)
// the Lavalink Java child process. Safe to call once at process exit.
func Shutdown() {
	if botSession != nil {
		_ = botSession.Close()
	}
	if linkClient != nil {
		linkClient.Close()
	}
	stopDevRunner()
}

func stopDevRunner() {
	if devRunner != nil {
		devRunner.Stop()
		devRunner = nil
	}
}

// findLavalinkDir locates the dev Lavalink directory regardless of the
// process's working directory. Mirrors the config loader's multi-path
// approach since the cwd differs between `go run`, IDE debug launches,
// and built binaries.
func findLavalinkDir() (string, bool) {
	for _, dir := range []string{"./lavalink", "../lavalink", "../../lavalink"} {
		if _, err := os.Stat(filepath.Join(dir, "Lavalink.jar")); err == nil {
			return dir, true
		}
	}
	return "", false
}

// fetchSelfUserWithRetry calls User("@me") with bounded retries. Discord's
// REST endpoint occasionally returns 503 (Cloudflare "no healthy upstream")
// transiently, and 429s (rate limit) are common after restart loops —
// without a retry a single blip kills bot startup. The retry honors the
// Retry-After header when Discord supplies one and bails fast on auth
// errors that won't get better.
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

// shouldRetrySelfUser reports whether a User("@me") error is worth retrying.
// 4xx auth/path errors won't get better with retries; everything else
// (network errors, 5xx, 429) might.
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

// retryAfter extracts an integer-seconds Retry-After header from a Discord
// REST error, returning 0 if absent or unparseable. HTTP-date Retry-After
// values aren't parsed — they're rare from Discord/Cloudflare for this
// endpoint.
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

	// Gateway state observability — brackets any heartbeat-error spam with
	// clear "disconnected" / "resumed" markers so we can tell at a glance
	// whether the reconnect actually happened.
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

// filteredDiscordLogger mirrors discordgo's default logger format but drops
// the "websocket: close sent" line that the heartbeat goroutine emits ~1×/s
// after a TCP-level connection abort. The original abort error still logs,
// and the Disconnect/Resumed handlers bracket the recovery window — the
// repeating close-sent line adds no signal.
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

// registerVoiceForwarders bridges discordgo's voice gateway events to the
// disgolink client. Lavalink uses these to open and maintain its own
// voice WebSocket (DAVE-capable) — the bot itself never opens one.
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

	// added sleep timer to allow time for
	// ApplicationCommandBulkOverwrite after creating bot session
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

// countSubCommands recursively counts SubCommand leaves, descending into
// SubCommandGroups so nested entries (e.g. effects under /image filter)
// are included.
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

// countCommandChoices counts choices on `type`-named string options.
// These act as command dispatchers (/get type:joke, /daily type:horoscope)
// and are distinct from parameter-only choices (/image meme pride flag:gay).
func countCommandChoices(opts []*discordgo.ApplicationCommandOption) int {
	count := 0
	for _, opt := range opts {
		if opt.Name == "type" && len(opt.Choices) > 0 {
			count += len(opt.Choices)
		}
	}
	return count
}
