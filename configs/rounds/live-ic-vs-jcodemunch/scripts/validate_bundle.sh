#!/usr/bin/env bash
set -euo pipefail
d="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
while [[ ! -x "$d/src/searchbench-go/buck_invoke.sh" && $d != "/" ]]; do d="$(dirname "$d")"; done
bundle="$d/configs/rounds/live-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/live-ic-vs-jcodemunch-001"
exec "$d/src/searchbench-go/buck_invoke.sh" __buck validate-bundle --bundle-path="$bundle"
