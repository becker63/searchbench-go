# Tests, lint helpers, Repomix, flake check, vocabulary warning.
{ pkgs, mkInRepo }:
let
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

  searchbench-update-repomix = mkInRepo {
    name = "searchbench-update-repomix";
    text = ''
      repomix \
        --output repomix-output.xml \
        --style xml \
        --compress \
        --no-git-sort-by-changes
      git add repomix-output.xml
    '';
    runtimeInputs = [ pkgs.repomix ];
  };

  searchbench-repomix-fresh-check = mkInRepo {
    name = "searchbench-repomix-fresh-check";
    text = ''
      ${searchbench-update-repomix}/bin/searchbench-update-repomix

      if [[ ! -f repomix-output.xml ]]; then
        echo "searchbench-repomix-fresh-check: repomix-output.xml missing after regeneration" >&2
        exit 1
      fi

      if grep -qE '^<file path="repomix-output\.xml"' repomix-output.xml; then
        echo "searchbench-repomix-fresh-check: repomix-output.xml appears to include itself (check repomix ignore patterns)" >&2
        exit 1
      fi
      if grep -qE '^<file path="configs/rounds/[^"]+/artifacts/' repomix-output.xml; then
        echo "searchbench-repomix-fresh-check: repomix-output.xml includes configs/rounds/*/artifacts (should be excluded)" >&2
        exit 1
      fi
      if grep -q '^<file path="internal/adapters/config/pkl/generated/' repomix-output.xml; then
        echo "searchbench-repomix-fresh-check: repomix-output.xml includes generated Pkl paths (should be excluded)" >&2
        exit 1
      fi

      if ! git diff --quiet -- repomix-output.xml || ! git diff --quiet --cached -- repomix-output.xml; then
        echo "searchbench-repomix-fresh-check: repomix-output.xml is not committed at HEAD after regeneration." >&2
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
    '';
    runtimeInputs = [ pkgs.gnugrep ];
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

  searchbench-staticcheck = mkInRepo {
    name = "searchbench-staticcheck";
    text = ''
      exec staticcheck ./...
    '';
    runtimeInputs = [
      pkgs.go
      pkgs.go-tools
    ];
  };

  searchbench-golangci = mkInRepo {
    name = "searchbench-golangci";
    text = ''
      exec golangci-lint run ./...
    '';
    runtimeInputs = [
      pkgs.go
      pkgs.golangci-lint
    ];
  };

  searchbench-go-mod-tidy-check = mkInRepo {
    name = "searchbench-go-mod-tidy-check";
    text = ''
      go mod tidy
      if [[ -f go.sum ]]; then
        git diff --exit-code -- go.mod go.sum
      else
        git diff --exit-code -- go.mod
      fi
    '';
    runtimeInputs = [ pkgs.go ];
  };

  searchbench-prompt-contract-check = mkInRepo {
    name = "searchbench-prompt-contract-check";
    text = ''
      exec go test ./internal/agents/evaluator/prompt ./internal/agents/optimizer/prompt
    '';
    runtimeInputs = [ pkgs.go ];
  };

  searchbench-go-test-race = mkInRepo {
    name = "searchbench-go-test-race";
    text = ''
      exec go test -race ./...
    '';
    runtimeInputs = [ pkgs.go ];
  };

  searchbench-refresh-pkl-example-fixtures = mkInRepo {
    name = "searchbench-refresh-pkl-example-fixtures";
    text = ''
      root="$(pwd)"
      mkdir -p "$root/.tmp"
      GOCACHE="$root/.tmp/go-cache" SEARCHBENCH_WRITE_EXAMPLE_FIXTURES=1 \
        go test ./internal/app/round -run TestWriteExampleRoundFixtures -count=1
    '';
    runtimeInputs = [
      pkgs.go
      pkgs.pkl
    ];
  };

  searchbench-go-build-root = mkInRepo {
    name = "searchbench-go-build-root";
    text = ''
      exec go build -o searchbench ./cmd/searchbench
    '';
    runtimeInputs = [ pkgs.go ];
  };

  searchbench-no-scripts-check = mkInRepo {
    name = "searchbench-no-scripts-check";
    text = ''
      if [[ -d scripts ]]; then
        echo "searchbench-no-scripts-check: scripts/ exists; project commands should be Nix-defined tools under nix/tools" >&2
        find scripts -maxdepth 2 -type f -print >&2
        exit 1
      fi
    '';
    runtimeInputs = [ pkgs.findutils ];
  };
in
{
  inherit
    searchbench-go-test-all
    searchbench-e2e
    searchbench-architecture-check
    searchbench-update-repomix
    searchbench-repomix-fresh-check
    searchbench-nix-flake-check
    searchbench-vocabulary-check
    searchbench-staticcheck
    searchbench-golangci
    searchbench-go-mod-tidy-check
    searchbench-prompt-contract-check
    searchbench-go-test-race
    searchbench-refresh-pkl-example-fixtures
    searchbench-go-build-root
    searchbench-no-scripts-check
    ;
}
