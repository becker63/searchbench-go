# SearchBench-Go

**Test the interfaces your coding agents use.**

SearchBench is a **work-in-progress** harness for evaluating **agent-facing interfaces** over benchmark tasks — not a polished public API or general benchmark platform yet. It wraps task families as **games**, then asks which **interface** (tools, search backends, MCP servers, configs, validation surfaces) makes the **same agent** perform better on the same dataset slice.

The **first stress-test game** is **code localization**: bug-localization dataset slices test whether symbol/code-search tools with **lookahead** help an agent find the files that need to change. Long term, the same `Game` model is intended to wrap other task families (for example SWE-bench-style issue resolution, proof-target selection, docs/context selection) — without claiming those are fully supported today.

```text
Game → Round → Match → Evidence → Decision → NextChallenger
```

A **round** compares an **incumbent** interface or policy against a **challenger** on the same **matches**, records **evidence**, and yields a **decision**: promote, review, or reject.

## What SearchBench does

```text
Benchmark task family
  → SearchBench Game
  → Dataset slice
  → Incumbent interface vs Challenger interface
  → Tool-based agent execution
  → Evidence bundle
  → Report / visualization
  → Promote, review, or reject
```

Most benchmarks ask which **model** is better. SearchBench also asks which **interface** made the same model behave like a better engineer.

**Today:** research workbench, infra-native evaluation harness, durable evidence bundles from controlled rounds.

**Not yet:** stable public API, finished autonomous meta-harness, replacement for SWE-bench, polished SaaS.

Broader thesis: [AGENT_INTERFACE_RESEARCH.md](AGENT_INTERFACE_RESEARCH.md) (research notes; not required to use the harness).

## Core vocabulary

| Term | Meaning |
| --- | --- |
| **Game** | Wrapper around a benchmark task family (first game: code localization) |
| **Interface** | Tools, MCP, graphs, configs, and surfaces the agent uses during a game |
| **Round** | Incumbent vs challenger on the same dataset slice |
| **Match** | One benchmark instance in the slice |
| **Evidence** | Durable facts from executions |
| **Decision** | `PROMOTE`, `REVIEW`, or `REJECT` |

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

Uses the **offline fake-local** path (no live MCP or model APIs). See [docs/start-here.md](docs/start-here.md).

## Where to read next

| Audience | Start |
| --- | --- |
| Everyone | [docs/start-here.md](docs/start-here.md) |
| Contributors / agents | [AGENTS.md](AGENTS.md) |
| Full index | [docs/README.md](docs/README.md) |
| Docs site | [becker63.github.io/searchbench-go](https://becker63.github.io/searchbench-go/) |

## Requirements (summary)

| Tool | Notes |
| --- | --- |
| [Go 1.26.2](https://go.dev/dl/) | Module under `src/searchbench-go/` |
| [Pkl](https://pkl-lang.org/) | Round manifests |
| [Nix](https://nixos.org/) *(optional)* | `nix develop` for buck2, hooks, repomix |

Development: [docs/development.md](docs/development.md). Docs build: `npm run docs:build`, `buck2 test //docs:check`.

## Module

`github.com/becker63/searchbench-go`
