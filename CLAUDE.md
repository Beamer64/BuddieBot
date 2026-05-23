# BuddieBot — project context

Discord bot in Go: music (Lavalink), image manipulation, games, utilities. Hosted on a Linux Mint home server; CI/CD via GitHub Actions self-hosted runner. Dev on Windows.

## Sibling repos

`go.mod` `replace` directives point at:
- `../bb_images` — image processing library
- `../bb_data` — static data files (jokes, roasts, etc.)

Keep them as on-disk siblings. The deploy workflow checks all three out side-by-side.

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

Pre-defer helpers (`helper.SendResponseErrorToUser`, `helper.ReturnUserError`) only for validation that happens before the defer (e.g. `/get landsat` text length check). They assume the initial-response slot is still open.

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

All commands that generate an image attach the bytes directly to the Discord interaction response (or channel message for prefix commands) and reference them from an embed via `attachment://<filename>`. No third-party host. The pattern:

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
- `landsatLimiter` (in `getCmds.go`) — 30s per-user cooldown on `/get landsat`
- `landsatSem` (in `getCmds.go`) — 2-permit semaphore around the headless-Chrome work

To add a new limiter: `helper.NewRateLimiter(cooldown)` at package scope, call `.Allow(userID)` *before* the defer so the rate-limit message can use the immediate response slot.

## Lavalink (audio)

Config has `prod*` and `test*` fields for host/port/password; resolved at load via `isLaunchedByDebugger()`. Dev (Delve attached) spawns a child Lavalink via the `lavalink_runner` package; prod connects to a systemd-managed Lavalink. Don't call `lavalink_runner.Start()` in prod paths.

## Build / test / deploy

```bash
# BuddieBot/
go mod vendor
go build ./...
go vet ./...

# bb_images/
go test ./...

# LOC count
./scripts/count_loc.sh

# Deploy: push to master, self-hosted runner picks it up
git push origin master
```

Production config (`/opt/buddiebot/config.yaml` on the Mint server) is written from the `BOT_CONFIG_YAML` GitHub secret on every deploy. Don't hand-edit on the server — next deploy clobbers it.

## Gotchas

- **Landsat needs `chromedp.Sleep(5*time.Second)`** — three smarter wait conditions (`nth-of-type img`, `:last-of-type.active`, etc.) have all been tried and fire before tiles paint. Don't try a fourth without inspecting the live page mid-load.
- **dagpi is gone.** Don't reintroduce `client.X(...)` calls or the `Configs.Clients` field. Every image command uses bb_images now.
- **`static-ɢʟɨȶƈɦ`** is intentional — Unicode lookalike letters in the command name. Not a typo.
- **Avatar fetch** uses the `fetchImage(URL)` helper in `imgCmds.go`, not bare `http.Get`.
- **`replace` directives matter for CI** — the deploy workflow already checks out bb_images and bb_data as siblings. Don't "fix" them with v0.0.0 pseudo-versions.

## Style

- Comments explain **why**, not what. No per-function preambles that restate the signature.
- Prefer constants over magic numbers when the meaning isn't obvious from context.
- Keep file-level docblocks 1-3 lines when present at all.
