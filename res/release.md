# Release Notes 05/27/26

<h2> Here is a more detailed look at the most recent release notes.</h2>
These notes cover bot behavior changes, new commands, retired commands, and what's coming next.

<h3> Comments: </h3>
This one's a big one. A lot of the work was under the hood — pulling image and data work into dedicated in-house libraries so BuddieBot stops needing other people's services to be online — and laying a real datastore foundation so per-server settings, privacy controls, and (soon) an economy can actually exist.

&nbsp;

If something looks broken, please yell at me. As usual, retry a couple of times before sounding the alarm — the occasional hiccup is almost always a passing API or a Discord cache being weird.

&nbsp;

Thanks big bunches luv u.

\- Harley

* ...poo poo pee pee stinky

\- Also Harley

<h3> Highlights </h3>

* **Built to last** — image & data commands now run on BuddieBot's own libraries instead of third-party services.
* **Help that actually helps** — `/tuuck cmd-list` lets you browse every command and drill into one for its options + a real usage example.
* **Datastore foundation** — per-server settings + the groundwork for a coming-soon economy.
* **Privacy controls** — `/user profile` shows what's stored; `/user forget-me` wipes it all across every server.
* **Audio under one roof** — `/audio play|queue|skip|clear|stop|resume-queue`, with YouTube playlist support.
* **Custom prefix per server** — `/admin set-prefix` to override the default `$`.
* **Rate limiting** — heavier commands now have per-user cooldowns so one person can't hog the bot.

&nbsp;

<h3> Stability & Architecture </h3>

The image effects and the static data (jokes, roasts, kanye quotes, etc.) used to lean on a third-party service (dagpi) and a pile of inline `go:embed` files. Both have been extracted into their own sibling libraries:

* **bb_images** — every image effect now runs in-process. 60+ effects split across six packages: `color` (filters), `spatial` (distortions), `edges` (sketch/sobel), `overlays` (templates over your avatar), `signs` (avatar-in-template w/ text), `animated` (procedural GIFs), and `special` (algorithmic stuff like lego/ascii). GIF effects use a shared "decode once, render frames in parallel" pipeline.
* **bb_data** — all the static content (jokes, 8-ball, pickup lines, kanye quotes, affirmations, facts, etc.) embedded once and accessed through clean `Random()` accessors.

The net effect: no more dagpi outages taking image commands offline, faster effect rendering, and a cleaner separation between "what the bot does" and "what assets it uses."

&nbsp;

<h3> Help System — /tuuck </h3>

`/tuuck` got rewritten into a real help command instead of the static text dump it used to be.

* `/tuuck cmd-list` — paginated, embed-style browse of every command with its top-line description.
* `/tuuck cmd-list command:<name>` — drill into one command (e.g. `command:image`) to see its subcommands or choices, paginated 25 per page, with a curated usage example pinned to the embed description so it's visible on every page.
* Pagination uses Prev/Next buttons that work for everyone the embed reaches.
* The list is **sourced from the live registered commands** — there's no separate spec to keep in sync, so it can never go stale. Only the per-command example strings are curated by hand.

If you previously used `/tuuck cmd-help`, that's been folded into `cmd-list` — same info, fewer commands to remember.

&nbsp;

<h3> Data, Privacy & Coming Soon: Economy </h3>

BuddieBot now has a real datastore (SQLite under the hood) so it can remember things between restarts. Two new user-facing commands:

* `/user profile` — shows what BuddieBot has on you (joined date, basic per-server info, etc.).
* `/user forget-me` — a button-confirmed, **global** wipe across every server you share with the bot. There's a confirmation step so you don't nuke yourself by accident.

Privacy stance, plain English:

* Tracking is **opt-in**. The bot only stores info about you once you actually use a data-touching command.
* **Message content is never logged.**
* `/user forget-me` is a hard delete, not a soft "marked deleted" flag — your row is gone.

This is also the foundation for the next big chunk of work: **an economy system** ("Dosh") where you'll be able to earn, spend, gamble, and generally hoard imaginary currency. Coming in a future release; not in this one.

&nbsp;

<h3> Audio </h3>

All music commands now live under `/audio`. The old `$play`/etc. prefix commands are gone.

* `/audio play url-1[, url-2, url-3]` — resolves YouTube URLs. If it's a YouTube **playlist** URL, the whole playlist gets queued.
* `/audio queue` — shows the current track + everything upcoming (or the saved/stopped track if there's no active playback).
* `/audio skip` — next track; if the queue is empty, the bot leaves voice.
* `/audio stop` — leaves voice but **saves** the current track + queue for later.
* `/audio resume-queue` — rejoins voice and picks up where Stop left off.
* `/audio clear` — wipes the upcoming queue and any saved track; full reset.

> *Note: `/audio` is enabled in select servers only (currently the master + test guilds). In other servers the command still shows up but tells you it's not enabled here. Talk to me if you want it turned on somewhere new.*

&nbsp;

<h3> Server Customization </h3>

* `/admin set-prefix new-prefix:!` — change the prefix this server uses for the remaining `$`-style commands. Pass an empty value to reset to the default (`$`). Validated to ≤ 5 characters with no whitespace. Cached per-guild so it doesn't hit the database on every message.
* Other per-server settings (audio enablement, etc.) are edited directly in the database for now rather than through bot commands — keeps the surface area small.

&nbsp;

<h3> Reliability </h3>

* **Rate limiting** — `/image *` has a 5-second per-user cooldown; `/generate type:landsat` has a 30-second per-user cooldown plus a 2-permit semaphore around the headless-Chrome work (it's expensive). These exist so one person rapid-firing a heavy command doesn't bog down the bot for everyone else.
* **Graceful degradation** — if the per-guild prefix lookup fails, the bot uses the default prefix and logs locally instead of spamming the error channel on every message during a database hiccup.
* **Panic recovery** — every slash and component handler runs through a `wrap` that recovers from panics, logs the stack trace, and sends the user a "something went wrong" message instead of leaving them staring at a never-resolving interaction.

&nbsp;

<h3> Moved / Changed Commands </h3>

**Slash:**

* `$cistercian` → `/generate cistercian`
* `/get fake-person` → `/generate fake-person`
* `/generate landsat` — new; generates a real-time Landsat satellite image (heavy, rate-limited).
* `/tuuck` rewritten (see Help System above); `/tuuck cmd-help` retired (folded into `cmd-list`).
* `/user profile`, `/user forget-me` — new.
* `/admin set-prefix` — new.

**Prefix (`$`) commands** — slimmed down to:

* `$release` — admin-only, posts these release notes.
* `$weast` — easter egg.
* `$palindrome <text>` — palindrome check.
* `$romans <number-or-roman>` — convert between Arabic and Roman numerals.

All audio prefix commands have been retired in favor of `/audio`.

&nbsp;

<h3> Bug Fixes </h3>

* Fixed `/pick` panicking when picking from subcommand-style choices (was calling `.StringValue()` on a non-string option). The steam rate-limiter wired up at the same time now actually triggers.
* Per-guild prefix lookups now use an in-memory cache with a context-bound DB fallback — no more "hit the database on every message" hot-path concern.
