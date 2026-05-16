#!/usr/bin/env bash
# Export LCA JSONL from Hugging Face (requires network + python3 with `datasets`).
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/.." && pwd)"
cd "$root"
exec python3 "$here/lca_hf_export.py" "$@"
