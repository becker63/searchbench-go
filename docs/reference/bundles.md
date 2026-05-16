# Bundles

A **bundle** is the durable artifact tree for one completed (or in-progress) round.

## Layout (typical)

Under `--bundle-root`, paths are organized by game and round id, for example:

```text
{bundle-root}/games/{game}/rounds/{round-id}/
  manifest / resolved inputs
  evidence.pkl
  objective.json
  reports and match artifacts
  decision record
```

Exact filenames depend on the game and round version; see a local run under `.tmp-artifacts` or `configs/rounds/*/artifacts/` examples.

## Contents

Bundles record what is needed to **review** a round without replaying live services:

- resolved manifest inputs
- per-match outputs and usage
- round evidence and objective result
- **Decision** (`PROMOTE` | `REVIEW` | `REJECT`)
- workspace seed / candidate identity when applicable

## Code

- Write path: `internal/adapters/bundle/fs`
- Models: `internal/pure/report`, `internal/pure/score`

Immutable after completion — new rounds get new bundle directories.
