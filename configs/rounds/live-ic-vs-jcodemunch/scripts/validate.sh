#!/usr/bin/env bash
set -euo pipefail
d="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
while [[ ! -x "$d/src/searchbench-go/buck_invoke.sh" && $d != "/" ]]; do d="$(dirname "$d")"; done
exec "$d/src/searchbench-go/buck_invoke.sh" __buck round --mode=validate \
	--repo-root="$d" --manifest="$d/configs/rounds/live-ic-vs-jcodemunch/round.pkl"
