#!/usr/bin/env bash
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
cd "$root/src/searchbench-go"
exec go build -o /dev/null ./cmd/searchbench
