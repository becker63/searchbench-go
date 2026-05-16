# Pkl round manifests

**Pkl is the manifest surface. Go owns execution semantics.**

| Concern | Path |
| --- | --- |
| Schema | `configs/schema/SearchBenchRound.pkl` |
| Game helpers | `configs/schema/games/code-localization-helpers.pkl` |
| Go loader + validation | `src/searchbench-go/internal/adapters/config/pkl/` |
| Round execution | `src/searchbench-go/internal/app/round/` |

**Proves with:** `buck2 test //src/searchbench-go:check` · after schema edits: `buck2 build //src/searchbench-go:pkl_go_types` then `buck2 test //src/searchbench-go:pkl_go_types_check`

## From-scratch round

**File:** `configs/rounds/local-ic-vs-jcodemunch/round.pkl`

```pkl
amends "../../schema/games/code-localization.pkl"
import "../../schema/games/code-localization-helpers.pkl" as game

round = (game.defineFromScratch("round-001")) {
  incumbent = game.jcodemunch()
  challenger = (game.iterativeContext("policies/challenger_policy.py")) {
    selectionPolicy { id = "challenger-policy-round-001" }
  }
  matches = game.lca("py", "dev", 5)
  evaluator = game.fakeEvaluator()
  scoring = game.objective("scoring/localization-objective.pkl")
}
```

| Field | Role |
| --- | --- |
| `defineFromScratch` | New round id and bundle layout |
| `incumbent` / `challenger` | Policies (interfaces) to compare |
| `matches` | Dataset slice |
| `evaluator` | How matches run (fake = deterministic offline) |
| `scoring` | Pkl objective module path |

**Also see:** `configs/rounds/fake-local-e2e/round.pkl` — same shape with explicit fake backends.

## Continuation round

**File:** `configs/rounds/optimize-ic/round.pkl`

```pkl
amends "../local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001/continuation.pkl"
import "../../schema/games/code-localization-helpers.pkl" as game

round {
  id = "round-002"
  challenger {
    generate {
      optimizer = game.fakeOptimizer()
      artifactName = "next_challenger_policy.round-002.py"
    }
  }
}
```

**Meaning:** `amends` loads survivor state from the prior bundle’s `continuation.pkl`; only the delta (new challenger generation) is authored.

**Output bundle:** `configs/rounds/optimize-ic/artifacts/games/code-localization/rounds/round-002/`

## Flow

```text
round.pkl → pkl-go → Go config validation → app/round → bundle/
```

Do not mutate a completed round’s manifest. New rounds use new ids and bundle directories.

## Workspace seed (optional)

In manifests that materialize IC for optimization:

```pkl
runtime {
  workspaceSeed {
    provider = "local_path"
    localPath = "src/iterative-context"
  }
}
```

Details: [../candidate-workspaces.md](../candidate-workspaces.md).

## Regenerate Go bindings

Debugging fallback (canonical: Buck targets in [development.md](../development.md)):

```bash
cd src/searchbench-go
pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl \
  --output-path=. ../../configs/schema/SearchBenchRound.pkl
```
