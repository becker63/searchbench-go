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
- `docs/architecture/build-system.md`
- `BUCK` (root Buck targets: `//:check`, `//:check_full`, `//:repomix_fresh_check`)
- `src/searchbench-go/BUCK` (Go `check` suite; opt-in `pkl_go_types`)
- `src/iterative-context/BUCK` (Iterative Context `check` / `check_full` Buck suites)
- `repomix_fresh_check.sh` (Repomix freshness gate used by `//:check_full`)
- `configs/schema/SearchBenchRound.pkl`
- `src/searchbench-go/internal/app/round`
- `src/searchbench-go/internal/agents/evaluator`
- `src/searchbench-go/internal/agents/optimizer`
- `src/searchbench-go/internal/pure/report`
- `src/searchbench-go/internal/pure/score`
- `src/searchbench-go/internal/pure/optimizer`
- `src/searchbench-go/internal/adapters/bundle/fs`
- `src/searchbench-go/internal/adapters/config/pkl`

## Boundaries

Keep deterministic SearchBench model code in `src/searchbench-go/internal/pure`. Keep round lifecycle orchestration in `src/searchbench-go/internal/app`. Colocate evaluator- and optimizer-specific behavior under `src/searchbench-go/internal/agents` (prompt + Eino + agent-local persistence/fakes). Keep shared filesystem, Pkl, pipeline, and other non-agent integrations in `src/searchbench-go/internal/adapters`. Keep CLI and terminal rendering in `src/searchbench-go/internal/surface`.

Do not add real MCP, LangSmith, provider execution, dataset materialization, or visualization UI unless the current task explicitly asks for it.

## Naming Rules

Use `Game`, `Round`, `Match`, `IncumbentPolicy`, `ChallengerPolicy`, `Evidence`, `Decision`, and `NextChallenger` for active architecture. Old terms are acceptable only in historical prompts/docs or at explicit external dataset boundaries where upstream data uses that term. Add a short comment for intentional external-boundary exceptions.

## Validation

The routine gate is Git-driven: use `nix develop`, then rely on **`git commit`** and **`git push`** to run hooks. Before handing off code changes, run **`go test ./...`** from `src/searchbench-go` and **`buck2 test //:check_full`** from the repo root (or rely on the hooks after `nix develop`).

For schema changes, regenerate Pkl bindings with:

```sh
cd src/searchbench-go
pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl --output-path=. ../../configs/schema/SearchBenchRound.pkl
```

## Nix + Buck2

The flake provides a **dev shell**, **git-hooks.nix** wiring, and the **Buck2 Nix cell** under `toolchains/` (see `toolchains/flake.nix` and `toolchains/BUCK`). It does **not** ship a separate `nix/tools/` command layer: substantive checks run through **Buck2**.

**`nix develop`** installs Git hooks from git-hooks.nix. The generated `.pre-commit-config.yaml` is gitignored.

| Stage | What runs |
| --- | --- |
| **`git commit` (pre-commit)** | Hygiene (Go/Nix/shell formatting, JSON/YAML/TOML, merge conflicts, …), then **Repomix + `buck2 test //:check`** (regenerates and stages `repomix-output.xml`, then Go tests + CLI build + Iterative Context `//src/iterative-context:check`: import smoke + pytest subset without `TEST_REPO_*` fixtures) |
| **`git push` (pre-push)** | **`buck2 test //:check_full`** (Go `check`, Iterative Context `check_full` including basedpyright, plus Repomix freshness — fails if the snapshot is not already committed at `HEAD`) |

| Command | Purpose |
| --- | --- |
| `nix develop` | Dev shell: Go, Pkl, **buck2**, golangci-lint (manual), `repomix`, pre-commit; writes `.buckconfig.d/buck2-nix.config` for Buck’s `nix` cell |
| `nix develop -c buck2 test //:check` | Same aggregate as the **pre-commit** Buck step (from repo root) |
| `nix develop -c buck2 test //:check_full` | Same as **pre-push** (includes Repomix gate) |
| `nix develop -c pre-commit run --all-files` | Full dev hook run (same family as `git commit` / `git push` stages) |
| `nix flake check` | Sandboxed **non-mutating** checks — formatting / Nix / shell only; **no** Buck tests in the default sandbox |

**Repomix:** This repository **intentionally commits** `repomix-output.xml`. **`repomix.config.json` excludes `repomix-output.*`** so the pack does not recurse into the snapshot. The pre-commit hook uses **`--no-git-sort-by-changes`** and omits git diffs/logs in the committed pack so output stays reproducible. For a richer one-off pack, run **`repomix`** manually with `--include-diffs` / `--include-logs`.

**Go / lint policy:** `.golangci.yml` is for **local** and CI use; hooks no longer invoke golangci automatically — use your editor or `nix develop -c golangci-lint run ./...` from `src/searchbench-go` when you want that pass.

**Orchestration outside this repo:** Worktrees, branch lifecycle, task assignment, agent summary packs, and merge orchestration are owned by an **external meta harness**, not by SearchBench-Go.
