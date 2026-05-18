# Buck work graph (BXL planners)

SearchBench exposes an **internal** work-graph layer: Buck2/BXL scripts under `tooling/bxl/` that print versioned JSON documents. This is for contributors, meta-harness experiments, and future optimizer traces—not required for running rounds.

## Document kinds

| `kind` | Purpose |
|--------|---------|
| `searchbench.bxl_target_summary.v1` | Smoke: label + coarse target class |
| `searchbench.backend_resolution.v1` | Descriptor path, launcher, validator metadata |
| `searchbench.backend_inventory.v1` | Known optimizable backend targets |
| `searchbench.proof_plan.v1` | Minimal / acceptable / fallback / too-live proof targets |
| `searchbench.affected_plan.v1` | Heuristic impact from changed paths |
| `searchbench.evaluation_matrix.v1` | Rounds × backends comparison matrix |

Schemas: `tooling/bxl/schemas/*.schema.json`.

## Proof tiers

Proof plans classify Buck targets for future **proof distance** metrics (how far an agent’s chosen checks were from the minimal correct set):

- **minimal** — smallest known validation for a round (e.g. `//configs/rounds/live-ic-vs-jcodemunch:validate` for optimize-ic today).
- **acceptable** — minimal plus repo/round checks (`//:check`, `//src/searchbench-go:check`, bundle validate).
- **fallback** — broad gate (`//:check_full`).
- **too_live_for_default_gate** — live smoke, full run, provider-backed evaluate targets.

## Split of responsibilities

```text
Buck targets     → stable graph-addressed capabilities
BXL (tooling/bxl)→ JSON plans over those capabilities
Go SearchBench   → lifecycle, evaluation, scoring, bundles
Meta-harness     → branches, worktrees, retries (out of repo)
```

BXL does **not** replace Pkl manifests, Go round runners, or public CLI workflows.

## Running planners

From repo root inside `nix develop`:

```bash
buck2 bxl //tooling/bxl/searchbench.bxl:list_backends
```

Full command list: [tooling/bxl/README.md](../../tooling/bxl/README.md).

## Graph traversal (not a static registry)

Planners query the Buck graph via BXL `uquery` and configured target attributes:

- **Backends:** `attrfilter(name, optimizable_backend, //...)`
- **Rounds with Buck ops:** `kind(searchbench_round_op, //...)` + `manifest` / `manifest_dir` attrs
- **Pkl-only rounds:** `//:searchbench_go_test_resources` configured `srcs` (expanded artifacts)
- **Affected proof targets:** `owner` / `targets_in_buildfile` seeds → `rdeps` on gate `test_suite` targets; `go_test` `resources` scanning for filegroup deps

`rdeps` universes are limited to gate suites because traversing arbitrary `go_test` / `searchbench_round_op` deps can fail on `toolchains//:cxx_no_default_deps` in this workspace.

Research context: [bxl-meta-harness.md](../research/bxl-meta-harness.md).
