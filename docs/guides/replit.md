# SearchBench-Go (Replit)

**Documentation:** [Documentation index](../README.md) · [Architecture](../architecture/architecture.md) · [Package boundaries](../architecture/package-boundaries.md) · [Roadmap](../roadmap/todo.md)

## Overview

SearchBench is a game-pluggable evaluation system for comparing AI policy changes across datasets. It runs structured "Rounds" to compare an Incumbent policy against a Challenger policy, producing a Decision (PROMOTE, REVIEW, or REJECT).

The first implemented game is `CodeLocalizationGame` — identifying relevant code regions for a bug or task.

## Tech Stack

- **Language:** Go 1.26.2
- **CLI Framework:** Kong
- **AI Agent Framework:** Eino (CloudWeGo)
- **Configuration/Scoring:** Pkl (Apple's configuration language)
- **Logging:** Zap (structured) + Lipgloss (terminal rendering)
- **Templating:** Templ (type-safe HTML/text for prompt rendering)

## Project Structure

```
cmd/searchbench/      # Main CLI entrypoint
doc.go                  # Anchor for root-level integration test package searchbench
e2e_test.go             # Canonical local fake CLI + round.Run integration tests
internal/
  pure/               # Deterministic SearchBench model
  app/                # Round lifecycle orchestration (internal/app/round)
  agents/             # Evaluator/optimizer prompts, Eino runners, consolidated local fakes
  adapters/           # Shared runtime (config/pkl, scoring/pkl, bundle/fs, pipeline/exec)
  ports/              # Narrow contracts for pipeline execution and datasets
  surface/            # CLI routing and console rendering
  games/              # Game implementations (codelocalization)
  architecture/       # Import boundary enforcement tests
configs/
  schema/             # Pkl schemas
  rounds/             # Round configuration files
```

## Building & Running

```bash
# Build the binary
go build -o searchbench ./cmd/searchbench

# Run help
./searchbench --help

# Run a round
./searchbench round run --manifest=configs/rounds/<manifest>.pkl
```

### Nix (optional, recommended for agents)

```bash
nix develop                    # dev shell + pre-commit (Buck2 + hygiene)
nix flake check                # sandboxed checks (no Buck — quick Nix/shell/format gate)
nix develop -c buck2 test //:check
nix develop -c buck2 test //:check_full
cd src/searchbench-go && golangci-lint run ./...
cd src/searchbench-go && go test -count=1 .
```

See the root [`AGENTS.md`](../../AGENTS.md) for Git hook stages (**`git commit`** / **`git push`**) and Repomix.

---

- **Game:** The ruleset/domain (e.g., Code Localization)
- **Round:** A single contest comparing IncumbentPolicy vs ChallengerPolicy
- **Match:** A single task/dataset item within a round
- **Evidence:** Durable facts produced during matches
- **Decision:** Round outcome — PROMOTE, REVIEW, or REJECT

## Architecture Principle

Strict "Pure Center" layering: `pure/` has no external dependencies. `adapters/` implement `ports/` interfaces. Import boundaries are enforced by tests in `internal/architecture/`.

## User Preferences

- CLI tool only — no web frontend
- Build with `go build -o searchbench ./cmd/searchbench` from `src/searchbench-go`, or `nix develop -c buck2 test //:check` to exercise the same aggregate the hooks use.
