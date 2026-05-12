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

Use the flake for a reproducible toolchain, pre-commit hooks, and the `searchbench-*` helper commands (defined under `nix/tools/`, no ad hoc `scripts/` for project automation).

| Command | Purpose |
| --- | --- |
| `nix develop` | Dev shell: Go, Pkl, golangci-lint, hooks, `searchbench-*` tools |
| `nix develop -c pre-commit run --all-files` | Full hook run: Go linters, `go test` hooks, Repomix in the dev hook set, etc. |
| `nix flake check` | Sandboxed, non-mutating checks (no network) — Nix/shell/formatting; not full Go analysis |
| `nix develop -c searchbench-update-repomix` | Regenerate `repomix-output.xml` and `git add` it |

The file `.pre-commit-config.yaml` is generated when you enter `nix develop` and is gitignored.

**Repomix:** This repository intentionally commits `repomix-output.xml` so the current tree can be shared quickly with AI assistants for architectural review. That is intentional workflow hygiene for this project, not a general recommendation for every repo.

**Go dependencies:** There is no checked-in `vendor/` tree. The module cache is populated from the network (or your local cache) when you build and test. **Hooks that load the full module graph** (`govet`, `staticcheck`, `golangci-lint`, `searchbench-architecture`, `searchbench-prompt-contract`, pre-push Go tests) run in `nix develop` and on pre-push — not inside `nix flake check`, because that runs in a sandbox without internet ([git-hooks.nix](https://github.com/cachix/git-hooks.nix) documents this). You can run `go mod vendor` locally if you want a `./vendor` directory; it is gitignored.

**`staticcheck` binary:** Provided on `PATH` via `nixpkgs` `go-tools` (same family as the git-hooks `staticcheck` integration). Run `nix develop -c staticcheck ./...` or `nix develop -c searchbench-staticcheck`.

**Go / lint policy:** `.golangci.yml` enables high-signal checks (`govet`, `staticcheck`, `ineffassign`, `unused`, `errcheck`, `copyloopvar`, `unconvert`) — not broad style linters.

**Quality gate tiers:**

| Tier | What runs |
| --- | --- |
| `nix flake check` | Sandboxed: `gofmt`, Nix (`nixfmt`, `statix`, …), shell, `searchbench-no-scripts`, vocabulary warning — **no** full Go module graph |
| `nix develop` + pre-commit | Full dev hook set: Go vet, staticcheck, golangci-lint, architecture + prompt contract tests, generated checks, Repomix refresh, etc. |
| `git push` (pre-push) | `go test ./...`, root e2e, `searchbench-check-generated`, `searchbench-go-mod-tidy-check`, `searchbench-staticcheck`, `searchbench-golangci` |
| `nix develop -c searchbench-agent-merge-check` | Strictest local gate: pre-commit, `go test`, `go test -race`, staticcheck, golangci-lint, e2e, `nix flake check`, generated checks, go mod tidy check, Repomix refresh, `git diff --check` |

**Handy commands:**

| Command | Purpose |
| --- | --- |
| `nix develop -c searchbench-staticcheck` | `staticcheck ./...` |
| `nix develop -c searchbench-golangci` | `golangci-lint run ./...` (uses `.golangci.yml`) |
| `nix develop -c searchbench-go-mod-tidy-check` | Fail if `go mod tidy` would change `go.mod` / `go.sum` |
| `nix develop -c searchbench-prompt-contract-check` | Tests for `.templ` XML prompt contracts |
| `nix develop -c searchbench-refresh-pkl-example-fixtures` | Regenerate optimize-IC fixtures from the local round (long-running) |
| `nix develop -c searchbench-openai-netwatch` | Optional HTTPS connection diagnostics helper (migrated from legacy `scripts/`) |
| `nix develop -c searchbench-go-build-root` | `go build -o searchbench ./cmd/searchbench` |
