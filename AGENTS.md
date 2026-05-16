# AGENTS.md

Operational contract for contributors and coding agents working in SearchBench-Go.

## Product vocabulary

Use everywhere: **Game → Round → Match → Evidence → Decision → NextChallenger**. Prefer **IncumbentPolicy**, **ChallengerPolicy**, **Match** over legacy “task” unless at an explicit external dataset boundary (comment those exceptions).

## Start here (docs)

Read in order:

1. [docs/index.md](docs/index.md) — docs index
2. [docs/start-here.md](docs/start-here.md) — one-page orientation
3. [docs/components.md](docs/components.md) — monorepo source component map
4. [docs/architecture.md](docs/architecture.md) — Go layers and boundaries
5. [docs/development.md](docs/development.md) — validation commands

Deeper material lives under [docs/reference/](docs/reference/), not in this file.

## Package boundaries

| Layer | Path | Rule |
| --- | --- | --- |
| Pure model | `src/searchbench-go/internal/pure` | No agent prompts, Eino, or adapters |
| Ports | `src/searchbench-go/internal/ports` | Shared contracts only |
| App | `src/searchbench-go/internal/app` | Round orchestration |
| Agents | `src/searchbench-go/internal/agents` | Evaluator/optimizer; **must not import `app`** |
| Adapters | `src/searchbench-go/internal/adapters` | Pkl, bundles, pipelines |
| Surface | `src/searchbench-go/internal/surface` | CLI |

Enforced by `src/searchbench-go/internal/architecture/imports_test.go`. Details: [docs/reference/package-boundaries.md](docs/reference/package-boundaries.md).

## Validation (required before handoff)

```bash
nix develop
buck2 test //:check_full
```

Or **`git commit`** (`buck2 test //:check`) and **`git push`** (`buck2 test //:check_full`) after `nix develop`.

Targeted checks: `buck2 test //src/searchbench-go:check`, `buck2 test //src/iterative-context:check_full`, `buck2 test //docs:check`.

Pkl schema change: `buck2 build //src/searchbench-go:pkl_go_types` then `buck2 test //src/searchbench-go:pkl_go_types_check`.

Prefer **Buck targets** over raw commands; see [docs/development.md](docs/development.md). Raw `go test`, `npm`, and `pkl` are debugging fallbacks only.

## Non-goals (unless the task explicitly asks)

- Live MCP, LangSmith, provider execution, dataset materialization, visualization UI
- Buck as a requirement for public `local_path` round runs
- Rewriting long research docs when a spine doc update suffices

## Current examples

| What | Path |
| --- | --- |
| Round manifest | `configs/rounds/local-ic-vs-jcodemunch/round.pkl` |
| Continuation round | `configs/rounds/optimize-ic/round.pkl` |
| Objective | `configs/rounds/local-ic-vs-jcodemunch/scoring/localization-objective.pkl` |
| Example bundle | `configs/rounds/local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001/` |
| Schema | `configs/schema/SearchBenchRound.pkl` |
| Bundle docs | [docs/reference/bundles.md](docs/reference/bundles.md) |

## Key code paths (when implementing)

- Round lifecycle: `src/searchbench-go/internal/app/round`
- Schema: `configs/schema/SearchBenchRound.pkl`
- IC validation: `src/searchbench-go/internal/agents/optimizer/policy/`, [docs/candidate-workspaces.md](docs/candidate-workspaces.md)
- Buck gates: root `BUCK` — `//:check`, `//:check_full`

External meta-harness owns worktrees and merge orchestration outside this repo.
