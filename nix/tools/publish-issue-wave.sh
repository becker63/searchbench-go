# shellcheck shell=bash
# Fragment included after mk-in-repo preamble (see core.nix).

manifest=""
dry_run=0
want_develop=0

usage() {
	echo "usage: searchbench-publish-issue-wave [--dry-run] [--develop] <manifest.json>" >&2
	echo "" >&2
	echo "Publish GitHub issues from a JSON manifest. This is dev coordination tooling only;" >&2
	echo "it does not touch SearchBench runtime code. Requires gh + jq and repo auth." >&2
	echo "" >&2
	echo "  --dry-run    Print planned actions; no gh writes." >&2
	echo "  --develop    After each created issue with parallel_safe=true, run gh issue develop." >&2
	exit 1
}

normalize_title() {
	local s="$1"
	s="${s#"${s%%[![:space:]]*}"}"
	s="${s%"${s##*[![:space:]]}"}"
	printf '%s' "$s"
}

while [[ $# -gt 0 ]]; do
	case "$1" in
	--dry-run) dry_run=1 ;;
	--develop) want_develop=1 ;;
	-h | --help) usage ;;
	*)
		if [[ -n $manifest ]]; then
			usage
		fi
		manifest="$1"
		;;
	esac
	shift
done

[[ -n $manifest ]] || usage
[[ -f $manifest ]] || {
	echo "searchbench-publish-issue-wave: manifest not found: $manifest" >&2
	exit 1
}

command -v jq >/dev/null || {
	echo "searchbench-publish-issue-wave: jq not on PATH" >&2
	exit 1
}
command -v gh >/dev/null || {
	echo "searchbench-publish-issue-wave: gh not on PATH" >&2
	exit 1
}

repo_override=""
repo_override=$(jq -r '.repo // empty' "$manifest")
gh_repo=()
if [[ -n $repo_override ]]; then
	gh_repo=(--repo "$repo_override")
fi

count=$(jq '.issues | length' "$manifest")
if [[ ! $count =~ ^[0-9]+$ ]]; then
	echo "searchbench-publish-issue-wave: invalid manifest (.issues length)" >&2
	exit 1
fi

if [[ $count -eq 0 ]]; then
	echo "searchbench-publish-issue-wave: no issues in manifest" >&2
	exit 1
fi

tmp_body="$(mktemp)"
cleanup() {
	rm -f "$tmp_body"
}
trap cleanup EXIT

for ((i = 0; i < count; i++)); do
	title=$(jq -r --argjson idx "$i" '.issues[$idx].title' "$manifest")
	if [[ -z $title || $title == "null" ]]; then
		echo "searchbench-publish-issue-wave: missing title at issues[$i]" >&2
		exit 1
	fi

	body_inline=$(jq -r --argjson idx "$i" '.issues[$idx].body // empty' "$manifest")
	body_file=$(jq -r --argjson idx "$i" '.issues[$idx].body_file // empty' "$manifest")
	if [[ -n $body_inline && $body_inline != "null" && -n $body_file && $body_file != "null" ]]; then
		echo "searchbench-publish-issue-wave: issues[$i] sets both body and body_file" >&2
		exit 1
	fi

	: >"$tmp_body"
	if [[ -n $body_file && $body_file != "null" ]]; then
		if [[ ! -f $body_file ]]; then
			echo "searchbench-publish-issue-wave: body_file not found for issues[$i]: $body_file (paths are relative to repo root)" >&2
			exit 1
		fi
		cp "$body_file" "$tmp_body"
	elif [[ -n $body_inline && $body_inline != "null" ]]; then
		printf '%s\n' "$body_inline" >"$tmp_body"
	fi

	label_args=()
	nlabels=$(jq --argjson idx "$i" '(.issues[$idx].labels // []) | length' "$manifest")
	if [[ $nlabels =~ ^[0-9]+$ ]] && [[ $nlabels -gt 0 ]]; then
		for ((j = 0; j < nlabels; j++)); do
			lab=$(jq -r --argjson idx "$i" --argjson j "$j" '.issues[$idx].labels[$j]' "$manifest")
			if [[ -n $lab && $lab != "null" ]]; then
				label_args+=(--label "$lab")
			fi
		done
	fi

	nt="$(normalize_title "$title")"
	dup_found=0
	dup_url=""
	while IFS=$'\t' read -r ot ourl; do
		[[ "$(normalize_title "$ot")" == "$nt" ]] || continue
		dup_found=1
		dup_url="$ourl"
		break
	done < <(gh "${gh_repo[@]}" issue list --state open --limit 500 --json title,url --jq '.[] | "\(.title)\t\(.url)"')

	if [[ $dup_found -eq 1 ]]; then
		echo "skip (duplicate title): $title -> ${dup_url:-?}"
		continue
	fi

	if [[ $dry_run -eq 1 ]]; then
		echo "dry-run: would create issue: $title"
		if [[ ${#label_args[@]} -gt 0 ]]; then
			echo "         labels: ${label_args[*]}"
		fi
		parallel_safe=$(jq -r --argjson idx "$i" '.issues[$idx].parallel_safe // false' "$manifest")
		if [[ $want_develop -eq 1 && $parallel_safe == "true" ]]; then
			bn=$(jq -r --argjson idx "$i" '.issues[$idx].develop_branch_name // empty' "$manifest")
			echo "         would run: gh issue develop <new> --name ${bn:-<auto>}"
		fi
		continue
	fi

	out=$(gh "${gh_repo[@]}" issue create --title "$title" --body-file "$tmp_body" "${label_args[@]}")
	num=""
	if [[ $out =~ issues/[0-9]+$ ]]; then
		num="${out##*/}"
	fi
	if [[ -z $num ]]; then
		echo "searchbench-publish-issue-wave: could not parse issue number from: $out" >&2
		exit 1
	fi
	echo "created #$num: $title"

	parallel_safe=$(jq -r --argjson idx "$i" '.issues[$idx].parallel_safe // false' "$manifest")
	if [[ $want_develop -eq 1 && $parallel_safe == "true" ]]; then
		bn=$(jq -r --argjson idx "$i" '.issues[$idx].develop_branch_name // empty' "$manifest")
		if [[ -z $bn || $bn == "null" ]]; then
			bn="issue-$num-develop"
		fi
		if ! gh "${gh_repo[@]}" issue develop "$num" --name "$bn"; then
			echo "searchbench-publish-issue-wave: gh issue develop failed for #$num (set branch manually: $bn)" >&2
		fi
	fi
done
