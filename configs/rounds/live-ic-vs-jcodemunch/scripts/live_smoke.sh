#!/usr/bin/env bash
set -euo pipefail
d="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
while [[ ! -x "$d/src/searchbench-go/buck_invoke.sh" && $d != "/" ]]; do d="$(dirname "$d")"; done
round="$d/configs/rounds/live-ic-vs-jcodemunch"
bundle="$round/artifacts/games/code-localization/rounds/live-ic-vs-jcodemunch-001"
exec "$d/src/searchbench-go/buck_invoke.sh" __buck round --mode=live_smoke \
	--repo-root="$d" --manifest="$round/round.pkl" --artifact-root="$round/artifacts" --bundle-path="$bundle"
