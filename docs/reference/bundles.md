# Bundles

A **bundle** is the durable artifact tree for one completed round. It is the product output reviewers and tools inspect.

**Write path:** `src/searchbench-go/internal/adapters/bundle/fs`
**Models:** `src/searchbench-go/internal/pure/report`, `src/searchbench-go/internal/pure/score`

## First files to inspect

Open **`report.json`**, then **`report.txt`**. These are the canonical human/agent summaries (mode, freshness, pass/fail, failure counts, attempt aggregates). Detailed evidence lives in `round-report.json` and related files.

## Example tree

**Path:** `configs/rounds/local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001/`

The same shape appears under `{bundle-root}/games/code-localization/rounds/<round-id>/` after a repo-owned Buck run.

```text
COMPLETE
report.json
report.txt
resolved-round.json
round-report.json
round-report.txt
evidence.pkl
objective.json
decision.json
metadata.json
continuation.json
continuation.pkl
policies/challenger_policy.py
```

| File | Role |
| --- | --- |
| `COMPLETE` | Marker that the round finished |
| `report.json` / `report.txt` | **Canonical** summary — inspect first |
| `resolved-round.json` | Fully resolved manifest + config snapshot |
| `round-report.json` / `.txt` | Detailed comparison report (evidence-level) |
| `evidence.pkl` | Evidence document for Pkl scoring |
| `objective.json` | Result of `localization-objective.pkl` |
| `decision.json` | `PROMOTE_CHALLENGER` / `REVIEW` / `REJECT` (or `NO_DECISION` for stability probes) |
| `metadata.json` | Bundle ids, hashes, provenance |
| `continuation.json` / `.pkl` | Survivor state for the next round manifest |
| `policies/` | Staged challenger (and related) policy files |

## `evaluate_n` attempts layout

Multi-attempt promotion evaluation (`buck2 run //configs/rounds/live-ic-vs-jcodemunch:evaluate_n`) publishes a consolidated bundle plus per-attempt artifacts:

```text
report.json          # consolidated metrics and promotion gate
report.txt
attempts/
  attempt-001/
    report.json      # per-attempt canonical summary
    ...              # raw round bundle / evidence for that attempt
  attempt-002/
    report.json
    ...
metadata.json
COMPLETE
```

Each attempt directory records attempt id, input fingerprint, request/response hashes where available, result status, and failure classification. Failed attempts are included in the tree and in consolidated `report.json` aggregates; promotion uses median/consolidated metrics, not the best single attempt.

## Short excerpts

**decision.json:**

```json
{
  "decision": "PROMOTE_CHALLENGER",
  "reason": "challenger improves the composite score in local fake comparison"
}
```

**continuation.json** (start):

```json
{
  "schema_version": "searchbench.continuation.v1",
  "bundle_id": "round-001",
  "game": { "id": "code-localization", "kind": "code_localization" }
}
```

**Next round input:** `configs/rounds/optimize-ic/round.pkl` amends `continuation.pkl` from this bundle.

## Round 002 bundle

Optimizer continuation example:

`configs/rounds/optimize-ic/artifacts/games/code-localization/rounds/round-002/`

Adds e.g. `policies/next_challenger_policy.round-002.py` from `game.fakeOptimizer()`.

Optimizer rounds may also include `attempts/attempt-NNN-prompt.txt` and `attempts/attempt-NNN-result.json` for policy-generation retries (separate from live `evaluate_n` attempt trees).

## Immutability

Bundles are not rewritten after completion. New rounds get new directories and ids.
