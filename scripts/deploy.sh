#!/usr/bin/env bash
# Build BuddieBot and deploy into /opt/buddiebot.
#
# Layout produced:
#   /opt/buddiebot/
#   ├── current -> builds/<utc-ts>-<short-sha>/   (symlink, atomically swung)
#   ├── builds/
#   │   ├── <newest>/   <- this deploy
#   │   ├── <middle>/   <- previous deploy
#   │   └── <oldest>/   <- two deploys ago; pruned next time
#   └── config.yaml      <- shared, server-managed; symlinked into each build
#
# Runs on the self-hosted GitHub runner as the buddiebot user. The only
# privileged action is `sudo systemctl ...`, narrowly scoped via
# /etc/sudoers.d/buddiebot-runner.

set -euo pipefail

BUILD_ROOT="${BUILD_ROOT:-/opt/buddiebot}"
KEEP="${KEEP:-3}"
GO_BIN="${GO_BIN:-/usr/local/go/bin/go}"

short_sha="${GITHUB_SHA:0:7}"
[[ -n "$short_sha" ]] || short_sha="manual"
timestamp="$(date -u +%Y%m%d%H%M%S)"
build_id="${timestamp}-${short_sha}"
build_dir="${BUILD_ROOT}/builds/${build_id}"

echo "==> Build ID: ${build_id}"

# --- sanity check the shared secrets file exists before doing any work
if [[ ! -f "${BUILD_ROOT}/config.yaml" ]]; then
    echo "ERROR: ${BUILD_ROOT}/config.yaml is missing — set it up once before the first deploy" >&2
    exit 1
fi

# --- build
echo "==> Compiling"
mkdir -p "${build_dir}/config_files"
"${GO_BIN}" version
"${GO_BIN}" build -trimpath -ldflags="-s -w" \
    -o "${build_dir}/buddiebot" \
    ./cmd/discord-bot

# --- stage runtime files
echo "==> Staging runtime files"
cp config_files/cmd.yaml "${build_dir}/config_files/cmd.yaml"
ln -s "${BUILD_ROOT}/config.yaml" "${build_dir}/config_files/config.yaml"
cp -r datasets "${build_dir}/datasets"

# --- atomically swing the current symlink
echo "==> Swinging current symlink"
ln -sfn "${build_dir}" "${BUILD_ROOT}/current.new"
mv -Tf "${BUILD_ROOT}/current.new" "${BUILD_ROOT}/current"

# --- restart and verify
echo "==> Restarting buddiebot service"
sudo systemctl restart buddiebot

# Give systemd a moment to actually start the process before checking is-active
sleep 3
if [[ "$(sudo systemctl is-active buddiebot)" != "active" ]]; then
    echo "ERROR: buddiebot did not come up. Recent status:" >&2
    sudo systemctl status buddiebot >&2 || true
    exit 1
fi
echo "==> buddiebot is active"

# --- prune oldest builds beyond the KEEP newest
echo "==> Pruning old builds (keeping ${KEEP} newest)"
cd "${BUILD_ROOT}/builds"
# Build dir names sort chronologically because they're prefixed with a UTC
# timestamp. sort -r gives newest first; we keep the first KEEP entries.
mapfile -t all < <(ls -1 | sort -r)
if (( ${#all[@]} > KEEP )); then
    for old in "${all[@]:KEEP}"; do
        echo "  removing ${old}"
        rm -rf -- "${old}"
    done
fi

echo "==> Deploy ${build_id} complete"
