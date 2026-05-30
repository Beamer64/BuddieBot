# pull-deploy

Manual, server-initiated deploy for BuddieBot. Replaces the self-hosted GitHub runner flow with a tiny Go program that pulls a pre-built binary from a GitHub Release on demand.

## Why

The self-hosted runner makes the home server an inbound target — any workflow on the public repo could potentially execute code on it. Pull-deploy inverts the direction: the Mint server initiates an HTTPS connection to `api.github.com`, downloads a release artifact, verifies its checksum, and applies it locally. **No inbound ports, no runner registration, no code from PRs ever runs on the box.**

## How it works

1. Reads `/opt/buddiebot/current-version`.
2. `GET https://api.github.com/repos/Beamer64/BuddieBot/releases/latest`.
3. If the tag matches the current version and `--force` wasn't passed → exit 0 quietly.
4. Otherwise: downloads `buddiebot` + `buddiebot.sha256`, verifies the checksum, stages into `/opt/buddiebot/builds/<timestamp>-<sha>/`.
5. Swings the `/opt/buddiebot/current` symlink atomically.
6. `sudo systemctl restart buddiebot`.
7. Tails `journalctl -u buddiebot` for up to 30s looking for the `Logged in as` ready marker.
8. On success: writes the new tag to `current-version`, prunes builds beyond `--keep`.
9. On health-check failure: leaves the new build staged but logs instructions for rolling back manually.

The directory layout matches `scripts/deploy.sh` exactly — `current` symlink, `builds/<id>/` with a `config_files/config.yaml` symlink to the server-managed `/opt/buddiebot/config.yaml` — so the existing systemd unit for `buddiebot.service` works without change.

## Install (one-time, on the Mint server)

```sh
# 1. Build the binary
git clone https://github.com/Beamer64/BuddieBot.git ~/buddiebot-src
cd ~/buddiebot-src/scripts/pull-deploy
go build -o /opt/buddiebot/bin/pull-deploy .

# 2. Install the systemd unit
sudo cp ~/buddiebot-src/scripts/systemd/buddiebot-deploy.service /etc/systemd/system/
sudo systemctl daemon-reload

# 3. Make sure /opt/buddiebot/config.yaml exists and is readable by the buddiebot user.
#    (This is the same shared config the bot already relies on.)
sudo install -m 600 -o buddiebot -g buddiebot /path/to/your/config.yaml /opt/buddiebot/config.yaml

# 4. Grant the buddiebot user read access to the system journal (see "Journal
#    access requirement" below). Without this, pull-deploy's health-check
#    can't see the bot's "Logged in as" log line.
sudo usermod -aG systemd-journal buddiebot

# 5. (Optional but recommended) seed current-version so the first run doesn't
#    re-pull the version that's already deployed. Use whatever tag is currently live.
echo "release-XXXXXXX" | sudo -u buddiebot tee /opt/buddiebot/current-version
```

### Sudoers requirement

pull-deploy invokes `sudo systemctl restart buddiebot` for the one privileged step. The `buddiebot` user already has a narrowly-scoped sudoers entry for this. If you're setting up a clean box, `/etc/sudoers.d/buddiebot-runner` should contain:

```
buddiebot ALL=(root) NOPASSWD: /bin/systemctl restart buddiebot
```

### Journal access requirement

pull-deploy reads `journalctl -u buddiebot` to detect the bot's `"Logged in as"` line after restart — that's the health check. By default a service user like `buddiebot` can't read the system journal; permission requires `systemd-journal` group membership:

```sh
sudo usermod -aG systemd-journal buddiebot
```

If you skip this step, pull-deploy fails fast on the first poll with a clear error message pointing back here, rather than silently timing out. Group membership only applies to *new* processes, so the next `sudo systemctl start buddiebot-deploy.service` picks it up automatically.

## Run

```sh
# Recommended — logs land in the journal:
sudo systemctl start buddiebot-deploy.service
sudo journalctl -u buddiebot-deploy.service -n 100 --no-pager

# Or directly, as the buddiebot user:
sudo -u buddiebot /opt/buddiebot/bin/pull-deploy

# Redeploy the current version (e.g., after you've manually edited config.yaml):
sudo -u buddiebot /opt/buddiebot/bin/pull-deploy --force

# Rehearse without touching the live symlink or the service:
sudo -u buddiebot /opt/buddiebot/bin/pull-deploy --dry-run
```

## Sample output

A successful deploy of a new release:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 pull-deploy · 2026-05-29 22:18:46 UTC · pid 121604
 build-root=/opt/buddiebot  force=false  dry-run=false
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
==> latest release: release-5bfe4cc (2 assets)
==> current version: release-0a126f4 → upgrading to release-5bfe4cc
==> staging release-5bfe4cc → /opt/buddiebot/builds/20260529231846-5bfe4cc
    downloaded binary (12.3 MiB)
    checksum verified: b969fe7a84b940cd…
==> swapped 'current' symlink → builds/20260529231846-5bfe4cc
==> restarting buddiebot.service
==> waiting for ready marker "Logged in as" (deadline 30s)
    observed at t+1.2s
==> updated current-version → release-5bfe4cc
==> pruned 1 old build(s) (kept 3 newest)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 pull-deploy · OK · 4.1s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

A no-op (already at the latest release):

```
━━━…opening banner…━━━
==> latest release: release-5bfe4cc (2 assets)
==> current version: release-5bfe4cc
    nothing to do (use --force to redeploy)
━━━ pull-deploy · OK · 0.4s ━━━
```

The heavy bars on either end make consecutive runs in the journal easy to tell apart at a glance.

## Flags

| Flag | Default | Purpose |
|---|---|---|
| `--build-root` | `/opt/buddiebot` | Base directory |
| `--binary-asset` | `buddiebot` | Name of the binary asset in the release |
| `--sha256-asset` | `buddiebot.sha256` | Name of the checksum asset |
| `--keep` | `3` | How many old builds to retain after a successful deploy |
| `--force` | `false` | Redeploy even if the current version is up to date |
| `--dry-run` | `false` | Skip the symlink swap and service restart |
| `--health-deadline` | `30s` | Max time to wait for the ready marker after restart |
| `--ready-marker` | `Logged in as` | Substring expected in the journal |
| `--service` | `buddiebot` | Systemd unit to restart |

## Manual rollback

Each deploy keeps the last `--keep` builds under `/opt/buddiebot/builds/`. To roll back:

```sh
ls /opt/buddiebot/builds/                                  # pick the previous build dir
ln -sfn /opt/buddiebot/builds/<previous>/ /opt/buddiebot/current.new
mv -Tf /opt/buddiebot/current.new /opt/buddiebot/current
sudo systemctl restart buddiebot
echo "<previous-tag>" | sudo -u buddiebot tee /opt/buddiebot/current-version
```

The current-version file lives outside the swap target, so updating it last keeps state consistent if you have to abort mid-rollback.

## How releases get published

`.github/workflows/release.yml` runs on every push to `master` on a **GitHub-hosted** runner (no self-hosted runner anywhere in the loop). It builds the bot for `linux/amd64`, computes the SHA-256, and publishes both as a GitHub Release tagged `release-<short-sha>`. Pull-deploy on the server reads from that release.

## Update the deploy tool itself

If you change `main.go`, rebuild on the server:

```sh
cd ~/buddiebot-src && git pull
cd scripts/pull-deploy && go build -o /opt/buddiebot/bin/pull-deploy .
```

(The deploy tool isn't part of the release artifact — it doesn't auto-update itself. That's deliberate: a buggy pull-deploy that auto-updated itself could brick the deploy path.)