# Freshness checks for Pkl and templ outputs.
{ pkgs, mkInRepo }:
let
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
in
{
  inherit
    searchbench-check-pkl-generated
    searchbench-check-templ-generated
    searchbench-check-generated
    ;
}
