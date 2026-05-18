# Go Buck layout (prelude + gobuckify)

SearchBench-Go is modeled in Buck with prelude `go_library` / `go_test` / `go_binary`. Third-party deps are projected explicitly with [gobuckify](https://buck2.build/docs/users/languages/go/third_party_packages/).

## Public Buck surface

| Layer | Targets | Role |
| --- | --- | --- |
| Native fast suite | `//src/searchbench-go:go_native_fast` | Deterministic prelude `go_test` targets (no live/network) |
| Native full suite | `//src/searchbench-go:go_native_full` | `go_native_fast` + heavier deterministic targets + round manifest validation |
| Public check gate | `//src/searchbench-go:check` | Alias of `go_native_full` |
| CLI artifact | `//src/searchbench-go:searchbench`, `//src/searchbench-go/cmd/searchbench:searchbench` | Native prelude `go_binary` |
| Pkl regen | `//src/searchbench-go:pkl_go_types` | Run target that refreshes checked-in Go bindings |

## Explicit regeneration flows

From the repo root (`nix develop` recommended):

```bash
nix run .#project-go-deps
```

This runs `go mod vendor` and `gobuckify` under `src/searchbench-go/` and writes `vendor/.searchbench-vendor-projection.json` (input hashes + tool versions). The `vendor/` tree is **not** committed.

```bash
buck2 run //src/searchbench-go:pkl_go_types
buck2 test //src/searchbench-go:pkl_go_types_check
```

Buck tests do not run vendor projection implicitly, and `pkl_go_types_check` diffs a scratch output tree instead of mutating checked-in sources.

## Gobuckify / vendor status

- The repo does not check in Buck prelude sources.
- `.buckconfig` pins `prelude` as a **git external cell** (`buck2-prelude` commit `169c93f58201d1bb62370ba0563b293e139bb2c7`). Buck fetches it on demand; no repo-local `prelude/` tree or symlink is required.
- The prelude **bundled** inside the Nix `buck2` package is older and still references `prelude.go_unittest` in `gobuckify`; do not switch `[external_cells] prelude` back to `bundled` without upgrading `buck2` first.
- `src/searchbench-go/gobuckify.json` enables `generate_embed_srcs` so vendored packages using `go:embed` emit valid Buck targets.
- Generated third-party targets live under `//src/searchbench-go/vendor/...` after `nix run .#project-go-deps`.
- `vendor/` is gitignored; refresh when lockfiles or `gobuckify.json` change.

## Native test runtime (libstdc++)

Prelude `go_test` binaries link `libstdc++.so.6` with `RUNPATH` `$PROJECT_ROOT/outputs/out/lib`. `nix develop` creates `outputs/out/lib` symlinks into the Nix `stdenv` GCC runtime (ignored by git). Without `nix develop`, those tests fail at load time even when builds succeed.

Round/live targets that shell out to the CLI also use `//tools:libstdcxx_libdir` (same path) via `searchbench_round_op`.

## Go test coverage audit

Status key: **A** covered in suite · **B** has `go_test`, not in suite · **C** has `*_test.go`, no `go_test` yet · **D** intentionally excluded (live/manual/cgo/heavy) · **E** blocked or deferred.

| Package | Test files | Buck target | Suite | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| `internal/architecture` | `imports_test.go` | `:imports_boundary_test` | fast | A | Import-boundary proof |
| `internal/pure/domain` | `*_test.go` (4) | `:domain_test` | fast | A | Pure unit |
| `internal/pure/codegraph` | `localdistance_test.go` | `:codegraph_test` | fast | A | Pure unit |
| `internal/pure/liveconfig` | `dotenv_test.go`, `fingerprint_test.go` | `:liveconfig_test` | fast | A | Deterministic config helpers |
| `internal/pure/optimizer` | `workspace_test.go` | `:optimizer_test` | fast | A | Pure unit |
| `internal/pure/report` | `evidence_test.go`, `match_execution_set_test.go` | `:report_test` | fast | A | Pure unit |
| `internal/pure/score` | `*_test.go` (5) | `:score_test` | fast | A | Pure unit |
| `internal/pure/usage` | `collector_test.go`, `hash_registry_test.go` | `:usage_unit_test` | fast | A | Pure unit |
| `internal/games/codelocalization` | `game_test.go` | `:codelocalization_test` | fast | A | Game rules |
| `internal/ports/pipeline` | `pipeline_test.go` | `:pipeline_test` | fast | A | Port contract |
| `internal/adapters/materialize/git` | `materialize_test.go` | `:git_test` | fast | A | Git materialize |
| `internal/adapters/pipeline/exec` | `ic_allowlist_test.go`, `runner_test.go` | `:exec_test` | fast | A | Pipeline exec |
| `internal/adapters/workspace/*` | `*_test.go` | `:buckdescriptor_test`, etc. | fast | A | Workspace adapters |
| `internal/adapters/report/text` | `report_test.go` | `:text_test` | fast | A | Needs `outputs/out/lib` at runtime |
| `internal/adapters/trace/langsmith` | `factory_test.go` | `:langsmith_test` | fast | A | Factory wiring only |
| `internal/agents/evaluator/eino/callbacks` | `callbacks_test.go`, `usage_test.go` | `:callbacks_test` | fast | A | Test-only `domain` dep |
| `internal/agents/evaluator/policy` | `evaluator_test.go` | `:policy_test` | fast | A | Pkl-backed |
| `internal/agents/evaluator/prompt` | `prompt_test.go` | `:prompt_test` | fast | A | Templ render |
| `internal/agents/optimizer/prompt` | `prompt_test.go` | `:prompt_test` | fast | A | Templ render |
| `internal/surface/logging` | `attrs_test.go`, `events_test.go` | `:logging_test` | fast | A | Surface logging |
| `internal/adapters/config/pkl` | `load_test.go`, `validate_test.go`, `workspace_seed_test.go` | `:pkl_test` | full | A | Pkl + fixtures |
| `internal/adapters/dataset/lca` | `export_test.go`, `source_test.go` | `:lca_test` | full | A | Dataset adapter |
| `internal/adapters/providers/evaluatormodel` | `*_test.go` (2) | `:evaluatormodel_test` | full | A | Provider factory |
| `internal/agents/evaluator/eino` | `evaluator_test.go`, `retry_test.go` | `:eino_test` | full | A | Eino evaluator |
| `internal/agents/optimizer/eino` | `optimizer_test.go` | `:eino_test` | full | A | Eino optimizer |
| `internal/agents/optimizer/policy` | `candidate_pipeline_test.go`, etc. (3) | `:policy_test` | full | A | Optimizer policy |
| `internal/testing/modeltest` | `openai_server_test.go`, `scripted_model_test.go` | `:modeltest_test` | full | A | HTTP test fixtures (`embed_srcs`) |
| `configs/rounds/live-ic-vs-jcodemunch` | (round Pkl) | `:validate` | full | A | Manifest validation (not Go) |
| `internal/adapters/backend/iterativecontext` | `candidate_launch_test.go`, `iterativecontext_test.go` | — | — | E | Native `go_library` only; runtime tests need MCP/IC + `libstdc++` — add `go_test` when runner/toolchain story is stable |
| `internal/adapters/backend/jcodemunch` | `jcodemunch_test.go` | — | — | E | Same as IC backend |
| `internal/adapters/bundle/fs` | `continuation_test.go`, `validate_test.go`, `writer_test.go` | — | — | C | Deterministic; candidate for `go_native_full` (Pkl + bundle I/O) |
| `internal/adapters/codegraph/treesitter` | `index_test.go` | — | — | D | tree-sitter / cgo; keep out of fast gates until native graph is ready |
| `internal/adapters/scoring/pkl` | `run_test.go` | — | — | C | Pkl runtime; add `:pkl_scoring_test` under **full** when needed |
| `internal/app/round` | many `*_test.go` | — | — | D | Large integration surface; prove via round Buck targets + `validate_bundle`, not package `go_test` in `check` |
| `internal/app/round/internal/compare` | `*_test.go` (4) | — | — | D | Subpackage of round orchestration |
| `internal/surface/cli` | `buck_test.go`, `cli_test.go` | — | — | C | CLI manifest tests; partially covered by `//configs/rounds/live-ic-vs-jcodemunch:validate` |
| `(module root)` | `e2e_test.go` | — | — | D | Live e2e; use round live Buck targets |
| `(module root)` | `live_e2e_test.go` | — | — | D | Removed from Buck gates; see `docs/reference/run-entrypoints.md` |

`internal/testing/importcheck` exposes `:importcheck` (library helper), not a `go_test` suite member.

## Suite membership (`src/searchbench-go/BUCK`)

- **`go_native_fast`**: 23 deterministic package `go_test` targets (labels: `fast`, `deterministic`, `no-cgo` where set).
- **`go_native_full`**: `go_native_fast` + `pkl_test`, `lca_test`, `evaluatormodel_test`, both `eino_test`, `optimizer/policy_test`, `modeltest_test`, and `//configs/rounds/live-ic-vs-jcodemunch:validate`.
- **`check`**: alias of `go_native_full` (native Go graph + deterministic round manifest check).

Live/network MCP evaluation stays on explicit round targets (`live_smoke`, `run`, etc.) — not in `go_native_*` or `//:check`.
