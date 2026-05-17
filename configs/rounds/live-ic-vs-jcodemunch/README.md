# Live IC vs jCodeMunch

Repo-owned live round comparing Iterative Context (challenger) and jCodeMunch (incumbent) on LCA code localization.

**Manifest:** `round.pkl`
**Published bundle:** `artifacts/games/code-localization/rounds/live-ic-vs-jcodemunch-001/`

Buck is the only supported public interface. After any target, inspect **`report.json`** in the bundle directory first.

## Buck targets

| Target | Command | Network | Secrets (`.env`) |
| --- | --- | --- | --- |
| `validate` | `buck2 test` | No | — |
| `validate_bundle` | `buck2 test` | No | — |
| `materialize_dataset` | `buck2 run` | Optional HF | `HF_TOKEN` optional |
| `run` | `buck2 run` | Depends | `CEREBRAS_API_KEY` if live manifest |
| `live_smoke` | `buck2 test` | Yes | `CEREBRAS_API_KEY` |
| `e2e` | `buck2 test` | Yes | alias for `live_smoke` |
| `evaluate_n` | `buck2 run` | Yes | `CEREBRAS_API_KEY` |
| `stability_probe` | `buck2 run` | Yes | `CEREBRAS_API_KEY` |

Prefix: `//configs/rounds/live-ic-vs-jcodemunch:`

Example:

```bash
buck2 test //configs/rounds/live-ic-vs-jcodemunch:validate_bundle
buck2 test //configs/rounds/live-ic-vs-jcodemunch:live_smoke
buck2 run  //configs/rounds/live-ic-vs-jcodemunch:evaluate_n
```

## Secrets

Repo-root `.env` — secrets only:

- `CEREBRAS_API_KEY` — live evaluator
- `HF_TOKEN` — optional Hugging Face export for `materialize_dataset`

MCP launchers, manifest paths, and artifact roots are resolved by Buck/Go defaults, not `.env`.

## Modes

| Target | Purpose |
| --- | --- |
| `validate_bundle` | Deterministic replay of checked-in bundle |
| `live_smoke` | One fresh live integration run (not a benchmark) |
| `evaluate_n` | N fresh attempts → consolidated `report.json` + `attempts/` |
| `stability_probe` | Same-input variance; no promotion |

See [docs/reference/live-e2e.md](../../../docs/reference/live-e2e.md) and [docs/reference/run-entrypoints.md](../../../docs/reference/run-entrypoints.md).

## Dataset

Materialize before live runs if JSONL is missing:

```bash
buck2 run //configs/rounds/live-ic-vs-jcodemunch:materialize_dataset
```

See [datasets/README.md](datasets/README.md).
