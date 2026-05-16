# LCA dataset slice (generated)

Live rounds read `JetBrains-Research/lca-bug-localization` from JSONL under this tree.

**Do not commit `*.jsonl` here.** Materialize before a live run:

```bash
./tooling/lca_hf_export.sh \
  --config py --split dev --max-items 1 \
  --output-dir configs/rounds/live-ic-vs-jcodemunch
```

The live e2e test (`//src/searchbench-go:live_e2e`) calls the same exporter automatically unless `SEARCHBENCH_SKIP_HF_EXPORT=1`.
