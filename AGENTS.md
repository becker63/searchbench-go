# AGENTS.md

Operational contract for contributors and coding agents working in SearchBench-Go.

## Product vocabulary

Use everywhere: **Game → Round → Match → Evidence → Decision → NextChallenger**. Prefer **IncumbentPolicy**, **ChallengerPolicy**, **Match** over legacy “task” unless at an explicit external dataset boundary (comment those exceptions).

## Start here (docs)

Read in order:

1. [docs/index.md](docs/index.md) — docs index
2. [docs/start-here.md](docs/start-here.md) — one-page orientation
3. [docs/architecture.md](docs/architecture.md) — layers and boundaries
4. [docs/development.md](docs/development.md) — validation commands

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
nix develop   # once per shell
cd src/searchbench-go && go test ./...
# from repo root:
nix develop -c buck2 test //:check_full
```

Or rely on **`git commit`** (`//:check`) and **`git push`** (`//:check_full`, includes `//docs:check`) after `nix develop`.

Docs site: `npm run docs:dev` / `npm run docs:build`; `buck2 test //docs:check`. GitHub Pages deploys from `main` (not from local hooks).

Pkl schema change — regenerate bindings:

```bash
cd src/searchbench-go
pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl \
  --output-path=. ../../configs/schema/SearchBenchRound.pkl
```

## Non-goals (unless the task explicitly asks)

- Live MCP, LangSmith, provider execution, dataset materialization, visualization UI
- Buck as a requirement for public `local_path` round runs
- Rewriting long research docs when a spine doc update suffices

## Key code paths (when implementing)

- Round lifecycle: `src/searchbench-go/internal/app/round`
- Schema: `configs/schema/SearchBenchRound.pkl`
- IC validation: `src/searchbench-go/internal/agents/optimizer/policy/`, [docs/workspace-seeds.md](docs/workspace-seeds.md)
- Buck gates: root `BUCK` — `//:check`, `//:check_full`

External meta-harness owns worktrees and merge orchestration outside this repo.
