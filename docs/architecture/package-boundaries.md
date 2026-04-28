# Package Boundaries

SearchBench-Go now makes its architecture explicit in the `internal/` tree:

```text
internal/
  pure/
  generic/
  ports/
  app/
  adapters/
  surface/
  testing/
```

The intended dependency direction is:

```text
surface / app / adapters
    ↓
ports
    ↓
pure / generic
```

`pure` and `generic` must not depend outward.

## `internal/pure/`

SearchBench-specific deterministic model code.

Belongs here:
- domain vocabulary
- run records and failures
- score models, evidence, and objective results
- report models and report-to-score projection
- pure codegraph models
- pure prompt input and rendering
- harness-owned usage accounting

Does not belong here:
- Pkl runtime execution
- Eino
- provider SDKs
- MCP
- tracing SDKs
- `os/exec`
- filesystem bundle writing
- CLI presentation

Current packages:
- `internal/pure/domain`
- `internal/pure/run`
- `internal/pure/score`
- `internal/pure/report`
- `internal/pure/codegraph`
- `internal/pure/prompts`
- `internal/pure/usage`

## `internal/generic/`

Effect-free helpers that are not SearchBench-specific.

Use this layer only for code that could plausibly be copied into another repo
without bringing SearchBench nouns with it. Do not create a junk-drawer
`common` or `utils` package.

This layer is intentionally empty right now. Add packages here only when they
are genuinely generic.

## `internal/ports/`

Project-owned contracts for effectful systems.

Belongs here:
- backend/session interfaces
- pipeline step/result/classification contracts
- future dataset/materialize/trace/model ports

Does not belong here:
- concrete SDK bindings
- filesystem or subprocess implementations
- CLI presentation

Current packages:
- `internal/ports/backend`
- `internal/ports/pipeline`

## `internal/app/`

Use-case orchestration and composition.

Belongs here:
- comparison orchestration
- local fake end-to-end composition
- cross-cutting app logging used by orchestrators and surfaces

Does not belong here:
- core domain model ownership
- concrete provider or filesystem adapters

Current packages:
- `internal/app/compare`
- `internal/app/localrun`
- `internal/app/logging`

## `internal/adapters/`

Concrete world-touching or runtime-binding implementations.

Belongs here:
- filesystem bundle writing
- Pkl config loading
- Pkl objective scoring execution
- Eino execution
- subprocess pipeline execution

Future adapters should land here:
- `internal/adapters/materialize/worktree`
- `internal/adapters/dataset/huggingface`
- `internal/adapters/dataset/lca`
- `internal/adapters/trace/langsmith`
- `internal/adapters/backend/iterativecontext`
- `internal/adapters/backend/jcodemunch`
- `internal/adapters/providers/openai`
- `internal/adapters/providers/openrouter`
- `internal/adapters/providers/cerebras`

Current packages:
- `internal/adapters/artifact/fsbundle`
- `internal/adapters/config/pkl`
- `internal/adapters/scoring/pkl`
- `internal/adapters/executor/eino`
- `internal/adapters/pipeline/exec`

## `internal/surface/`

User- and developer-facing entrypoints and presentation.

Belongs here:
- CLI routing
- console rendering

Current packages:
- `internal/surface/cli`
- `internal/surface/console`

## `internal/testing/`

Test-only infrastructure and fake worlds.

Belongs here:
- scripted models
- fake OpenAI-compatible servers
- reusable test fixtures

Production code must not import `internal/testing/...`.

Current packages:
- `internal/testing/modeltest`

## Import Rules

- `internal/pure/...` must not import `internal/adapters/...`, `internal/surface/...`, `internal/testing/...`, `github.com/apple/pkl-go`, Eino, provider SDKs, tracing SDKs, or `os/exec`.
- `internal/generic/...` must remain SearchBench-agnostic and must not import any `internal/...` package.
- `internal/ports/...` must not import `internal/adapters/...`, `internal/surface/...`, or `internal/testing/...`.
- Production packages must not import `internal/testing/...`.

These rules are enforced by `internal/architecture/imports_test.go`.
