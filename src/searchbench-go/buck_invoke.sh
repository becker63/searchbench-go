#!/usr/bin/env bash
# Builds searchbench into $TMPDIR and execs the full __buck command line from Buck.
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
bin="$(mktemp "${TMPDIR:-/tmp}/searchbench-buck-XXXXXX")"
trap 'rm -f "$bin"' EXIT
(cd "$here" && go build -trimpath -o "$bin" ./cmd/searchbench)
exec "$bin" "$@"
