# Bundles

A **bundle** is the durable artifact tree for one completed round. It is the product output reviewers and tools inspect.

**Write path:** `src/searchbench-go/internal/adapters/bundle/fs`
**Models:** `src/searchbench-go/internal/pure/report`, `src/searchbench-go/internal/pure/score`

## Example tree

**Path:** `configs/rounds/local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001/`

After `./searchbench run`, the same shape appears under `{bundle-root}/games/code-localization/rounds/round-001/`.

```text
COMPLETE
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
| `resolved-round.json` | Fully resolved manifest + config snapshot |
| `round-report.json` / `.txt` | Human- and machine-readable comparison report |
| `evidence.pkl` | Evidence document for Pkl scoring |
| `objective.json` | Result of `localization-objective.pkl` |
| `decision.json` | `PROMOTE_CHALLENGER` / `REVIEW` / `REJECT` |
| `metadata.json` | Bundle ids, hashes, provenance |
| `continuation.json` / `.pkl` | Survivor state for the next round manifest |
| `policies/` | Staged challenger (and related) policy files |

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

## Immutability

Bundles are not rewritten after completion. New rounds get new directories and ids.
