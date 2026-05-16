# Pkl objectives (scoring)

**Go types:** `internal/pure/score` · **Runner:** `internal/adapters/scoring/pkl`

## Flow

1. Go builds `score.RoundEvidenceDocument` from the comparison report.
2. Persist `evidence.pkl` in the bundle.
3. Pkl objective module evaluates evidence → named values + final selection.
4. Go validates `score.ObjectiveResult` and writes `objective.json`.

Pkl owns **scoring math** on explicit evidence inputs; Go owns evidence construction and validation.

## Lifecycle in a round

Evidence projection happens after match execution, before bundle finalization. Invalid objective output fails the round.

## Non-goals

Objective Pkl does **not** define round manifests, datasets, evaluator backends, or optimizer wiring — only scoring over resolved evidence.

## Example modules

See objective modules referenced from round manifests under `configs/rounds/`.
