# Optimizer policy validation (Iterative Context)

SearchBench-Go treats **NextChallenger** proposals as **file-backed artifacts**. For manifests targeting **`iterative_context.selection_policy.v1`**, the default optimizer validator runs a **typed local pipeline** (see `internal/ports/pipeline` and `internal/adapters/pipeline/exec`) rather than a single `py_compile` check.

## Why `py_compile` is insufficient

Syntax-only validation does not prove:

- the expected callable exists and matches the IC surface,
- Iterative Context can load/install/verify the policy,
- the Iterative Context repo still passes its own type/lint/test gates.

The harness therefore mirrors the old Python SearchBench idea: **structured steps**, **typed `StepResult`s**, **classification**, and **bounded retry feedback** (`pipeline.FormatPipelineFeedback`).

## MCP install is not the artifact gate

`install_score` binds an **already staged file** into a runtime session. Validation must succeed **before** MCP installation during real runs; MCP must never be treated as the proof that a proposal is acceptable Python + IC policy.

## Canonical callable

Generated IC policies must expose **`score_fn(node, graph, depth) -> float`**. The manifest symbol forwarded to MCP `install_score` matches this contract (`internal/agents/optimizer/policy/canonical.go`, `selectionPolicyV1DefaultSymbol` usages).

## Default pipeline steps (IC)

For IC-targeted proposals, `internal/agents/optimizer/policy` stages a workspace and runs, in order:

1. **`stage_policy`** — writes `policy.json` metadata (ids, interface, symbol, sha256) under the staged workspace (Go-only step result).
2. **`policy_static_precheck`** — `python -m py_compile` on the staged module (fast syntax gate only).
3. **`ic_validate_policy`** — `uv run python -m iterative_context.validate_policy --policy-path … --policy-id … --symbol score_fn --json` from the **`iterative-context`** submodule root.
4. **`basedpyright`** — `uv run basedpyright`
5. **`ruff_check`** — `uv run ruff check`
6. **`pytest`** — `uv run pytest`

Commands are executed without `sh -c`, without raw `pytest`/`pyright` binaries on PATH, and without `PYTHONPATH` hacks. Dynamic argv is constrained by `execpipeline.ICOptimizerAllowlist`.

## Discovering the submodule root

By default the validator searches upward from the process working directory for `iterative-context/pyproject.toml` or `src/iterative-context/pyproject.toml`.

Override explicitly with:

`SEARCHBENCH_ITERATIVE_CONTEXT_ROOT=/absolute/path/to/iterative-context`

## Tests vs production

Round package tests set `Input.OptimizerValidateProposal` to a lightweight stub (`stubOptimizerPipelinePass`) so CI stays fast and does not require `uv`/full IC gates on every test run. Production paths leave this field **nil**, which selects `optimizepolicy.Validate` (full IC pipeline).

To exercise the full pipeline locally, run an optimizer-backed manifest from the repo root with `uv` installed and leave `OptimizerValidateProposal` unset (CLI / integration harness).

## Related modules

- IC CLI: `iterative-context` → `iterative_context.validate_policy`
- Go routing: `internal/agents/optimizer/policy/validate.go`
- Allowlist: `internal/adapters/pipeline/exec/ic_allowlist.go`
