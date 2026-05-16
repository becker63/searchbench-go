# Pkl round manifests

**Schema:** `configs/schema/SearchBenchRound.pkl` · **Go loader:** `internal/adapters/config/pkl` · **Execution:** `internal/app/round`

## Flow

1. Author `round.pkl` (human-facing).
2. `pkl-go` → typed Go config; Go validates cross-field rules.
3. App resolves manifest → runs round → writes **bundle**.

Pkl is the **manifest surface**; Go owns semantics, validation, and execution.

## Immutability

Do not mutate a completed round’s manifest. New rounds use new manifests and bundle IDs.

## Shape (summary)

Typical manifest fields include round identity, game binding, dataset/match selection, incumbent and challenger policies, backends, scoring module reference, and optional optimizer / workspace-seed blocks. See `configs/rounds/*/round.pkl` for examples.

## Workspace seeds

`runtime.workspaceSeed` — `local_path` or `buck_descriptor`. Details: [../workspace-seeds.md](../workspace-seeds.md).

## Regenerate Go bindings

```bash
cd src/searchbench-go
pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl \
  --output-path=. ../../configs/schema/SearchBenchRound.pkl
```
