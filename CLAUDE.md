# BuddieBot — project context

Discord bot in Go: music (Lavalink), image manipulation, games, utilities. Hosted on a Linux Mint home server. CI builds on GitHub-hosted Actions and publishes a GitHub Release; the server pulls the release artifact on demand via the `pull-deploy` tool. No self-hosted runner. Dev on Windows.

## Sibling repos

`go.mod` `replace` directives point at:
- `../bb_images` — image processing library
- `../bb_data` — static data files (jokes, roasts, etc.) embedded via `go:embed`

Keep them as on-disk siblings. The release workflow checks all three out side-by-side. **Any code or data you add to bb_data or bb_images needs to be committed AND pushed to that sibling repo** — the vendored copy in BuddieBot covers the build, but CI re-clones the siblings from GitHub.

## Config

`config.Configs` embeds `*configuration` anonymously, so callers read `cfg.Keys.X` / `cfg.DiscordIDs.X` / etc. (no doubled `cfg.Configs.Configs.X`). The yaml file at `config_files/config.yaml` (prod) is loaded into the embedded struct; orphan yaml keys are silently ignored.

## Slash command conventions

**Every handler defers the interaction first**, before any work that could fail or take time:

```go
if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
}); err != nil {
    return fmt.Errorf("failed to defer interaction: %w", err)
}
```

Then:
- **Post-defer errors** → `helper.ReturnUserErrorDeferred(s, i, userMsg, err)`
- **Final response** → `s.InteractionResponseEdit(...)` with `discordgo.WebhookEdit`

Pre-defer helpers (`helper.SendEphemeralError`, `helper.ReturnUserError`) only for validation that happens before the defer (e.g. `/generate type:landsat` text length check). They assume the initial-response slot is still open.

### Validator-registry pattern (used by `/generate`)

When a command has per-subcommand validation and you want validation to be pluggable as new subcommands are added, define a map from subcommand name to a `func(s, i, opts) bool` validator. The validator returns `false` (after calling `helper.ReturnUserError`) to abort cleanly; `true` to proceed. The dispatcher just does a map lookup, no inline switch over subcommand names. See `generateCmds.go` (`generateValidators`).

## Command landscape

| Group | Lives in | Notes |
|---|---|---|
| `/animals doggo|katz` | `animalCmds.go` | dog/cat APIs |
| `/audio play|stop|resume-queue|queue|skip|clear` | `audioCmds.go` | Lavalink playback (see below) |
| `/daily advice|kanye|affirmation|fact|tongue-twister|horoscope` | `dailyCmds.go` | One-shot daily content |
| `/game just-lost|wyr` | `gameCmds.go` | Mini-games. wyr has component buttons. |
| `/generate cistercian|landsat|fake-person` | `generateCmds.go` | Produces images/embeds |
| `/get rekd|joke|8ball|yomomma|pickup-line|xkcd` | `getCmds.go` | Text-based responses |
| `/image …` | `imgCmds.go` | 60+ effects (see Image command organization) |
| `/pick steam|choices|poll` | `pickCmds.go` | Picks one of N |
| `/rate-this …` | `rateThisCmds.go` | Random scoring of a target |
| `/tuuck` | `utilityCmds.go` | Help — sources directly from `slash.Commands` |
| `/txt …` | `txtCmds.go` | Text-transformation effects |

Prefix commands (`$`-prefixed) are limited to `$release`, `$weast`, `$palindrome`, `$romans` now — the audio commands migrated to `/audio`.

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

Each subpackage embeds its dataset via `go:embed`, exposes `Load(fs.FS)` for one-time startup wiring, and a `Random()` accessor (or similar) for runtime use. `bb_data.Load()` is called once from `bot.Init` and chains every subpackage's `Load`.

| Package | Source file | Accessor |
|---|---|---|
| `affirmations` | `affirmations.jsonl` | `Random() string` |
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

The `datasets/` directory contains additional JSON files (`captcha.json`, `countries.json`, `pokemons.json`, etc.) that don't have corresponding subpackages — currently unused. The `internal/pick` package has the shared `Random[T]`, `LoadLines`, `LoadJSON`, `LoadJSONL` helpers.

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

### Rate limiting / semaphores

- `imgCmdLimiter` (in `imgCmds.go`) — 5s per-user cooldown on `/image *`
- `landsatLimiter` (in `generateCmds.go`) — 30s per-user cooldown on `/generate type:landsat`
- `landsatSem` (in `generateCmds.go`) — 2-permit semaphore around the headless-Chrome work

To add a new limiter: `helper.NewRateLimiter(cooldown)` at package scope, call `.Allow(userID)` *before* the defer so the rate-limit message can use the immediate response slot.

## Audio (`/audio` + Lavalink)

User-facing commands all live under `/audio`:

| Subcommand | What it does |
|---|---|
| `play url-1[,url-2,url-3]` | Resolves URL(s) via Lavalink; YouTube playlist URLs auto-queue every track in the playlist. Starts playback if idle; queues otherwise. |
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

`PlayResult.Playlist` is non-nil when Play resolved a playlist URL; `voice_chat.FormatPlayResult(r, resumeCmd)` is the shared user-facing formatter (single track or playlist, with stopped-state hints).

Config has `prod*` and `test*` Lavalink fields for host/port/password, resolved at load via `helper.IsLaunchedByDebugger()`. Dev (Delve attached) spawns a child Lavalink via the `lavalink_runner` package; prod connects to a systemd-managed Lavalink. Don't call `lavalink_runner.Start()` in prod paths.

## Tests

Test files mirror their production file names: `fooCmds.go` → `fooCmds_test.go`. INTEGRATION-tagged tests skip by default (they need real Discord/API credentials) and only run with `INTEGRATION=true` in the environment — use them sparingly, they're mostly debug scripts.

A few invariant-style tests are worth knowing about:
- `TestPrefixNamesMatchSwitch` ([prefixCmds_test.go](pkg/commands/prefix/prefixCmds_test.go)) catches drift between the `Names` slice and the dispatch switch in `prefixCmds.go`. If you add/remove a prefix command and only update one of those two places, this test will tell you.
- `TestCommandHandlers_AllNonNil`, `TestComponentHandlers_AllNonNil`, `TestCommands_AllHaveDescriptions` ([handlers_test.go](pkg/commands/slash/handlers_test.go)) catch nil-handler entries and missing spec descriptions (which Discord rejects at registration).

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
# See scripts/pull-deploy/README.md for the deploy tool, install, and rollback.
git push origin master
```

Production config (`/opt/buddiebot/config.yaml` on the Mint server) is **server-managed** — edit in place via SSH and `sudo systemctl restart buddiebot`. Pull-deploy only swaps the binary; it never touches config. (Earlier setups wrote config from a `BOT_CONFIG_YAML` GitHub secret on every deploy — that path is gone with the self-hosted runner.)

## Gotchas

- **Landsat needs `chromedp.Sleep(5*time.Second)`** — three smarter wait conditions (`nth-of-type img`, `:last-of-type.active`, etc.) have all been tried and fire before tiles paint. Don't try a fourth without inspecting the live page mid-load.
- **dagpi is gone.** Don't reintroduce `client.X(...)` calls or the `Configs.Clients` field. Every image command uses bb_images now.
- **`static-ɢʟɨȶƈɦ`** is intentional — Unicode lookalike letters in the command name. Not a typo.
- **Avatar fetch** uses the `fetchImage(URL)` helper in `imgCmds.go`, not bare `http.Get`.
- **`replace` directives matter for CI** — the release workflow already checks out bb_images and bb_data as siblings. Don't "fix" them with v0.0.0 pseudo-versions.
- **bb_data / bb_images changes need pushing** — re-vendoring in BuddieBot only helps the local build. CI clones the sibling repos fresh from GitHub; uncommitted changes there cause "no required module provides package …" failures.

## Style

- Comments explain **why**, not what. No per-function preambles that restate the signature.
- Prefer constants over magic numbers when the meaning isn't obvious from context.
- Keep file-level docblocks 1-3 lines when present at all.
