#!/usr/bin/env bash
# Writes <out>/coverage.json (Shields.io endpoint) and <out>/coverage.html.
# Usage: scripts/coverage_badge.sh <out-dir>
set -euo pipefail
out="${1:?usage: coverage_badge.sh <out-dir>}"
mkdir -p "$out"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

go test -coverpkg=./... -coverprofile="$tmp/raw.out" ./... >/dev/null
grep -v -e "/mocks/" -e "/examples/" "$tmp/raw.out" > "$tmp/cover.out"

pct="$(go tool cover -func="$tmp/cover.out" | awk '/^total:/ { sub("%", "", $3); print $3 }')"
color="$(awk -v p="$pct" 'BEGIN { print (p >= 80) ? "brightgreen" : (p >= 70) ? "yellow" : "red" }')"

printf '{"schemaVersion":1,"label":"coverage","message":"%s%%","color":"%s"}\n' "$pct" "$color" > "$out/coverage.json"
go tool cover -html="$tmp/cover.out" -o "$out/coverage.html"
echo "coverage: ${pct}% -> $out/coverage.json, $out/coverage.html"
