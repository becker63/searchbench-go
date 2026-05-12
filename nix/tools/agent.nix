# Parallel agent workflow commands.
{
  pkgs,
  mkInRepo,
  searchbench-e2e,
  searchbench-check-generated,
  searchbench-update-repomix,
  searchbench-staticcheck,
  searchbench-golangci,
  searchbench-nix-flake-check,
  searchbench-go-mod-tidy-check,
  searchbench-go-test-race,
}:
let
  searchbench-agent-start = mkInRepo {
    name = "searchbench-agent-start";
    text = ''
      if [[ $# -lt 2 ]]; then
        echo "usage: searchbench-agent-start <issue-slug> \"goal description\"" >&2
        exit 1
      fi
      slug="$1"
      shift
      goal="$*"
      here="$(pwd)"
      parent="$(dirname "$here")"
      worktrees="$parent/searchbench-go.worktrees"
      wt="$worktrees/$slug"
      mkdir -p "$worktrees"
      if [[ -e "$wt" ]]; then
        echo "searchbench-agent-start: worktree path already exists: $wt" >&2
        exit 1
      fi
      git worktree add -b "agent/$slug" "$wt"
      {
        echo "## Goal"
        echo ""
        echo "$goal"
        echo ""
        echo "## Non-goals"
        echo ""
        echo "- Real MCP, LangSmith, or production datasets unless the issue says otherwise."
        echo ""
        echo "## Allowed areas"
        echo ""
        echo "- Follow scope in the GitHub issue; prefer the smallest change that satisfies acceptance criteria."
        echo ""
        echo "## Required checks"
        echo ""
        echo "- Run \`nix develop -c searchbench-agent-check\` before requesting review."
        echo ""
        echo "## Merge checklist"
        echo ""
        echo "- Human reviews the diff and merges; this command does not merge."
        echo ""
        echo "## Architecture reminders"
        echo ""
        echo "- One round path; evaluator and optimizer agents; keep \`internal/pure\` free of adapters, agents, and surface imports."
      } > "$wt/AGENT_TASK.md"
      echo "Created worktree at $wt on branch agent/$slug"
      echo "Next:"
      echo "  cd \"$wt\""
      echo "  nix develop -c searchbench-agent-check"
    '';
    runtimeInputs = [ ];
  };

  searchbench-agent-check = mkInRepo {
    name = "searchbench-agent-check";
    text = ''
      if ! command -v pre-commit >/dev/null; then
        echo "searchbench-agent-check: pre-commit not found (use nix develop)" >&2
        exit 1
      fi
      pre-commit run --all-files
      go test ./...
      ${searchbench-e2e}/bin/searchbench-e2e
      ${searchbench-check-generated}/bin/searchbench-check-generated
      ${searchbench-update-repomix}/bin/searchbench-update-repomix
      git status --short
    '';
    runtimeInputs = [
      pkgs.pre-commit
      pkgs.go
      pkgs.pkl
      pkgs.templ
      pkgs.repomix
      searchbench-e2e
      searchbench-check-generated
      searchbench-update-repomix
    ];
  };

  searchbench-agent-pack = mkInRepo {
    name = "searchbench-agent-pack";
    text = ''
      branch="$(git branch --show-current)"
      echo "branch=$branch"
      echo "--- diffstat ---"
      git diff --stat
      echo "--- recent commits ---"
      git log -5 --oneline
      echo "--- repomix ---"
      if [[ -f repomix-output.xml ]]; then
        echo "repomix-output.xml (present)"
      else
        echo "repomix-output.xml missing — run searchbench-update-repomix"
      fi
      echo "--- git status ---"
      git status --short
      review="$PWD/AGENT_REVIEW.md"
      {
        echo "# Agent review pack"
        echo "branch: $branch"
        echo ""
        git diff --stat
        echo ""
        git log -5 --oneline
      } >"$review"
      echo "Wrote $review"
    '';
    runtimeInputs = [ ];
  };

  searchbench-agent-merge-check = mkInRepo {
    name = "searchbench-agent-merge-check";
    text = ''
      if ! command -v pre-commit >/dev/null; then
        echo "searchbench-agent-merge-check: pre-commit not found (use nix develop)" >&2
        exit 1
      fi
      pre-commit run --all-files
      go test ./...
      ${searchbench-go-test-race}/bin/searchbench-go-test-race
      ${searchbench-staticcheck}/bin/searchbench-staticcheck
      ${searchbench-golangci}/bin/searchbench-golangci
      ${searchbench-e2e}/bin/searchbench-e2e
      ${searchbench-nix-flake-check}/bin/searchbench-nix-flake-check
      ${searchbench-check-generated}/bin/searchbench-check-generated
      ${searchbench-go-mod-tidy-check}/bin/searchbench-go-mod-tidy-check
      ${searchbench-update-repomix}/bin/searchbench-update-repomix
      git diff --check
    '';
    runtimeInputs = [
      pkgs.pre-commit
      pkgs.go
      pkgs.pkl
      pkgs.templ
      pkgs.nix
      pkgs.repomix
      pkgs.golangci-lint
      pkgs.go-tools
      searchbench-e2e
      searchbench-check-generated
      searchbench-update-repomix
      searchbench-nix-flake-check
      searchbench-go-mod-tidy-check
      searchbench-staticcheck
      searchbench-golangci
      searchbench-go-test-race
    ];
  };
in
{
  inherit
    searchbench-agent-start
    searchbench-agent-check
    searchbench-agent-pack
    searchbench-agent-merge-check
    ;
}
