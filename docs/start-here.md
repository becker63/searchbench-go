# Start here

SearchBench tests **agent environments**: which tools, search interfaces, and configs help the **same agent** perform better on a fixed benchmark slice.

**Work in progress** — research harness, not a polished platform. **First game:** code localization (bug-localization slices, symbol/code-search with lookahead).

## What you can run today

From the repo root after building the CLI ([README § Run one local round](../README.md#run-one-local-round)):

```bash
./searchbench run \
  --manifest=configs/rounds/local-ic-vs-jcodemunch/round.pkl \
  --bundle-root="$(pwd)/.tmp-artifacts"
```

Offline **fake-local** path (no live MCP or models). Contributor setup: [development.md](./development.md).

## How a round works

```text
Pkl manifest → matches (evaluator) → evidence → objective → decision → bundle
```

Optional optimizer proposes a **NextChallenger** for a later round.

## Read next

1. [concepts.md](./concepts.md) — Game, Interface, Round, bundle, …
2. [architecture.md](./architecture.md) — `pure` / `app` / `agents` / `adapters`
3. [development.md](./development.md) — `nix develop`, `buck2 test //:check_full`
4. [workspace-seeds.md](./workspace-seeds.md) — IC candidate workspaces

Docs home: [index.md](./index.md). Contributors: [AGENTS.md](../AGENTS.md).
