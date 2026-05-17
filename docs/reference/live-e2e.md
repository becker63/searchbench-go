# Live MCP evaluation (IC vs jCodeMunch)

Real end-to-end round with Cerebras evaluator, jCodeMunch incumbent, and Iterative Context challenger.

**Manifest:** `configs/rounds/live-ic-vs-jcodemunch/round.pkl`
**Artifact root:** `configs/rounds/live-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/live-ic-vs-jcodemunch-001/`
**Not in `//:check`** — live targets require secrets and network.

## Buck-only interface

Repo-owned live work uses config-local Buck targets only. See [run-entrypoints.md](./run-entrypoints.md) and [configs/rounds/live-ic-vs-jcodemunch/README.md](../../configs/rounds/live-ic-vs-jcodemunch/README.md).

```bash
# Deterministic (no network)
buck2 test //configs/rounds/live-ic-vs-jcodemunch:validate
buck2 test //configs/rounds/live-ic-vs-jcodemunch:validate_bundle

# Dataset + single run (product operations)
buck2 run //configs/rounds/live-ic-vs-jcodemunch:materialize_dataset
buck2 run //configs/rounds/live-ic-vs-jcodemunch:run

# Live modes (require repo-root .env)
buck2 test //configs/rounds/live-ic-vs-jcodemunch:live_smoke
buck2 run //configs/rounds/live-ic-vs-jcodemunch:evaluate_n
buck2 run //configs/rounds/live-ic-vs-jcodemunch:stability_probe
```

Inspect **`report.json`** first in the published bundle directory.

## Truth model

| Mode | Command | Freshness | Proves |
| --- | --- | --- | --- |
| `validate_bundle` | `buck2 test` | `archive` | A checked-in completed bundle still validates (no MCP, no model) |
| `live_smoke` | `buck2 test` | `fresh` | One fresh live run succeeded — **no archive fallback** |
| `evaluate_n` | `buck2 run` | `fresh` | N fresh attempts consolidated into top-level `report.json` |
| `stability_probe` | `buck2 run` | `fresh` | Repeated same-input attempts; variance metrics only (`decision` = no promotion) |

Deterministic replay is not live proof. Live smoke is not a benchmark. Promotion decisions belong in `evaluate_n` consolidated reports, not in smoke or single-run paths.

## Secrets

Only secrets belong in repo-root `.env`:

```bash
CEREBRAS_API_KEY=...
HF_TOKEN=...   # optional; Hugging Face dataset export
```

Do not put MCP command overrides, mode selectors, or attempt counts in `.env` for normal repo-owned runs. Non-secret defaults (manifest path, artifact root, materialize cache, MCP launchers) come from Buck targets and [`internal/pure/liveconfig`](../../src/searchbench-go/internal/pure/liveconfig/config.go).

## Workspace seed

The live challenger uses `buck_descriptor` → `//src/iterative-context:optimizable_backend` in `round.pkl` (repo-owned default). Do not use `local_path` for repo-owned live/eval configs.

## LCA materialization

```bash
buck2 run //configs/rounds/live-ic-vs-jcodemunch:materialize_dataset
```

Writes `datasets/JetBrains-Research_lca-bug-localization/py/dev.jsonl` under the round directory.

## Bundle interface

Each completed bundle includes canonical **`report.json`** and **`report.txt`** — inspect these first — plus `round-report.json`, evidence, objective, and `COMPLETE`.

For `evaluate_n`, per-attempt trees live under `attempts/`; see [bundles.md](./bundles.md).

## What live smoke proves

- LCA `py/dev` row(s) from Hugging Face
- Git materialization into `.cache/searchbench/materialized-repos`
- MCP tool loops for both backends
- Cerebras via Eino
- Config-owned artifact publish path

Fast offline manifest test: `TestLiveICVsJCodeMunchManifestResolves` in `internal/app/round/live_manifest_test.go` (harness test, not a product entrypoint).
