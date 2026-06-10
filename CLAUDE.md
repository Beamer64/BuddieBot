# BuddieBot — project context

Discord bot in Go: music (Lavalink), image manipulation, games, utilities. Hosted on a Linux Mint home server. CI builds on GitHub-hosted Actions and publishes a GitHub Release; the server pulls the release artifact on demand via the `pull-deploy` tool. No self-hosted runner. Dev on Windows.

## Sibling repos

`go.mod` `replace` directives point at:
- `../bb_images` — image processing library (60+ effects)
- `../bb_data` — static content datasets (jokes, roasts, kanye quotes, plus the embedded "buddie" images) embedded via `go:embed`

Keep them as on-disk siblings. The release workflow checks all three out side-by-side. **Any code or data you add to bb_data or bb_images needs to be committed AND pushed to that sibling repo** — the vendored copy in BuddieBot covers the local build, but CI re-clones the siblings from GitHub.

## Config

`config.Configs` embeds `*configuration` anonymously, so callers read `cfg.Keys.X` / `cfg.DiscordIDs.X` / etc. (no doubled `cfg.Configs.Configs.X`). The yaml file at `config_files/config.yaml` (prod) is loaded into the embedded struct; orphan yaml keys are silently ignored. Config also embeds `*database.DB` so handlers read it as `cfg.DB.X`.

Production config (`/opt/buddiebot/config.yaml`) is **server-managed** — edit in place via SSH and `sudo systemctl restart buddiebot`. Pull-deploy only swaps the binary; it never touches config.

## Database

SQLite via `modernc.org/sqlite` (pure Go, no CGO — matters for the cross-platform build). `sqlx` wraps it for struct-scanning, and `pressly/goose/v3` runs migrations from `pkg/database/migrations/*.sql` on `Open`. The DB file lives at `cfg.Database.ProdPath` (server) or `DevPath` (local).

Migrations applied in order:

| # | Table / change |
|---|---|
| 0001 | `Guild`, `User` — base tables, FK cascade User→Guild |
| 0002 | `ApiURL` — bot-wide external-URL catalog with `IsActive` flag (editable in DBeaver without redeploy) |
| 0003 | `Guild.LeftAt` — soft-delete marker on bot kick, cleared on rejoin |
| 0004 | `CommandUsage` — global per-command-path counter (e.g. `image filter blur`, `$goodboy`) |
| 0005 | `UserRating` — per-(user, rating-type) value, FK cascade User |
| 0006 | `UserCommandUsage` — per-(user, command-path) counter, FK cascade User |
| 0007 | `BannedUser` — global ban list, `Discord_UserID` UNIQUE, no FK (bans outlive `/user forget-me` and can be pre-emptive) |

Conventions: PascalCase columns, `ID INTEGER PRIMARY KEY` on every table, `Discord_` prefix for Discord snowflakes (e.g. `Discord_UserID`), FK column names match the referenced table (`GuildID` → `Guild.ID`), singular table names.

### Privacy stance

- **First-command materialization.** A user gets a `User` row on their first **slash OR prefix** command in a guild — both dispatchers funnel through `TrackCommandInvocation`, which calls `EnsureUser` before counting. Bots and DM-context invocations don't materialize (no guild to anchor the row to). Component clicks (buttons, selects) aren't tracked and don't materialize either.
- **No message content logged.** The error channel logs stack traces, not message bodies.
- **`/user forget-me` hard-deletes** the User row across every guild. FK `ON DELETE CASCADE` from `User.ID` removes the matching `UserRating` and `UserCommandUsage` rows automatically — no extra code path.
- **Command-usage recording happens *after* the handler runs.** The slash dispatcher (`events/eventHandlers.go: CommandHandler`) and the prefix dispatcher (`commands/prefix/prefixCmds.go: ParsePrefixCmds`) both call `cfg.DB.TrackCommandInvocation` post-handler. That helper bumps the aggregate `CommandUsage` row, ensures the (guild, user) row exists, then bumps the per-(user, command) counter — so `/user profile`'s Commands field is accurate from the very first invocation. The bare `IncrementUserCommandUsage` is still a SELECT-driven no-op when the row is missing, kept as defense-in-depth if upstream ever skips the ensure step.
- **Prefix-command keys are canonicalized to `$`.** Prefix invocations store as `$goodboy`, `$romans`, etc. regardless of the guild's actual prefix override — keeping the aggregate counter unified across guilds with different prefix styles. The `/user profile` renderer (`formatCommandStats`) swaps the `$` sentinel for the guild's current prefix at render time, so a user on a `plz `-prefix guild sees `plz goodboy`, not `$goodboy`. New prefix commands added to the `ParsePrefixCmds` switch are tracked automatically — the tracking block lives after the switch, with `default` early-returning so unknown commands don't get counted as typos.
- **Ban gate runs first.** Both dispatchers check `db.IsUserBanned(invokerID)` before anything else — slash/component clicks get an ephemeral "Your access to BuddieBot has been revoked." and bail; prefix commands silently ignore (no native ephemeral on text channels). Banned users never reach a handler, never get usage-tracked. Bans are GLOBAL (`BannedUser.Discord_UserID` is UNIQUE) but `/admin banned-list` filters per-guild via a JOIN on `User` so server admins only see entries relevant to their server. **Ban/unban preserve all user data** — profile, ratings, command counts — so an unban resumes where the user left off.

### Hot-path cache

`prefixCache` (a `map[string]string` + `RWMutex`) memoizes per-guild prefix overrides so `ParsePrefixCmds` (runs on **every** message) doesn't build a context or hit the DB after the first lookup. `SetGuildPrefixOverride` writes through. `CachedGuildPrefix(guildID)` is the pure cache read; `GetGuildPrefixOverride(ctx, guildID)` is cache-first with DB fallback and returns the default `$` on error (so a temporary DB hiccup doesn't break message parsing).

### Guild lifecycle & welcome message

`GuildCreate` fires for every guild on bot startup (backfill) AND on each new join. To welcome new servers without re-spamming existing ones on every restart, the handler uses `Guild.LeftAt` as the state signal:

| Pre-existing row state | Meaning | Welcome? |
|---|---|---|
| No row | First-ever sighting | ✓ |
| Row exists, `LeftAt IS NOT NULL` | Was kicked, just re-added | ✓ |
| Row exists, `LeftAt IS NULL` | Reconnect/backfill (already in guild) | ✗ |

`db.WelcomeNeeded(ctx, discordGuildID)` reads the prior state and returns a bool. **Call order matters**: `WelcomeNeeded` must run BEFORE `MarkGuildJoined`, since the latter clears `LeftAt` and would destroy the rejoin signal. The handler does both inside a single 5-second context.

The welcome embed lives in `pkg/events/welcome.go` — `welcomeEmbed(guildName, botUser)` builds it; `sendWelcomeMessage(s, e)` picks a channel via `pickWelcomeChannel` and posts. The channel picker prefers `Guild.SystemChannelID` (Discord's designated join/leave/boost channel), falls back to the first text channel the bot can write in, and returns `""` when nothing works (caller logs and skips rather than spamming random channels). Layout: 6 inline feature fields (Discord auto-arranges as 2 rows of 3 at typical widths) followed by 3 full-width fields (Getting Started, Privacy, Feedback).

## Per-server settings

`/admin set-prefix new-prefix:<x> [trailing-space:<bool>]` rewrites the guild's command prefix (default `$`). The boolean opts into "trailing-space mode" (`plz ` style) because Discord strips trailing whitespace from String options at the client layer — typing `new-prefix:"plz "` arrives server-side as `"plz"`. With `trailing-space:true` the server appends a space before validation. Validation in `adminCmds.go`:

- ≤ 5 chars total (counted AFTER the optional space is appended, so `abcde` + toggle is rejected as 6).
- No tabs or newlines.
- Spaces only at the end. `"plz "` is valid (so users type `plz roman 5`); `"my prefix"` is rejected.

The parser `splitPrefixCommand` in `prefixCmds.go` is **strict literal**: it consumes the prefix byte-for-byte and rejects any extra whitespace between the prefix and the command word. `"$roman 5"` matches `"$"`; `"$ roman 5"` does not. `"plz roman 5"` matches `"plz "`; `"plz  roman 5"` (double space) and `"plzroman 5"` (no space) do not. **`param` keeps its original whitespace** because `/palindrome` and `/romans` need it. `sendPrefixUpdateMsg` appends "(the trailing space is required)" to the announcement embed when the new prefix ends in a space, since Discord visually collapses trailing whitespace inside single backticks.

Empty prefix resets to the default `$`. Other per-server settings (audio enablement, event-notif channel, etc.) live on the `Guild` row and are edited via DBeaver / direct SQL — there's no `/admin <setting>` for each.

## Slash command conventions

**Every handler defers the interaction first**, before any work that could fail or take time:

```go
if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
}); err != nil {
    return fmt.Errorf("failed to defer interaction: %w", err)
}
```

Error/status routing follows one rule: **system failures are private + logged; status messages stay public.** The full helper map:

| Situation | Helper | Visibility | Channel-logged? |
|---|---|---|---|
| Pre-defer / component-click info or rate-limit notice | `helper.SendEphemeralMsgPreDeferred(s, i, msg)` | Ephemeral | No |
| Pre-defer system failure (DB lookup broke, etc.) | `helper.LogSendEphemeralMsgPreDeferred(s, i, msg, err)` | Ephemeral | Yes (via wrap) |
| Post-defer status message ("Nothing playing", validation hint) | `helper.EditMsgPostDeferred(s, i, msg)` or `audioEditMessage(...)` | Public (inherits defer) | No |
| Post-defer system failure | `helper.LogSendEphemeralFollowUpPostDeferred(s, i, msg, err)` | Ephemeral followup (deletes the orphan "Bot is thinking…" placeholder first) | Yes |
| Post-defer ephemeral status (rare) | `helper.SendEphemeralFollowUpPostDeferred(s, i, msg)` | Ephemeral followup | No |

`Log…` variants return the err so `slash.wrap()` channel-logs it; bare variants return the send-error (or nil) for the caller's normal control flow. See `pkg/helper/discord.go` for the section-divided source — three groups: Interaction error helpers, Prefix-command error helpers, Channel logging.

**Final success response** → `s.InteractionResponseEdit(...)` with `discordgo.WebhookEdit`.

### Validator-registry pattern (used by `/generate`)

When a command has per-subcommand validation and you want validation to be pluggable as new subcommands are added, define a map from subcommand name to a `func(s, i, opts) bool` validator. The validator returns `false` (after calling `helper.SendEphemeralMsgPreDeferred` for user-input failures, or `LogSendEphemeralMsgPreDeferred` for system failures) to abort cleanly; `true` to proceed. The dispatcher just does a map lookup, no inline switch over subcommand names. See `generateCmds.go` (`generateValidators`).

## Component handlers

UI button / select-menu clicks land in `slash.ComponentHandlers` (in `handlers.go`) — keys are matched **by prefix** by the event handler. Custom IDs encode the parameters the handler needs:

| Component | Custom-ID format | Handler |
|---|---|---|
| `/tuuck cmd-list` pagination | `tuuck-page:<scope>:<page>` | `sendTuuckPageResponse` |
| `/user profile` pagination | `profile-page:<targetID>:<invokerID>:<page>` | `sendProfilePageResponse` — ownership-checked |
| `/user forget-me` confirm | `forget-me-confirm:<invokerID>` | `forgetMeConfirm` |
| `/user forget-me` cancel | `forget-me-cancel:<invokerID>` | `forgetMeCancel` |
| `/daily horoscope` sign select | `horo-select` | `sendHoroscopeCompResponse` |
| `/game wyr` reroll | `wyr-reroll` | `sendWYRrerollResp` |
| `/game wyr` votes | `wyr-votes` | `sendWYRvotesResp` |

Common pattern for ownership: encode the invoker's Discord ID in the custom ID, verify `i.Member.User.ID == invokerID` at the top of the handler, return `helper.SendEphemeralMsgPreDeferred` if not.

### Clickable command mentions

`helper.CommandMention(name, sub...)` renders `</name sub:ID>` when the bot has registered command IDs (captured at startup by `bot.registerCommands` → `helper.SetCommandIDs`). Falls back to plain `/name sub` text when the registry isn't populated (tests, fresh process). Discord only resolves these as clickable pills for **leaf** commands (no subcommands) or **full subcommand paths** — never for parent commands. `/tuuck`'s renderer (`commandIsLeaf`) picks plain-text vs mention accordingly.

## Command landscape

| Group | Lives in | Notes |
|---|---|---|
| `/admin set-prefix|ban|unban|banned-list` | `adminCmds.go` | Server-admin only (DefaultMemberPermissions=ManageGuild). `set-prefix` sets the per-server prefix. `ban`/`unban` write to `BannedUser` (global scope; user data preserved); audit embed posts to `cfg.DiscordIDs.BuddieBotHQBanChannelID`. `banned-list` shows users banned globally who have history in *this* guild (DB JOIN on `User`). |
| `/animals doggo|katz` | `animalCmds.go` | dog/cat APIs |
| `/audio play|stop|resume-queue|queue|skip|clear` | `audioCmds.go` | Lavalink playback. Gated by `helper.IsAudioGuild`. |
| `/daily advice|kanye|affirmation|fact|tongue-twister|horoscope` | `dailyCmds.go` | One-shot daily content |
| `/feedback` | `utilityCmds.go` | User-submitted feature suggestions / bug reports / other. Posts directly to BuddieBotHQ channels (no DB storage). 1 per (user, category) per 30 min — so the same user can fire off one of each category back-to-back. Details field ≥30 chars. Routing in `feedbackChannelFor`; destination channel IDs in `cfg.DiscordIDs.BuddieBotHQSuggestionChannelID` / `BuddieBotHQBugChannelID`. |
| `/game just-lost|wyr` | `gameCmds.go` | Mini-games. wyr has reroll/vote buttons. |
| `/generate cistercian|landsat|fake-person` | `generateCmds.go` | Image/embed generators. Validators in `generateValidators`. |
| `/get rekd|joke|8ball|yomomma|pickup-line|xkcd` | `getCmds.go` | Text-based responses (no image generation — those moved to `/generate`). |
| `/image …` | `imgCmds.go` | 60+ effects (see Image command organization). |
| `/pick steam|choices|poll` | `pickCmds.go` | Picks one of N |
| `/rate-this …` | `rateThisCmds.go` | Random scoring; persists the value against the target's `UserRating` row (unless target is a bot — those don't get materialized). |
| `/tuuck cmd-list` | `utilityCmds.go` | Help — paginated, sources from the live `slash.Commands` registry. Top-level page mixes prefix commands (with the guild's current prefix) and slash commands; per-command page lists subcommands/choices. |
| `/txt …` | `txtCmds.go` | Text-transformation effects |
| `/user profile|forget-me` | `userCmds.go` | Profile shows Recent Ratings corner, Dosh, Day-One badge, BB User Since, plus a Commands field with total invocations + most-used. forget-me is button-confirmed and hard-deletes across every guild. |

**Prefix commands** (`PrefCmdsList` in `prefixCmds.go`): `$release`, `$test`, `$weast`, `$palindrome`, `$romans`, `$goodboy`. `$release`, `$test`, and `$weast` are in `NoShowCmds` and hidden from `/tuuck`. `$release` and `$test` are additionally **bot-owner gated** (test-guild-only) via `helper.IsBotOwner(authorID, cfg.DiscordIDs.BotOwnerIDs)` — `$release` broadcasts release notes to every guild; `$test` is the in-progress-feature sandbox. `$goodboy` posts a random image from `bb_data/buddie`.

## Image command organization (`/image`)

6 SubCommandGroups, alphabetical within each, comment-numbered `// <group> - NN`:

| Group | Contents |
|---|---|
| `filter` | Color/tone transforms |
| `distort` | Spatial/structural transforms |
| `animated` | Procedural GIFs |
| `overlay` | Templates placed *over* the avatar |
| `sign` | Text-input templates (tweet/youtube/etc.) |
| `meme` | Visual-template memes |

Discord caps each group at 25 entries.

## bb_images structure

| Package | Output | Use |
|---|---|---|
| `color` | PNG | Per-pixel color transforms |
| `spatial` | PNG | Geometric transforms |
| `edges` | PNG | Edge detection (sobel, sketch) |
| `overlays` | PNG or GIF | Template on top of src |
| `signs` | PNG | Avatar in marker template, often with text |
| `animated` | GIF | Procedural frame generation |
| `special` | PNG | Stylized algorithmic effects (lego, ascii) |
| `internal/draw` | helpers | Decode/Encode/RenderFrames/LazyFrames/AnimateOverGIF |
| `internal/templates` | helpers | Detect, ConnectedRegions |

## bb_data structure

Each subpackage embeds its dataset via `go:embed`, exposes `Load(fs.FS)` for one-time startup wiring, and a `Random()` accessor (or similar) for runtime use. `bb_data.Load()` is called once from `bot.Init` and chains every subpackage's `Load`. Originally text-only; now bundles a small set of images too (the buddie subpackage).

| Package | Source | Accessor |
|---|---|---|
| `affirmations` | `affirmations.jsonl` | `Random() string` |
| `buddie` | `datasets/buddie/*.jpg|.jpeg` | `Random() Image`, `Count() int` — `Image{Filename, Data}` for direct Discord attachment |
| `eightball` | `8ball.txt` | `Random() string` |
| `emojis` | `emojis.txt` | `Random() string` |
| `facts` | `facts.txt` | `Random() string` |
| `jokes` | `shortjokes.json` | `Random() string` |
| `kanye` | `kanyequotes.json` | `Random() string` |
| `loadingmessages` | `loading_messages.txt` | `Random() string` |
| `pickuplines` | `pickuplines.json` | `Random() string` |
| `roasts` | `roasts.txt` | `Random() string` |
| `textfonts` | `text_fonts.json` | `Convert(text, group) string`, `Groups() []string` |
| `tonguetwister` | `tongue_twisters.txt` | `Random() string` |
| `wyr` | `WYR.csv` | `Random() Poll`, `Count() int` |
| `yomomma` | `yomomma.json` | `Random() string` |

The `internal/pick` package has the shared `Random[T]`, `LoadLines`, `LoadJSON`, `LoadJSONL` helpers. Accessors are zero-value-safe before `Load`: `Random()` returns `""` (or zero `Image`/`Poll`).

## Key helpers

### GIF overlays — `LazyFrames` + `AnimateOverGIF`

```go
var fooTemplate = draw.NewLazyFrames(fooBytes)  // decoded once, cached

func Foo(src image.Image) ([]byte, error) {
    return draw.AnimateOverGIF(fooTemplate, func(frame *image.RGBA) *image.Paletted {
        resized := imaging.Fill(frame, w, h, imaging.Center, imaging.Linear)
        composited := imaging.Overlay(src, resized, image.Point{}, fooOpacity)
        p := image.NewPaletted(composited.Bounds(), palette.Plan9)
        imgdraw.FloydSteinberg.Draw(p, composited.Bounds(), composited, image.Point{})
        return p
    })
}
```

Gives you free: per-process decode caching, per-frame parallelism, Plan9 + Floyd-Steinberg quantization. Use `imaging.Linear` not `Lanczos` — palette quantization eats the quality difference and Linear is ~2× faster.

### Marker-based signs/memes

Marker color conventions (the marker PNGs were painted to match — don't change):

| Color | Role |
|---|---|
| Green `{G: 255, A: 255}` | Primary avatar / "first" user slot |
| Blue `{B: 250, A: 255}` | Secondary avatar (note: B=250 not 255) |
| Magenta `{R:255, G:0, B:255}` | Text region (body) |
| Cyan `{R:0, G:255, B:255}` | Secondary text (display name) |
| Yellow `{R:255, G:255, B:26}` | Tertiary text (@handle) |

For "one avatar drawn into multiple slots" templates (5guys1girl, thanks-obama): use `templates.ConnectedRegions` and loop over the returned rectangles. `templates.Detect` returns a single union bounding box and will stretch one avatar across all blobs — almost always wrong.

### Image responses

All commands that generate an image attach the bytes directly to the Discord interaction response and reference them from an embed via `attachment://<filename>`. No third-party host. Pattern:

```go
embed := &discordgo.MessageEmbed{
    Image: &discordgo.MessageEmbedImage{URL: "attachment://" + fileName},
}
webhookEdit := &discordgo.WebhookEdit{
    Embeds: &[]*discordgo.MessageEmbed{embed},
    Files: []*discordgo.File{{
        Name:        fileName,
        ContentType: "image/png", // or image/gif
        Reader:      bytes.NewReader(imgBytes),
    }},
}
```

For prefix commands, the equivalent is `discordgo.MessageSend{ Embed: ..., Files: [...] }` via `s.ChannelMessageSendComplex` (see `sendGoodBoy`). `contentTypeFor(filename)` picks the right MIME based on the extension.

### Rate limiting / semaphores

- `imgCmdLimiter` (in `imgCmds.go`) — 5s per-user cooldown on `/image *`
- `landsatLimiter` (in `generateCmds.go`) — 30s per-user cooldown on `/generate type:landsat`
- `landsatSem` (in `generateCmds.go`) — 2-permit semaphore around the headless-Chrome work
- `feedbackLimiter` (in `utilityCmds.go`) — 30 min per (user, category) on `/feedback`; the composite key lets the same user fire off one of each category back-to-back

To add a new limiter: `helper.NewRateLimiter(cooldown)` at package scope, call `.Allow(userID)` *before* the defer so the rate-limit message can use the immediate response slot.

## Audio (`/audio` + Lavalink)

User-facing commands all live under `/audio`:

| Subcommand | What it does |
|---|---|
| `play url-1[,url-2,url-3]` | Resolves URL(s) via Lavalink; YouTube playlist URLs auto-queue every track. Starts playback if idle; queues otherwise. |
| `stop` | Disconnects from voice but **saves** the active track + queue. Use `resume-queue` to resume. |
| `resume-queue` | Rejoins voice, restarts the saved track from position 0, continues the queue. |
| `queue` | Shows currently playing OR the stopped/saved track, plus upcoming queue. |
| `skip` | Advances to next; if queue is empty, leaves voice. |
| `clear` | Wipes upcoming queue **and** the saved/stopped track. Full session reset. |

`/audio` is gated by `helper.IsAudioGuild(...)` — only the master and test guilds. In other servers the command shows but returns a user-facing "not enabled here" message.

State lives in `voice_chat.Player`:
- `queues[gID]` — upcoming track list
- `announceChannels[gID]` — channel for auto-advance "Now playing" announcements
- `pausedTracks[gID]` — saved active track when Stop was called (replayed by ResumeQueue from position 0)

`PlayResult.Playlist` is non-nil when Play resolved a playlist URL; `voice_chat.FormatPlayResult(r, resumeCmd)` is the shared user-facing formatter. The split between user-facing audio errors (use `audioEditMessage` → public) and system audio errors (`LogSendEphemeralFollowUpPostDeferred` → ephemeral + logged) is set by `voice_chat.IsUserFacingError(err)`.

Config has `prod*` and `test*` Lavalink fields for host/port/password, resolved at load via `helper.IsLaunchedByDebugger()`. Dev (Delve attached) spawns a child Lavalink via the `lavalink_runner` package; prod connects to a systemd-managed Lavalink. Don't call `lavalink_runner.Start()` in prod paths.

## Tests

Test files mirror their production file names: `fooCmds.go` → `fooCmds_test.go`. INTEGRATION-tagged tests skip by default (they need real Discord/API credentials) and only run with `INTEGRATION=true` in the environment — use them sparingly, they're mostly debug scripts.

Invariant-style tests worth knowing about:

- **`TestPrefCmdsListMatchesSwitch`** ([prefixCmds_test.go](pkg/commands/prefix/prefixCmds_test.go)) catches drift between the `PrefCmdsList` map and the dispatch switch in `ParsePrefixCmds`. Adding a prefix command requires updating both; this test makes that visible.
- **`TestNoShowCmdsAreKnown`** (same file) — every `NoShowCmds` entry must be a real key in `PrefCmdsList`. Catches typos that silently hide nothing.
- **`TestSplitPrefixCommand`** (same file) — the parser's contract across prefix styles (tight `$`, trailing-space `plz `, edge cases). Add a case if you change parsing.
- **`TestCommandHandlers_AllNonNil`**, **`TestComponentHandlers_AllNonNil`**, **`TestCommands_AllHaveDescriptions`** ([handlers_test.go](pkg/commands/slash/handlers_test.go)) — catches nil-handler entries and missing spec descriptions (which Discord rejects at registration).
- **`TestEnsureUserSeedsRatingsOnCreate`** + **`TestUserRatingFKCascadeOnUserDelete`** ([pkg/database/user_rating_test.go](pkg/database/user_rating_test.go)) — locks in "fresh user gets N seeded ratings" + "forget-me deletes ratings via FK cascade."
- **`TestIncrementUserCommandUsageSafeWhenUserMissing`** ([pkg/database/user_command_usage_test.go](pkg/database/user_command_usage_test.go)) — defense-in-depth: the bare-increment SQL is no-op-safe even if upstream skips `EnsureUser`. (The pre-first-command-materialization version of this asserted the opposite privacy contract; the rename reflects the policy shift.)
- **`TestDetectJournalPermissionIssue`** ([scripts/pull-deploy/main_test.go](scripts/pull-deploy/main_test.go)) — catches the journal-permission misconfig early instead of letting it time out as a misleading health-check failure.
- **`TestFlexStringUnmarshal`** ([generateCmds_test.go](pkg/commands/slash/generateCmds_test.go)) — randomuser.me's postcode is sometimes int, sometimes string; this guards the custom unmarshaler that handles both.
- **`TestMemberSinceDisplay`** ([userCmds_test.go](pkg/commands/slash/userCmds_test.go)) — three accepted timestamp formats; same-instant inputs must produce identical rendering.
- **`TestWelcomeNeeded`** ([pkg/database/database_test.go](pkg/database/database_test.go)) — locks in the 3-state guild-lifecycle contract (no row / `LeftAt` set / `LeftAt` null) that drives whether the welcome message fires.
- **`TestUnbanPreservesUserData`** ([pkg/database/banned_test.go](pkg/database/banned_test.go)) — locks in "ban/unban cycle preserves the user's profile, ratings, and command counts so they resume where they left off."
- **`TestListBannedUsersInGuild`** (same file) — locks in the per-guild filter on `/admin banned-list` — bans from other guilds and pre-emptive bans without User rows are excluded.
- **`TestIsBotOwner`** ([pkg/helper/discord_test.go](pkg/helper/discord_test.go)) — covers the owner-ID gate for `$release` / `$test`: matching ID true, non-match false, empty ID/slice false (safe-by-default for unpopulated config).

The voice_chat package has pure-function tests for `FormatPlayResult`, `FriendlyPlayError`, `IsUserFacingError`, `briefExceptionReason`. The stateful methods (`Play`, `Stop`, `ResumeQueue`, `Skip`, `Queue`) are untested — they'd need interface extraction for `disgolink.Client` and `*discordgo.Session` to be mockable.

## Build / test / deploy

```bash
# BuddieBot/
go mod vendor
go build ./...
go vet ./...
go test ./...

# bb_images/, bb_data/ — same commands in their dirs

# LOC count
./scripts/count_loc.sh

# Deploy: pushing to master triggers .github/workflows/release.yml on a
# GitHub-hosted runner, which publishes a GitHub Release with the binary +
# SHA-256. The server pulls it on demand:
#   sudo systemctl start buddiebot-deploy.service
# See scripts/pull-deploy/README.md for the deploy tool, install, rollback,
# and the systemd-journal group requirement (without it pull-deploy's health
# check times out misleadingly — TestDetectJournalPermissionIssue surfaces
# this fast).
git push origin master
```

Server prerequisites for pull-deploy (one-time):

1. `buddiebot` user in the `systemd-journal` group (so the deploy script can read the bot's journal for the ready-marker health check). Without it, deploys fail fast with a clear error pointing here.
2. `/etc/sudoers.d/buddiebot-runner` contains `buddiebot ALL=(root) NOPASSWD: /bin/systemctl restart buddiebot`.
3. `/opt/buddiebot/config.yaml` exists and is owned by `buddiebot:buddiebot` mode 0600.
4. Lavalink running as a separate systemd unit.

## Gotchas

- **Dev-vs-prod selector is `IsLaunchedByDebugger()` in `pkg/helper/debug.go`.** Returns true on either (a) Delve attached as the parent process, or (b) `BUDDIEBOT_FORCE_DEV` env var non-empty. False → prod config (prod token, prod Lavalink, prod DB path). The IDE's plain "Run" mode does NOT attach Delve, so without the env var it lands on prod and would (1) take over the production gateway connection from the server bot, (2) re-welcome every prod guild because the local DB is empty, (3) generally cause chaos. **`bot.Init` refuses to start with prod config on Windows** — the only sanctioned local-dev paths are IDE "Debug" mode (Delve attached) or setting `BUDDIEBOT_FORCE_DEV=1` in the IDE's Run configuration. The Windows guard is intentionally narrow: prod servers are Linux, so this never triggers in real production.
- **Helper-name pairs.** `Log…` variants surface the err to `wrap()` so the error channel gets a stack; bare variants don't. Use `Log…` for system failures, bare for status messages / rate-limit notices / "not yours" component-check messages. Misusing them silently drops failures from the error channel.
- **Status messages stay public, system failures go ephemeral.** Reach for `EditMsgPostDeferred` / `audioEditMessage` for "Nothing playing" / "Prefix must be ≤5 chars" type feedback; `LogSendEphemeralFollowUpPostDeferred` for "the API broke." The deferred-placeholder cleanup happens inside the followup helper — don't reproduce it inline.
- **Trailing-space prefixes are first-class.** `splitPrefixCommand` in `prefixCmds.go` handles both `"$"` and `"plz "`. Adding a new prefix command means updating `PrefCmdsList` AND a case in the switch — `TestPrefCmdsListMatchesSwitch` catches drift.
- **Adding a prefix command's row to `PrefCmdsList`** automatically surfaces it in `/tuuck cmd-list` with the guild's current prefix character stamped on. Add to `NoShowCmds` to hide it (e.g. admin-gated commands).
- **Adding a new bb_data subpackage** requires wiring its `Load(fs.FS)` into `bb_data.Load()` so the umbrella loader picks it up. Forgetting this is silent — `Random()` just returns zero values.
- **Landsat needs `chromedp.Sleep(5*time.Second)`** — three smarter wait conditions (`nth-of-type img`, `:last-of-type.active`, etc.) have all been tried and fire before tiles paint. Don't try a fourth without inspecting the live page mid-load.
- **dagpi is gone.** Don't reintroduce `client.X(...)` calls or the `Configs.Clients` field. Every image command uses bb_images now.
- **`static-ɢʟɨȶƈɦ`** is intentional — Unicode lookalike letters in the command name. Not a typo.
- **Avatar fetch** uses the `fetchImage(URL)` helper in `imgCmds.go`, not bare `http.Get`.
- **`replace` directives matter for CI** — the release workflow already checks out bb_images and bb_data as siblings. Don't "fix" them with v0.0.0 pseudo-versions.
- **bb_data / bb_images changes need pushing** — re-vendoring in BuddieBot only helps the local build. CI clones the sibling repos fresh from GitHub; uncommitted changes there cause "no required module provides package …" failures.
- **fake-person postcode is `flexString`**, not `int` — randomuser.me serves it as either a JSON number (US/CA) or a string (UK/etc., for alphanumeric codes). The custom unmarshaler handles both; don't "simplify" it back to int.
- **Bot targets in `/rate-this` and `/user profile`** are filtered out — bots don't get materialized in the User table. The check happens at the call site (`if !target.Bot`), not inside `EnsureUser`, since the helper has no awareness of who's a bot.

## Style

- Comments explain **why**, not what. No per-function preambles that restate the signature.
- Prefer constants over magic numbers when the meaning isn't obvious from context.
- Keep file-level docblocks 1-3 lines when present at all.
