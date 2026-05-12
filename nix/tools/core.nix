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
in
{
  inherit
    searchbench-go-test-all
    searchbench-e2e
    searchbench-architecture-check
    searchbench-update-repomix
    searchbench-nix-flake-check
    searchbench-vocabulary-check
    ;
}
