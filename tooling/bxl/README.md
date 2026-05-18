# SearchBench BXL work-graph planners

Internal Buck2/BXL layer that emits **JSON plans** from **Buck graph traversal** (`uquery` / configured target attrs). No static registry. Does not run evaluations, call models, or mutate Git.

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

buck2 bxl //tooling/bxl/searchbench.bxl:evaluation_matrix
```

Changed paths are passed on the CLI (`--changed-file` repeatable or `--changed-files-text`); BXL has no `read_text` for host files.

## Layout

| Path | Role |
|------|------|
| `searchbench.bxl` | `bxl_main` entrypoints and JSON document builders |
| `graph.bzl` | `uquery` / `rdeps` / `owner` / configured-attr traversal helpers |
| `schemas/` | JSON Schema for each `kind` |
| `fixtures/` | Golden JSON samples from local runs |

## Graph traversal (summary)

| Planner | Graph operations |
|---------|------------------|
| `list_backends` | `attrfilter(name, optimizable_backend, //...)` |
| `resolve_backend` | backend discovery + `owner` of descriptor JSON |
| `proof_plan` | `searchbench_round_op` attrs (`manifest`, `mode`); else `go_test` resources + `test_suite` membership |
| `affected_plan` | `owner` + `targets_in_buildfile` seeds → `rdeps` on gate suites; round ops by `manifest` / `manifest_dir` |
| `evaluation_matrix` | round ops `manifest_dir` + `//:searchbench_go_test_resources` `srcs` |

Proof `rdeps` uses a **safe universe** of `test_suite` gate targets only (`//:check`, `go_native_fast`, IC `check`, etc.) because full-graph `rdeps(//..., go_test)` hits a toolchain resolution error in this repo.

## Extending

1. Add a `genrule` (or similar) named `optimizable_backend` — discovered automatically.
2. Add `searchbench_round` ops under `configs/rounds/<name>/` — discovered via `kind(searchbench_round_op)`.
3. Wire Pkl-only rounds into `//:searchbench_go_test_resources` `srcs` globs if they lack round ops.

Go SearchBench still resolves backends via `BuckDescriptorProvider`; BXL is an optional planner for meta-harness / agent traces.

See also [docs/reference/buck-work-graph.md](../../docs/reference/buck-work-graph.md) and [docs/research/bxl-meta-harness.md](../../docs/research/bxl-meta-harness.md).
