# SearchBench run entrypoints

Repo-owned runs and evaluations use **Buck targets** only. Buck is the only supported public interface; the Go binary is private implementation plumbing invoked by Buck. Do not start runs via ad-hoc shell scripts or direct `searchbench` commands.

After any target completes, inspect **`report.json`** (then `report.txt`) in the bundle directory first.

## Live IC vs jCodeMunch (`configs/rounds/live-ic-vs-jcodemunch`)

| Target | Command | Mode | Network / secrets |
| --- | --- | --- | --- |
| `validate` | `buck2 test` | Pkl manifest validation | No |
| `validate_bundle` | `buck2 test` | Deterministic bundle replay | No |
| `materialize_dataset` | `buck2 run` | LCA HF export | Optional `HF_TOKEN` |
| `run` | `buck2 run` | Single round execution | Depends on manifest |
| `live_smoke` | `buck2 test` | Fresh live MCP smoke | Yes (`.env`) |
| `e2e` | `buck2 test` | Alias for `live_smoke` | Yes |
| `evaluate_n` | `buck2 run` | Multi-attempt promotion evaluation | Yes |
| `stability_probe` | `buck2 run` | Same-input variance probe | Yes |

Full labels: `//configs/rounds/live-ic-vs-jcodemunch:<target>`.

Secrets: repo-root [`.env`](../../.env) — **`CEREBRAS_API_KEY`** and optional **`HF_TOKEN`** only. MCP launchers and paths come from Pkl/Go/Buck defaults, not `.env`.

Round README: [configs/rounds/live-ic-vs-jcodemunch/README.md](../../configs/rounds/live-ic-vs-jcodemunch/README.md).

## Deprecated / removed

- `scripts/run-live-e2e.sh`
- `src/searchbench-go/live_e2e.sh`
- `buck2 test //src/searchbench-go:live_e2e`
- `tooling/lca_hf_export.{py,sh}`
- `go test -tags=live_e2e` as a product run path
- Direct `searchbench run` / `searchbench round run` as repo-owned workflows

## Implementation details (not public)

Buck targets are declared as Starlark rules/macros (`searchbench_round_op`, `go_external_package_test`, `python_module_test`, etc.) that invoke toolchain executables directly. Round operations use the `go_binary` at `//src/searchbench-go/cmd/searchbench:searchbench`. There are no repo-owned `.sh` operation entrypoints.

The Go binary and its flags are not a stable user-facing API. Raw `go test`, `go build`, and `./searchbench` remain debugging fallbacks for harness developers only.
