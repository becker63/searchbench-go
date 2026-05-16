#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
if ! command -v npm >/dev/null 2>&1; then
	echo "docs_build: npm not found on PATH (use nix develop)" >&2
	exit 1
fi
npm ci
npm run docs:build
