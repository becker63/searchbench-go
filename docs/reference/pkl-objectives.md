# Pkl objectives (scoring)

**Pkl owns scoring math on evidence. Go builds evidence and validates output.**

| Concern | Path |
| --- | --- |
| Objective schema | `configs/schema/SearchBenchObjective.pkl` |
| Example module | `configs/rounds/local-ic-vs-jcodemunch/scoring/localization-objective.pkl` |
| Go types | `src/searchbench-go/internal/pure/score/` |
| Runner | `src/searchbench-go/internal/adapters/scoring/pkl/` |

## Example objective

**File:** `configs/rounds/local-ic-vs-jcodemunch/scoring/localization-objective.pkl`

```pkl
objectiveId = "localization-v1"

local challengerQuality =
  1.0 - (if (current.localizationDistance.goldHop.challenger < maxHop) ...)

local finalScore = base * regressionPenalty * invalidPredictionPenalty

values = new {
  helpers.intermediate("challengerQuality", challengerQuality)
  helpers.penalty("regressionPenalty", regressionPenalty)
  helpers.finalValue("final", finalScore)
}

final = "final"
```

| Piece | Role |
| --- | --- |
| `current.*` | Evidence fields from the round report |
| `helpers.intermediate` / `helpers.penalty` | Named intermediate values |
| `helpers.finalValue("final", …)` | Selected score |
| `final = "final"` | Which value is the round score |

Referenced from the round manifest:

```pkl
scoring = game.objective("scoring/localization-objective.pkl")
```

## Lifecycle

```text
matches → round-report → evidence.pkl → Pkl objective → objective.json
```

**Output example:** `configs/rounds/local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001/objective.json`

```json
{
  "objective_id": "localization-v1",
  "values": [
    { "name": "challengerQuality", "value": 0.833..., "kind": "intermediate" },
    { "name": "final", "value": 0.864..., "kind": "final" }
  ]
}
```

Invalid objective output fails the round in Go.

## Non-goals

Objective Pkl does **not** define round manifests, datasets, evaluator backends, or optimizer wiring.
