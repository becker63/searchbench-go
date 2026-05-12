# AGENTS.md

SearchBench-Go is a review-game system for evaluating AI policy changes across datasets. The public product model is:

```text
Game -> Round -> Match -> Evidence -> Decision -> NextChallenger
```

A round compares an `IncumbentPolicy` and a `ChallengerPolicy` over the same match slice, produces durable evidence, records a decision, and may generate a next challenger proposal. Prefer this vocabulary in code, docs, tests, logs, CLI output, schemas, and artifact names.

## Start Here

Read these files first when working in the repo:

- `docs/README.md` (documentation index)
- `docs/architecture/architecture.md`
- `docs/architecture/visualization.md`
- `docs/architecture/integration-shape.md`
- `docs/architecture/package-boundaries.md`
- `docs/engineering/agentic-development-flow.md`
- `configs/schema/SearchBenchRound.pkl`
- `internal/app/round`
- `internal/agents/evaluator`
- `internal/agents/optimizer`
- `internal/pure/report`
- `internal/pure/score`
- `internal/pure/optimizer`
- `internal/adapters/bundle/fs`
- `internal/adapters/config/pkl`

## Boundaries

Keep deterministic SearchBench model code in `internal/pure`. Keep round lifecycle orchestration in `internal/app`. Colocate evaluator- and optimizer-specific behavior under `internal/agents` (prompt + Eino + agent-local persistence/fakes). Keep shared filesystem, Pkl, pipeline, and other non-agent integrations in `internal/adapters`. Keep CLI and terminal rendering in `internal/surface`.

Do not add real MCP, LangSmith, provider execution, dataset materialization, or visualization UI unless the current task explicitly asks for it.

## Naming Rules

Use `Game`, `Round`, `Match`, `IncumbentPolicy`, `ChallengerPolicy`, `Evidence`, `Decision`, and `NextChallenger` for active architecture. Old terms are acceptable only in historical prompts/docs or at explicit external dataset boundaries where upstream data uses that term. Add a short comment for intentional external-boundary exceptions.

## Validation

Run `go test ./...` before handing off code changes. For schema changes, regenerate Pkl bindings with:

```sh
pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl --output-path=. configs/schema/SearchBenchRound.pkl
```
