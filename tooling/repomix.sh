#!/usr/bin/env bash
set -euo pipefail
root="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z $root || ! -f "$root/src/searchbench-go/go.mod" ]]; then
	echo "tooling/repomix: run from inside the searchbench-go git checkout" >&2
	exit 1
fi
cd "$root"
if ! command -v repomix >/dev/null 2>&1; then
	echo "tooling/repomix: repomix not on PATH (use nix develop)" >&2
	exit 1
fi
repomix \
	--output repomix-output.xml \
	--style xml \
	--compress \
	--no-git-sort-by-changes
git add repomix-output.xml
