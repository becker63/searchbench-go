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

## Nix development (preferred)

Use the flake for a reproducible toolchain, pre-commit hooks, and the `searchbench-*` helper commands (defined in `nix/dev-tools.nix`, no ad hoc `scripts/` for project automation).

| Command | Purpose |
| --- | --- |
| `nix develop` | Dev shell: Go, Pkl, golangci-lint, hooks, `searchbench-*` tools |
| `nix develop -c pre-commit run --all-files` | Full local hook run (includes Repomix refresh in the dev hook set) |
| `nix flake check` | Sandboxed, non-mutating checks (close to CI); uses `nix/vendor` (via root `vendor` symlink) for offline Go |
| `nix develop -c searchbench-update-repomix` | Regenerate `repomix-output.xml` and `git add` it |

The file `.pre-commit-config.yaml` is generated when you enter `nix develop` and is gitignored.

**Repomix:** This repository intentionally commits `repomix-output.xml` so the current tree can be shared quickly with AI assistants for architectural review. That is intentional workflow hygiene for this project, not a general recommendation for every repo.

**Go vendor directory:** Module sources are stored under [`nix/vendor`](nix/vendor); the repo root [`vendor`](vendor) is a symlink so `go mod vendor` and `-mod=vendor` keep working. Content is checked in so `nix flake check` can run Go hooks in a sandbox without network access. After dependency changes, run `go mod vendor` and commit the result.

**Staticcheck:** The standalone `staticcheck` pre-commit hook is not enabled yet because the current tree still has a backlog of `SA*`/`U1000` findings (mostly in tests and generated-adjacent helpers). `gofmt`, `govet`, and `golangci-lint` (minimal config: `govet` only) run in CI; run `staticcheck ./...` locally when you are cleaning that backlog.
