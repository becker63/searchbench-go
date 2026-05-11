# SearchBench-Go

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
cmd/searchbench/      # Main entry point
internal/
  pure/               # Core domain models (Game, Round, Match, Policy, Score)
  app/                # Use-case orchestration (round, compare, evaluation, optimizer)
  adapters/           # Concrete implementations (config/pkl, scoring/pkl, executor/eino, bundle/fs)
  ports/              # Interfaces for effectful operations
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

## Key Concepts

- **Game:** The ruleset/domain (e.g., Code Localization)
- **Round:** A single contest comparing IncumbentPolicy vs ChallengerPolicy
- **Match:** A single task/dataset item within a round
- **Evidence:** Durable facts produced during matches
- **Decision:** Round outcome — PROMOTE, REVIEW, or REJECT

## Architecture Principle

Strict "Pure Center" layering: `pure/` has no external dependencies. `adapters/` implement `ports/` interfaces. Import boundaries are enforced by tests in `internal/architecture/`.

## User Preferences

- CLI tool only — no web frontend
- Build with `go build -o searchbench ./cmd/searchbench`
