#!/usr/bin/env bash
set -euo pipefail
d="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
while [[ ! -x "$d/src/searchbench-go/buck_invoke.sh" && $d != "/" ]]; do d="$(dirname "$d")"; done
exec "$d/src/searchbench-go/buck_invoke.sh" __buck dataset materialize-lca \
	--manifest-dir="$d/configs/rounds/live-ic-vs-jcodemunch" --config=py --split=dev --max-items=1 --skip=50
