# SearchBench-Go Implementation Roadmap

**Docs hub:** [Documentation index](../README.md) · [Architecture](../architecture.md)

This roadmap is intentionally high-level. The current source of truth for package boundaries is [package boundaries](../architecture/package-boundaries.md); this file tracks future implementation pressure.

## Current Model

SearchBench-Go uses:

```text
Game -> Round -> Match -> Evidence -> Decision -> NextChallenger
```

The stable deterministic model lives under `internal/pure`. Prompt templates and agent-specific Eino runners live under `internal/agents`. Round orchestration stays in `internal/app/round`. Shared runtime integrations (filesystem helpers, config loading, Pkl scoring adapters, pipelines) remain in `internal/adapters`. CLI and terminal presentation remain in `internal/surface`.

## Near-Term Work

### Dataset Adapters

Add a narrow adapter for external datasets, starting with LCA if needed.

Rules:

- Project external rows into `domain.MatchSpec` before they cross inward.
- Keep upstream dataset vocabulary at the adapter boundary.
- Preserve prompt-visible input and scorer-only oracle separation.

### Real Backend Execution

Add concrete backend/provider execution only when the fake/local path is no longer enough.

Rules:

- Keep provider SDK types out of `internal/pure`.
- Convert runtime observations into `execution.ExecutedRun` and `execution.RunFailure`.
- Keep model/tool loop details inside adapters or executor-specific packages.

### Decision Rules

Move richer decision policy behind `compare.Decider` when simple local rules become insufficient.

Rules:

- Produce `report.Decision`.
- Consider objective result, regressions, failures, cost, and token efficiency.
- Do not put a large rule engine into `internal/app/compare`.

### Next-Challenger Feedback

Use `RoundReport`, `RoundEvidenceDocument`, and `ObjectiveResult` as the primary input to next-challenger generation.

Rules:

- Do not use raw traces as the primary interface.
- Do not recompute scoring in the feedback layer.
- Validate output as a next-challenger policy artifact.

### Round Bundle Persistence

Keep the canonical bundle shape:

```text
artifacts/games/<game-id>/rounds/<round-id>/
  COMPLETE
  resolved-round.json
  round-report.json
  round-report.txt
  evidence.pkl
  objective.json
  decision.json
  metadata.json
```

Rules:

- Reports remain JSON-serializable `report.RoundReport`.
- Filesystem behavior stays out of `internal/pure/report`.
- Optimizer bundles may add next-challenger-specific artifacts separately.

### Telemetry

Tracing and review-platform integrations can consume SearchBench artifacts, but must not define them.

Rules:

- No tracing SDK types in `internal/pure`.
- No telemetry-driven control flow.
- Prefer explicit artifact links and metadata over hidden state.

## CLI Direction

Keep the command surface small and round-centered:

```text
searchbench round run --manifest configs/rounds/local-ic-vs-jcodemunch/round.pkl
```

Future report inspection commands should consume round bundles or `round-report.json`, not invent a second report model.

## Guiding Rule

Translate legacy implementation concerns into the Game/Round/Match model instead of recreating old module shapes.
