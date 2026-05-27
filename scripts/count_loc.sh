#!/usr/bin/env bash
# Count lines of code in this repo.
#
# Counts ALL lines (including comments and blank lines) — useful for an
# at-a-glance sense of project size, not for measuring meaningful "logic
# lines."
#
# Includes:  *.go
# Skips:
#   - vendor/                          (third-party Go modules)
#   - any directory starting with "."  (.git, .github, .claude, etc.)
#   - non-code files                   (yaml, md, json, csv, txt are
#                                       excluded by the file-extension
#                                       filter; nothing to maintain)
#
# To count additional code types (Python, JS, etc.), append more
# `-o -name '*.ext'` clauses to the find command below.

set -euo pipefail

# Run from the repo root regardless of where the script was invoked.
cd "$(dirname "$0")/.."

declare -A by_ext
total_lines=0
total_files=0

while IFS= read -r -d '' file; do
    lines=$(wc -l < "$file")
    ext="${file##*.}"
    by_ext[$ext]=$(( ${by_ext[$ext]:-0} + lines ))
    total_lines=$(( total_lines + lines ))
    total_files=$(( total_files + 1 ))
    printf '%8d  %s\n' "$lines" "$file"
done < <(find . \
    -type d \( -name vendor -o -name '.?*' \) -prune -o \
    -type f \( -name '*.go' \) -print0 | sort -z)
# Note: -name '.?*' matches dot-directories like .git, .github, .claude
# but NOT the starting "." itself (which would prune everything and
# return zero files).

echo
echo "==== Summary ===="
for ext in "${!by_ext[@]}"; do
    printf '%8d  *.%s\n' "${by_ext[$ext]}" "$ext"
done
echo "-----------------"
printf '%8d  files\n' "$total_files"
printf '%8d  total lines\n' "$total_lines"

# If stdout is a real terminal (i.e. not piped/redirected), pause so the
# window doesn't auto-close when launched by double-click on Windows or
# similar. Piped invocations (`... | tail`) skip this branch.
if [[ -t 1 ]]; then
    echo
    read -n 1 -s -r -p "Press any key to close..."
    echo
fi
