# LCA dataset slice (generated)

Live rounds read `JetBrains-Research/lca-bug-localization` from JSONL under this tree.

**Do not commit `*.jsonl` here.** Materialize before a live run:

```bash
buck2 run //configs/rounds/live-ic-vs-jcodemunch:materialize_dataset
```

Optional: set `HF_TOKEN` in repo-root `.env` for Hugging Face access.
