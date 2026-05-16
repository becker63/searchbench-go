#!/usr/bin/env bash
# Run live IC vs jCodeMunch e2e once (cost-conscious defaults). Not for CI.
set -euo pipefail
set -o pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

if [[ -f "$root/src/searchbench/.env" ]]; then
	set -a
	# shellcheck disable=SC1091
	source "$root/src/searchbench/.env"
	set +a
fi

export SEARCHBENCH_RUN_LIVE_E2E=1
export SEARCHBENCH_MATERIALIZE_CACHE_DIR="${SEARCHBENCH_MATERIALIZE_CACHE_DIR:-$root/.cache/searchbench/materialized-repos}"
export SEARCHBENCH_JCODEMUNCH_COMMAND="${SEARCHBENCH_JCODEMUNCH_COMMAND:-uvx jcodemunch-mcp}"
export SEARCHBENCH_ITERATIVE_CONTEXT_COMMAND="${SEARCHBENCH_ITERATIVE_CONTEXT_COMMAND:-uv run --directory $root/src/iterative-context python -m iterative_context.server}"
export SEARCHBENCH_LIVE_USE_HF_ROW="${SEARCHBENCH_LIVE_USE_HF_ROW:-1}"
export SEARCHBENCH_LCA_HF_MAX_ITEMS="${SEARCHBENCH_LCA_HF_MAX_ITEMS:-1}"
export SEARCHBENCH_LCA_HF_SKIP="${SEARCHBENCH_LCA_HF_SKIP:-50}"
export SEARCHBENCH_LIVE_E2E_TIMEOUT="${SEARCHBENCH_LIVE_E2E_TIMEOUT:-45m}"

log="$root/.cache/searchbench/live-e2e-$(date -u +%Y%m%dT%H%M%SZ).log"
mkdir -p "$(dirname "$log")"
echo "Logging to $log"

run_go_test() {
	cd "$root/src/searchbench-go"
	env SEARCHBENCH_RUN_LIVE_E2E=1 \
		SEARCHBENCH_SKIP_HF_EXPORT="${SEARCHBENCH_SKIP_HF_EXPORT:-1}" \
		SEARCHBENCH_LIVE_USE_HF_ROW="${SEARCHBENCH_LIVE_USE_HF_ROW:-1}" \
		go test -tags=live_e2e -count=1 -timeout=45m -v \
		-run 'TestExportLCADatasetFromHuggingFace|TestSearchBenchLiveICVsJCodeMunchE2E' ./... 2>&1 | tee "$log"
}

if command -v nix >/dev/null 2>&1 && [[ -f "$root/flake.nix" ]]; then
	exec nix develop "$root" -c bash -lc "$(declare -f run_go_test); run_go_test"
fi
run_go_test
