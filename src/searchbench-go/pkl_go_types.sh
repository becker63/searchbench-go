#!/usr/bin/env bash
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
cd "$root/src/searchbench-go"
exec pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl \
	--output-path=. \
	"$root/configs/schema/SearchBenchRound.pkl"
