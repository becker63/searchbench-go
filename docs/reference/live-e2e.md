# Live MCP e2e (IC vs jCodeMunch)

Real end-to-end round with Cerebras evaluator, jCodeMunch incumbent, and Iterative Context challenger.

**Manifest:** `configs/rounds/live-ic-vs-jcodemunch/round.pkl`
**Not in default CI** — requires API keys, MCP launcher commands, and a Hugging Face download for the LCA slice.

## Environment

From repo root after `nix develop` (includes `python3` with `datasets` / `huggingface_hub`):

```bash
export SEARCHBENCH_RUN_LIVE_E2E=1
export CEREBRAS_API_KEY=...          # or src/searchbench/.env locally
export SEARCHBENCH_MATERIALIZE_CACHE_DIR="$PWD/.cache/searchbench/materialized-repos"
export SEARCHBENCH_JCODEMUNCH_COMMAND='uvx jcodemunch-mcp'
export SEARCHBENCH_ITERATIVE_CONTEXT_COMMAND='uv run --directory src/iterative-context python -m iterative_context.server'
```

Optional:

| Variable | Default | Purpose |
|----------|---------|---------|
| `SEARCHBENCH_LIVE_E2E_TIMEOUT` | `20m` | Go test / round timeout |
| `SEARCHBENCH_LCA_HF_CONFIG` | `py` | HF dataset config |
| `SEARCHBENCH_LCA_HF_SPLIT` | `dev` | HF split |
| `SEARCHBENCH_LCA_HF_MAX_ITEMS` | `1` | Rows exported from HF |
| `SEARCHBENCH_SKIP_HF_EXPORT` | — | Use existing JSONL under the live round `datasets/` tree |
| `SEARCHBENCH_LCA_HF_SKIP` | `50` | Skip N streaming HF rows before export (avoids huge default repos) |
| `SEARCHBENCH_LIVE_USE_HF_ROW` | `1` | Export HF slice (default); set `0` for local monorepo row |
| `SEARCHBENCH_LIVE_E2E_VERIFY_ARCHIVE_ONLY` | — | Validate `.cache/searchbench/live-e2e-bundle-latest` only |

Successful bundles are archived to `.cache/searchbench/live-e2e-bundle-latest/` (includes `round-report.json`, `COMPLETE`, …).

## LCA slice from Hugging Face

The live manifest uses `game.lca("py", "dev", 1)`. Rows are **not** checked in; they are exported from [JetBrains-Research/lca-bug-localization](https://huggingface.co/datasets/JetBrains-Research/lca-bug-localization) before the round:

```bash
./tooling/lca_hf_export.sh \
  --config py --split dev --max-items 1 \
  --output-dir configs/rounds/live-ic-vs-jcodemunch
```

This writes `configs/rounds/live-ic-vs-jcodemunch/datasets/JetBrains-Research_lca-bug-localization/py/dev.jsonl` (gitignored). The live e2e test runs the same exporter automatically unless `SEARCHBENCH_SKIP_HF_EXPORT=1`.

Repos are then materialized at each row’s `base_sha` into `SEARCHBENCH_MATERIALIZE_CACHE_DIR` via git (not HF repo zips).

## Run

**Buck (opt-in):**

```bash
buck2 test //src/searchbench-go:live_e2e
```

**CLI** (export the slice first, then):

```bash
./tooling/lca_hf_export.sh --config py --split dev --max-items 1 \
  --output-dir configs/rounds/live-ic-vs-jcodemunch

cd src/searchbench-go && go build -o ../../searchbench ./cmd/searchbench
./searchbench run \
  --manifest=configs/rounds/live-ic-vs-jcodemunch/round.pkl \
  --bundle-root=.searchbench/artifacts
```

## What it proves

- Real LCA `py/dev` row(s) from Hugging Face (not a hand-written smoke JSONL)
- Local git materialization into `SEARCHBENCH_MATERIALIZE_CACHE_DIR` (required for MCP backends)
- Real MCP tool loops for both backends
- Cerebras model via Eino (`provider = "cerebras"`)
- Complete bundle (`evidence.pkl`, `objective.json`, `decision.json`, `COMPLETE`, …)

Fast manifest resolver test (no secrets, stub JSONL in a temp round copy): `TestLiveICVsJCodeMunchManifestResolves` in `internal/app/round/live_manifest_test.go`.

## Reference only

`src/old-searchbench/` is temporary Python reference material — not a runtime dependency. See [candidate-workspaces.md](../candidate-workspaces.md) for workspace sandboxing.
