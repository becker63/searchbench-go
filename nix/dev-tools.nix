# Nix-generated project commands (issue #36 — no loose shell files under scripts/).
{ pkgs }:

let
  mkInRepo =
    {
      name,
      text,
      runtimeInputs ? [ ],
    }:
    pkgs.writeShellApplication {
      inherit name;
      runtimeInputs = [ pkgs.git ] ++ runtimeInputs;
      text = ''
        set -euo pipefail
        _root="$(git rev-parse --show-toplevel 2>/dev/null || true)"
        if [[ -z "$_root" || ! -f "$_root/go.mod" ]]; then
          echo "${name}: not at the root of the searchbench-go git repository" >&2
          exit 1
        fi
        cd "$_root"
      ''
      + text;
    };

  searchbench-go-test-all = mkInRepo {
    name = "searchbench-go-test-all";
    text = ''
      exec go test ./...
    '';
    runtimeInputs = [ pkgs.go ];
  };

  searchbench-e2e = mkInRepo {
    name = "searchbench-e2e";
    text = ''
      exec go test -count=1 .
    '';
    runtimeInputs = [
      pkgs.go
      pkgs.pkl
    ];
  };

  searchbench-architecture-check = mkInRepo {
    name = "searchbench-architecture-check";
    text = ''
      exec go test ./internal/architecture/...
    '';
    runtimeInputs = [ pkgs.go ];
  };

  searchbench-check-pkl-generated = mkInRepo {
    name = "searchbench-check-pkl-generated";
    text = ''
      if ! command -v pkl >/dev/null; then
        echo "searchbench-check-pkl-generated: pkl not found on PATH" >&2
        exit 1
      fi
      pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl --output-path=. configs/schema/SearchBenchRound.pkl
      if git diff --quiet -- internal/adapters/config/pkl/generated/; then
        exit 0
      fi
      echo "searchbench-check-pkl-generated: Pkl Go bindings are out of date. Regenerate with:" >&2
      echo '  pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl --output-path=. configs/schema/SearchBenchRound.pkl' >&2
      git diff -- internal/adapters/config/pkl/generated/ >&2 || true
      git checkout -- internal/adapters/config/pkl/generated/
      exit 1
    '';
    runtimeInputs = [ pkgs.pkl ];
  };

  searchbench-check-templ-generated = mkInRepo {
    name = "searchbench-check-templ-generated";
    text = ''
      if ! command -v templ >/dev/null; then
        echo "searchbench-check-templ-generated: templ not found on PATH" >&2
        exit 1
      fi
      templ generate -path ./internal/agents
      if git diff --quiet -- \
        internal/agents/evaluator/prompt/prompt_templ.go \
        internal/agents/optimizer/prompt/prompt_templ.go
      then
        exit 0
      fi
      echo "searchbench-check-templ-generated: templ outputs are out of date. Run: templ generate -path ./internal/agents" >&2
      git diff -- \
        internal/agents/evaluator/prompt/prompt_templ.go \
        internal/agents/optimizer/prompt/prompt_templ.go >&2 || true
      git checkout -- \
        internal/agents/evaluator/prompt/prompt_templ.go \
        internal/agents/optimizer/prompt/prompt_templ.go
      exit 1
    '';
    runtimeInputs = [ pkgs.templ ];
  };

  searchbench-check-generated = mkInRepo {
    name = "searchbench-check-generated";
    text = ''
      ${searchbench-check-pkl-generated}/bin/searchbench-check-pkl-generated
      ${searchbench-check-templ-generated}/bin/searchbench-check-templ-generated
    '';
    runtimeInputs = [
      searchbench-check-pkl-generated
      searchbench-check-templ-generated
    ];
  };

  searchbench-update-repomix = mkInRepo {
    name = "searchbench-update-repomix";
    text = ''
      repomix \
        --output repomix-output.xml \
        --style xml \
        --compress \
        --include-diffs \
        --include-logs \
        --include-logs-count 10
      git add repomix-output.xml
    '';
    runtimeInputs = [ pkgs.repomix ];
  };

  searchbench-nix-flake-check = mkInRepo {
    name = "searchbench-nix-flake-check";
    text = ''
      exec nix flake check
    '';
    runtimeInputs = [ pkgs.nix ];
  };

  searchbench-vocabulary-check = mkInRepo {
    name = "searchbench-vocabulary-check";
    text = ''
      set +e
      hits=0
      for pat in "local e2e" "projection as product noun" "evaluation as public app workflow" "optimizer as public app workflow"; do
        if rg -n --glob '*.md' --glob '*.go' --glob '!repomix-output.xml' "$pat" AGENTS.md README.md docs internal/app internal/surface 2>/dev/null; then
          hits=1
        fi
      done
      if [[ $hits -ne 0 ]]; then
        echo "searchbench-vocabulary-check: warning — review matches above for retired or risky architecture language (non-blocking)." >&2
      fi
      exit 0
    '';
    runtimeInputs = [ pkgs.ripgrep ];
  };

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
      ${searchbench-e2e}/bin/searchbench-e2e
      nix flake check
      ${searchbench-check-generated}/bin/searchbench-check-generated
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
      searchbench-e2e
      searchbench-check-generated
      searchbench-update-repomix
    ];
  };

  searchbench-go-test-all-push = pkgs.writeShellApplication {
    name = "searchbench-go-test-all-push";
    text = ''
      set -euo pipefail
      exec ${searchbench-go-test-all}/bin/searchbench-go-test-all
    '';
    runtimeInputs = [ searchbench-go-test-all ];
  };

  searchbench-nix-flake-check-push = pkgs.writeShellApplication {
    name = "searchbench-nix-flake-check-push";
    text = ''
      set -euo pipefail
      exec ${searchbench-nix-flake-check}/bin/searchbench-nix-flake-check
    '';
    runtimeInputs = [ searchbench-nix-flake-check ];
  };

  searchbench-e2e-push = pkgs.writeShellApplication {
    name = "searchbench-e2e-push";
    text = ''
      set -euo pipefail
      exec ${searchbench-e2e}/bin/searchbench-e2e
    '';
    runtimeInputs = [ searchbench-e2e ];
  };

in
{
  inherit
    searchbench-go-test-all
    searchbench-go-test-all-push
    searchbench-e2e
    searchbench-e2e-push
    searchbench-architecture-check
    searchbench-check-pkl-generated
    searchbench-check-templ-generated
    searchbench-check-generated
    searchbench-update-repomix
    searchbench-nix-flake-check
    searchbench-nix-flake-check-push
    searchbench-vocabulary-check
    searchbench-agent-start
    searchbench-agent-check
    searchbench-agent-pack
    searchbench-agent-merge-check
    ;
}
