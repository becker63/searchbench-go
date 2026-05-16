# SearchBench-Go

**SearchBench compares agent and tool candidates on the same match slice and produces release evidence** — durable artifacts that support promote, review, or reject decisions.

```text
Game → Round → Match → Evidence → Decision → NextChallenger
```

A **Round** pits an **IncumbentPolicy** against a **ChallengerPolicy** on shared **Matches**, gathers **Evidence**, records a **Decision**, and may propose a **NextChallenger** for another iteration.

## What problem it solves

AI policy changes are hard to ship safely. SearchBench structures the review: same dataset slice, comparable runs, explicit objective, and a bundle that records what happened — not ad-hoc scripts or trace dumps alone.

## Core vocabulary

| Term | Meaning |
| --- | --- |
| **Game** | Domain rules (first game: code localization) |
| **Round** | One incumbent vs challenger contest |
| **Match** | One dataset item in the round |
| **Evidence** | Durable facts from executions |
| **Decision** | `PROMOTE`, `REVIEW`, or `REJECT` |
| **NextChallenger** | Optimizer proposal for a future round |

Details: [docs/concepts.md](docs/concepts.md).

## Run one local round

```bash
git clone https://github.com/becker63/searchbench-go.git
cd searchbench-go
cd src/searchbench-go
go build -o ../../searchbench ./cmd/searchbench
cd ../..
./searchbench run \
  --manifest=configs/rounds/local-ic-vs-jcodemunch/round.pkl \
  --bundle-root="$(pwd)/.tmp-artifacts"
```

Uses the **offline fake-local** path (no live MCP or model APIs). Optional live backends and env vars are documented in [docs/start-here.md](docs/start-here.md).

## Where to read next

| Audience | Start |
| --- | --- |
| Everyone | [docs/start-here.md](docs/start-here.md) |
| Contributors / agents | [AGENTS.md](AGENTS.md) |
| Full index | [docs/README.md](docs/README.md) |

## Requirements (summary)

| Tool | Notes |
| --- | --- |
| [Go 1.26.2](https://go.dev/dl/) | Module under `src/searchbench-go/` |
| [Pkl](https://pkl-lang.org/) | Evaluate round manifests |
| [Nix](https://nixos.org/) *(optional)* | `nix develop` for buck2, hooks, repomix |

Development workflow (tests, hooks, Repomix): [docs/development.md](docs/development.md).

**Docs site** (VitePress): [https://becker63.github.io/searchbench-go/](https://becker63.github.io/searchbench-go/) — `npm run docs:dev` (preview), `npm run docs:build`, `buck2 test //docs:check`; deploys from `main` via GitHub Actions.

## Module

`github.com/becker63/searchbench-go`
