#!/usr/bin/env bash
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
cd "$root/src/searchbench-go"

if ! command -v pkl >/dev/null 2>&1; then
	echo "pkl_go_types_check: pkl not on PATH (use nix develop)" >&2
	exit 1
fi

pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl \
	--output-path=. \
	"$root/configs/schema/SearchBenchRound.pkl"

if ! git diff --quiet -- internal/adapters/config/pkl/generated/; then
	echo "pkl_go_types_check: generated Pkl Go bindings differ from HEAD." >&2
	echo "Run: buck2 build //src/searchbench-go:pkl_go_types" >&2
	echo "Then commit the updated files under internal/adapters/config/pkl/generated/." >&2
	exit 1
fi

if ! git diff --quiet --cached -- internal/adapters/config/pkl/generated/; then
	echo "pkl_go_types_check: staged generated bindings differ from working tree; commit consistently." >&2
	exit 1
fi
