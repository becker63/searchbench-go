#!/usr/bin/env bash
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"$here/repomix.sh"

root="$(git rev-parse --show-toplevel)"
cd "$root"

if [[ ! -f repomix-output.xml ]]; then
	echo "repomix_fresh_check: repomix-output.xml missing after regeneration" >&2
	exit 1
fi

if grep -qE '^<file path="repomix-output\.xml"' repomix-output.xml; then
	echo "repomix_fresh_check: repomix-output.xml appears to include itself (check repomix ignore patterns)" >&2
	exit 1
fi
if grep -qE '^<file path="configs/rounds/[^"]+/artifacts/' repomix-output.xml; then
	echo "repomix_fresh_check: repomix-output.xml includes configs/rounds/*/artifacts (should be excluded)" >&2
	exit 1
fi
if grep -q '^<file path="src/searchbench-go/internal/adapters/config/pkl/generated/' repomix-output.xml; then
	echo "repomix_fresh_check: repomix-output.xml includes generated Pkl paths (should be excluded)" >&2
	exit 1
fi

if ! git diff --quiet -- repomix-output.xml || ! git diff --quiet --cached -- repomix-output.xml; then
	echo "repomix_fresh_check: repomix-output.xml is not committed at HEAD after regeneration." >&2
	echo "" >&2
	echo "The Repomix snapshot must be committed before pushing." >&2
	echo "" >&2
	echo "Fix with one of:" >&2
	echo "  git add repomix-output.xml && git commit --amend --no-edit" >&2
	echo "  git add repomix-output.xml && git commit -m 'Update Repomix snapshot'" >&2
	echo "" >&2
	echo "Then push again." >&2
	exit 1
fi
