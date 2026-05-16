#!/usr/bin/env bash
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
mod="$root/src/searchbench-go"

run_go_test() {
	cd "$mod"
	exec go test -tags=live_e2e -count=1 -timeout=30m -run TestSearchBenchLiveICVsJCodeMunchE2E ./...
}

if command -v nix >/dev/null 2>&1 && [[ -f "$root/flake.nix" ]]; then
	exec nix develop "$root" -c bash -lc "$(declare -f run_go_test); run_go_test"
fi
run_go_test
