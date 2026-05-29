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
#    (This is the same shared config scripts/deploy.sh already relies on.)
sudo install -m 600 -o buddiebot -g buddiebot /path/to/your/config.yaml /opt/buddiebot/config.yaml

# 4. (Optional but recommended) seed current-version so the first run doesn't
#    re-pull the version that's already deployed. Use whatever tag is currently live.
echo "release-XXXXXXX" | sudo -u buddiebot tee /opt/buddiebot/current-version
```

### Sudoers requirement

pull-deploy invokes `sudo systemctl restart buddiebot` for the one privileged step. The `buddiebot` user already has a narrowly-scoped sudoers entry for this (the existing `scripts/deploy.sh` uses the same call). If you're setting up a clean box, `/etc/sudoers.d/buddiebot-runner` should contain:

```
buddiebot ALL=(root) NOPASSWD: /bin/systemctl restart buddiebot
```

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

## Coexistence with the runner

`.github/workflows/release.yml` publishes a GitHub Release on every push to master (built on a GitHub-hosted runner, so it doesn't touch the home server). `.github/workflows/deploy.yml` continues to deploy via the self-hosted runner as before — both run on the same push without interfering. The pull-deploy flow can be exercised in parallel without breaking anything; once it's proven across a few real deploys, `deploy.yml` and the self-hosted runner can be retired.

## Update the deploy tool itself

If you change `main.go`, rebuild on the server:

```sh
cd ~/buddiebot-src && git pull
cd scripts/pull-deploy && go build -o /opt/buddiebot/bin/pull-deploy .
```

(The deploy tool isn't part of the release artifact — it doesn't auto-update itself. That's deliberate: a buggy pull-deploy that auto-updated itself could brick the deploy path.)