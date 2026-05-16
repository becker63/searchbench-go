# Package Boundaries

**Related docs:** [Architecture spine](./architecture.md) · [Visualization](./visualization.md) · [Integration shape](./integration-shape.md) · [Engineering workflow](../engineering/agentic-development-flow.md)

SearchBench-Go now makes its architecture explicit in the `internal/` tree:

```text
internal/
  pure/
  generic/
  ports/
  app/
  agents/
  adapters/
  surface/
  testing/
```

The intended dependency direction is:

```text
surface / app / adapters / agents (agent-specific)
    ↓
ports
    ↓
pure / generic
```

Agents import `pure` for typed contracts but must not depend on `internal/app` or `internal/surface`. `pure` and `generic` must not depend outward (including agents and adapters).

## `internal/pure/`

SearchBench-specific deterministic model code.

Belongs here:
- game/round/match/policy vocabulary (where present)
- execution records and failures
- score models, round evidence, and objective results
- round report models and report-to-evidence construction
- pure codegraph models
- optimizer record types (`internal/pure/optimizer`)
- harness-owned usage accounting

Does not belong here:
- Agent prompt rendering (templates, Markdown assembly, Eino-specific prompt plumbing)
- Pkl runtime execution
- Eino
- provider SDKs
- MCP
- tracing SDKs
- `os/exec`
- filesystem bundle writing
- CLI presentation

Prompt *contracts* backed by deterministic pure types now render from `internal/agents/*/prompt` so `pure` never imports Eino.

Current packages include:
- `internal/pure/game`
- `internal/pure/round`
- `internal/pure/domain`
- `internal/pure/execution`
- `internal/pure/score`
- `internal/pure/report`
- `internal/pure/optimizer`
- `internal/pure/codegraph`
- `internal/pure/usage`

## `internal/agents/`

Vertical slices for the evaluator and optimizer (NextChallenger) agents.

Belongs here:
- evaluator/optimizer prompts (`internal/agents/*/prompt`)
- evaluator/optimizer Eino runners (`internal/agents/*/eino`)
- evaluator-local callbacks (`internal/agents/evaluator/eino/callbacks`)
- consolidated deterministic fakes backing local fake rounds (`internal/agents/evaluator/fake`)
- optimizer-only bundle persistence and Python-policy validation helpers

Does not belong here:
- Round lifecycle sequencing (`ResolveRound`, `Run`, ...) — stays in `internal/app/round`
- Shared generic filesystem adapters unrelated to optimizer proposals

Dependency rule: `internal/app/round` may import agents; agents must not import `internal/app/...`.

Current packages include:
- `internal/agents/evaluator`
- `internal/agents/evaluator/prompt`
- `internal/agents/evaluator/eino`
- `internal/agents/evaluator/eino/callbacks`
- `internal/agents/evaluator/fake`
- `internal/agents/optimizer`
- `internal/agents/optimizer/prompt`
- `internal/agents/optimizer/eino`
- `internal/agents/optimizer/bundle`
- `internal/agents/optimizer/policy`

## `internal/generic/`

Effect-free helpers that are not SearchBench-specific.

Use this layer only for code that could plausibly be copied into another repo
without bringing SearchBench nouns with it. Do not create a junk-drawer
`common` or `utils` package.

This layer is intentionally empty right now. Add packages here only when they
are genuinely generic.

## `internal/ports/`

Project-owned contracts for effectful policies.

Belongs here:
- backend/session interfaces
- pipeline step/result/classification contracts
- future dataset/materialize/trace/model ports

Does not belong here:
- concrete SDK bindings
- filesystem or subprocess implementations
- CLI presentation

Current packages:
- `internal/ports/dataset`
- `internal/ports/pipeline`

## `internal/app/`

Use-case orchestration and composition.

Belongs here:
- comparison orchestration
- canonical resolved round semantics
- manifest resolution, explicit evidence threading, and execution composition
- cross-cutting app logging used by orchestrators and surfaces

Does not belong here:
- core domain model ownership
- concrete provider or filesystem adapters

Current packages:
- `internal/app/round`

## `internal/adapters/`

Concrete world-touching or runtime-binding implementations.

Belongs here:
- filesystem round bundle helpers
- Pkl config loading
- Pkl objective scoring execution
- subprocess pipeline execution

Does not belong here:
- evaluator/optimizer-specific Eino execution (implemented under `internal/agents/*/eino`)

Future adapters should land here, for example:
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
- `internal/adapters/config/pkl`
- `internal/adapters/scoring/pkl`
- `internal/adapters/bundle/fs`
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

- `internal/pure/...` must not import `internal/adapters/...`, `internal/agents/...`, `internal/surface/...`, `internal/testing/...`, `github.com/apple/pkl-go`, Eino, provider SDKs, tracing SDKs, or `os/exec`.
- `internal/agents/...` must not import `internal/app/...`, `internal/surface/...`, or `internal/testing/...`; evaluator and optimizer subtrees must not import each other.
- `internal/surface/cli/...` must not import `internal/agents/...` (delegate through `internal/app/round`).
- `internal/generic/...` must remain SearchBench-agnostic and must not import any `internal/...` package.
- `internal/ports/...` must not import `internal/adapters/...`, `internal/surface/...`, `internal/testing/...`, or `internal/agents/...`.
- Production packages must not import `internal/testing/...`.

These rules are enforced by `internal/architecture/imports_test.go`.
