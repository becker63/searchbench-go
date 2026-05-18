# SearchBench BXL work-graph planners

Internal Buck2/BXL layer that emits **JSON plans** about backends, rounds, proof targets, and change impact. It does not run evaluations, call models, or mutate Git.

Implemented for [issue #94](https://github.com/becker63/searchbench-go/issues/94).

## Prerequisites

```bash
nix develop
# optional if Go vendor projection is stale:
nix run .#project-go-deps
```

## Commands

Entrypoints live in `searchbench.bxl` (label suffix is the function name):

```bash
buck2 bxl //tooling/bxl/searchbench.bxl:target_summary -- \
  --target //src/iterative-context:optimizable_backend

buck2 bxl //tooling/bxl/searchbench.bxl:resolve_backend -- \
  --backend //src/iterative-context:optimizable_backend

buck2 bxl //tooling/bxl/searchbench.bxl:list_backends

buck2 bxl //tooling/bxl/searchbench.bxl:proof_plan -- \
  --target //configs/rounds/optimize-ic:round

buck2 bxl //tooling/bxl/searchbench.bxl:affected_plan -- \
  --changed-file configs/rounds/optimize-ic/round.pkl \
  --changed-file src/iterative-context/src/iterative_context/policy.py

# or inline:
buck2 bxl //tooling/bxl/searchbench.bxl:affected_plan -- \
  --changed-files-text $'configs/rounds/optimize-ic/round.pkl\nsrc/iterative-context/src/iterative_context/policy.py'

buck2 bxl //tooling/bxl/searchbench.bxl:evaluation_matrix
```

**Note:** `ctx.fs` in BXL has no `read_text`; changed paths are passed on the CLI (repeat `--changed-file` or `--changed-files-text`), not via a host file path.

## Layout

| Path | Role |
|------|------|
| `searchbench.bxl` | `bxl_main` entrypoints, JSON output via `ctx.output.print_json` |
| `planner.bzl` | Builds plan documents from registry data |
| `registry.bzl` | Static catalogs: backends, rounds, path rules, proof tiers |
| `schemas/` | JSON Schema for each `kind` |
| `fixtures/` | Golden JSON samples from local runs |

## Extending

1. Add optimizable backends to `BACKEND_CATALOG` in `registry.bzl` (and a real `optimizable_backend.json` in-tree).
2. Add rounds with manifest paths and Buck validate targets to `ROUND_CATALOG`.
3. Extend `PATH_RULES` for `affected_plan` prefix heuristics (conservative false positives are OK).
4. Regenerate fixtures after behavior changes.

Go SearchBench still resolves backends via `BuckDescriptorProvider`; BXL is an optional planner for meta-harness / agent traces.

## Findings (initial)

- BXL is viable in this repo: deterministic JSON, no runtime side effects.
- Backend resolution is **catalog + on-disk descriptor path**, not a live Buck query of rule attributes.
- `optimize-ic` has a Pkl manifest but no dedicated round BUCK package; proof plans reuse `//configs/rounds/live-ic-vs-jcodemunch:validate`.
- jCodeMunch `optimizable_backend` is listed aspirationally in the evaluation matrix only.
- Affected planning uses prefix rules in Starlark, not full Buck graph traversal.

See also [docs/reference/buck-work-graph.md](../../docs/reference/buck-work-graph.md) and [docs/research/bxl-meta-harness.md](../../docs/research/bxl-meta-harness.md).
