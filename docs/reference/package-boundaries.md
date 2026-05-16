# Package boundaries

Enforced by `src/searchbench-go/internal/architecture/imports_test.go`.

```text
internal/
  pure/       deterministic model — no agents, adapters, Eino
  ports/      shared contracts
  app/        round orchestration
  agents/     evaluator + optimizer — must NOT import app
  adapters/   Pkl, bundles, pipelines, scoring execution
  surface/    CLI
  games/      game wiring
```

```text
surface / app / adapters / agents  →  ports  →  pure
```

## `internal/pure`

Game/round/match vocabulary, execution records, reports, scores, optimizer records, codegraph. No prompts, Pkl runtime, Eino, MCP, or provider SDKs.

## `internal/agents`

Evaluator and optimizer vertical slices (prompt, Eino, policy). Colocated with orchestration but **cannot** import `internal/app` or `internal/surface`.

## `internal/app`

`internal/app/round` composes pure, ports, adapters, and agents.

## `internal/adapters`

Filesystem bundles, Pkl config load, subprocess pipelines, Pkl scoring. No agent prompt logic.

## `internal/surface`

CLI only; delegates to `app`.

## Tests

`internal/testing` — harness helpers; not imported by production packages except tests.
