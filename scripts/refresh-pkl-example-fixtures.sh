#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

local_dir="$root/configs/experiments/local-ic-vs-jcodemunch"
optimize_dir="$root/configs/experiments/optimize-ic"

mkdir -p "$optimize_dir/policies" "$optimize_dir/scoring" "$optimize_dir/artifacts/runs"

cp "$local_dir/policies/candidate_policy.py" "$optimize_dir/policies/candidate_policy.py"
cp "$local_dir/scoring/localization-objective.pkl" "$optimize_dir/scoring/localization-objective.pkl"

rm -rf "$local_dir/artifacts/runs/example-round-001"

GOCACHE="$root/.tmp/go-cache" go run ./cmd/searchbench run \
  --manifest "$local_dir/experiment.pkl" \
  --bundle-root "$local_dir/artifacts/runs" \
  --bundle-id example-round-001

touch "$optimize_dir/artifacts/runs/.gitkeep"
