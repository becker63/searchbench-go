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

The routine gate is Git-driven: use `nix develop`, then rely on **`git commit`** and **`git push`** to run hooks.

For schema changes, regenerate Pkl bindings with:

```sh
pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl --output-path=. configs/schema/SearchBenchRound.pkl
```

## Nix development (preferred)

Use the flake for a reproducible toolchain, pre-commit hooks, and `searchbench-*` helpers (defined under `nix/tools/`, no ad hoc `scripts/` for project automation).

**`nix develop`** installs Git hooks from git-hooks.nix. The generated `.pre-commit-config.yaml` is gitignored.

| Stage | What runs |
| --- | --- |
| **`git commit` (pre-commit)** | Fast repo-local checks: formatting (Go/Nix/shell), hygiene, **golangci-lint** (includes **staticcheck** via `.golangci.yml`), **govet**, architecture + prompt contract tests, Pkl/templ generated-file checks, **Repomix** snapshot (`repomix-output.xml` regenerated and staged) |
| **`git push` (pre-push)** | **`go test ./...`**, root **e2e**, **searchbench-check-generated**, **go mod tidy** check, **standalone staticcheck** (`searchbench-staticcheck`), **standalone golangci-lint**, **`nix flake check`** |

Hook staging avoids duplicate **staticcheck** on the same stage: pre-commit uses **golangci-lint** with `staticcheck` enabled in `.golangci.yml`. Pre-push runs **explicit** `searchbench-staticcheck` and `searchbench-golangci` as a fuller proof pass. Manual `nix develop -c searchbench-staticcheck` / `searchbench-golangci` are for reproducing hook failures only — not a separate “daily routine” tier.

| Command | Purpose |
| --- | --- |
| `nix develop` | Dev shell: Go, Pkl, golangci-lint, hooks, `searchbench-*` tools |
| `nix develop -c pre-commit run --all-files` | Full dev hook run (same family as `git commit`) |
| `nix flake check` | Sandboxed, **non-mutating** checks — formatting / Nix / shell / lightweight gates; **no** full Go module graph over the network in the default sandbox |
| `nix develop -c searchbench-update-repomix` | Regenerate `repomix-output.xml` and `git add` it (normally the pre-commit Repomix hook handles this) |

**Repomix:** This repository intentionally commits `repomix-output.xml` so the current tree can be shared quickly with AI assistants for architectural review.

**Go dependencies:** There is no checked-in `vendor/` tree. Hooks that load the full module graph run in `nix develop` and on **pre-commit** / **pre-push**, not inside the default **`nix flake check`** sandbox (no network there).

**`staticcheck` binary:** On `PATH` via `nixpkgs` `go-tools`. Prefer **`nix develop -c searchbench-staticcheck`** when debugging a staticcheck failure.

**Go / lint policy:** `.golangci.yml` enables high-signal checks (`govet`, `staticcheck`, `ineffassign`, `unused`, `errcheck`, `copyloopvar`, `unconvert`) — not broad style linters.

**Orchestration outside this repo:** Worktrees, branch lifecycle, task assignment, agent summary packs, and merge orchestration are owned by an **external meta harness**, not by SearchBench-Go. This repository owns **repo-local Git hooks** and **debug-only** `searchbench-*` commands listed below — it does not provide `searchbench-agent-*` tooling, worktree creation, or `AGENT_TASK.md` / `AGENT_REVIEW.md` generation.

**Debugging commands** (use only when a hook failed and you need the same check ad hoc):

| Command | Purpose |
| --- | --- |
| `nix develop -c searchbench-staticcheck` | `staticcheck ./...` |
| `nix develop -c searchbench-golangci` | `golangci-lint run ./...` (uses `.golangci.yml`) |
| `nix develop -c searchbench-go-mod-tidy-check` | Fail if `go mod tidy` would change `go.mod` / `go.sum` |
| `nix develop -c searchbench-prompt-contract-check` | Tests for `.templ` XML prompt contracts |
| `nix develop -c searchbench-refresh-pkl-example-fixtures` | Regenerate optimize-IC fixtures from the local round (long-running) |
| `nix develop -c searchbench-go-build-root` | `go build -o searchbench ./cmd/searchbench` |
| `nix develop -c searchbench-architecture-check` | Import-boundary tests (`go test` under `internal/architecture`) |
| `nix develop -c searchbench-check-generated` | Pkl + templ + combined generated outputs |
| `nix develop -c searchbench-check-pkl-generated` | Pkl bindings vs schema |
| `nix develop -c searchbench-check-templ-generated` | Templ-generated prompts |
| `nix develop -c searchbench-e2e` | Root package integration tests |
| `nix develop -c searchbench-go-test-all` | `go test ./...` |
| `nix develop -c searchbench-nix-flake-check` | `nix flake check` (same as pre-push sandbox step) |
