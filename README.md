# SearchBench-Go

<p align="center">
  <strong>Test the interfaces your coding agents use.</strong>
</p>

<p align="center">
  <a href="https://becker63.github.io/searchbench-go/">📚 Docs</a>
  ·
  <a href="docs/start-here.md">🚀 Start here</a>
  ·
  <a href="docs/concepts.md">🧭 Concepts</a>
  ·
  <a href="docs/components.md">🧩 Components</a>
  ·
  <a href="docs/development.md">🛠️ Development</a>
  ·
  <a href="docs/research/agent-interface-research.md">🔬 Research</a>
</p>

---

SearchBench is a **work-in-progress harness** for evaluating **agent-facing interfaces** over benchmark tasks.

The current user-facing API is simple:

```text
Pkl round manifest
  → searchbench run
  → evidence bundle
  → decision/report
```

SearchBench wraps task families as **games**, then asks:

> Which interface makes the same agent perform better?

The first stress-test game is **code localization**. SearchBench uses bug-localization dataset slices to test whether symbol/code-search tools with **lookahead** help an agent find the files that need to change.

Long term, the same `Game` model is intended to wrap other task families — for example SWE-bench-style issue resolution, proof-target selection, and docs/context selection — without claiming those are fully supported today.

```text
Game → Round → Match → Evidence → Decision → NextChallenger
```

## Current API: Pkl games and rounds

SearchBench is configured through **Pkl manifests**.

A round manifest declares:

- the game being played
- the dataset slice
- the incumbent interface
- the challenger interface
- evaluator/tool policy
- scoring objective
- output bundle location

The CLI executes that manifest and writes a static evidence bundle.

```bash
./searchbench run \
  --manifest=configs/rounds/local-ic-vs-jcodemunch/round.pkl \
  --bundle-root="$(pwd)/.tmp-artifacts"
```

Conceptually:

```text
configs/rounds/*/round.pkl
  declares the round

configs/schema/
  defines the Pkl contract

searchbench run
  executes the round

.tmp-artifacts/
  receives bundles, reports, decisions, and continuation files
```

The Go library internals are still evolving. For now, the stable surface to look at is the **Pkl round/config interface** plus the CLI.

## Why SearchBench exists

Most benchmarks ask:

> Which model is better?

SearchBench also asks:

> Which interface made the same model behave like a better engineer?

Interfaces can include:

- code-search backends
- MCP servers
- symbol/reference graphs
- bounded lookahead policies
- validation/proof targets
- Pkl configs
- Buck work graphs
- docs/context packs
- visualization surfaces

SearchBench compares those interfaces under controlled rounds and records durable evidence about what happened.

## What it does

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

A **round** compares an incumbent interface or policy against a challenger on the same matches, records evidence, and yields a decision:

```text
PROMOTE | REVIEW | REJECT
```

## Status

SearchBench is early.

**Today:**

- research workbench
- infra-native evaluation harness
- Pkl round/config API
- code-localization stress test
- bundle-producing round runner
- local fake/offline evaluation path
- Buck/Nix development workflow

**Not yet:**

- stable Go library API
- finished autonomous meta-harness
- replacement for SWE-bench
- polished SaaS product

## Quickstart

Run one local fake/offline round:

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

This uses the **offline fake-local** path. It does not require live MCP servers or model APIs.

See [docs/start-here.md](docs/start-here.md) for the guided walkthrough.

## Core vocabulary

| Term | Meaning |
| --- | --- |
| **Game** | Wrapper around a benchmark task family. First game: code localization. |
| **Interface** | Tools, graphs, configs, MCP servers, and validation surfaces exposed to the agent. |
| **Round** | Incumbent vs challenger comparison on the same dataset slice. |
| **Match** | One benchmark instance in the slice. |
| **Evidence** | Durable facts from execution, scoring, validation, and reporting. |
| **Decision** | `PROMOTE`, `REVIEW`, or `REJECT`. |
| **Bundle** | Static artifact directory recording what happened. |

More: [docs/concepts.md](docs/concepts.md)

## Repository map

| Path | Purpose |
| --- | --- |
| `configs/` | Pkl schemas, round manifests, objectives, dataset slices |
| `src/searchbench-go/` | Go harness: games, rounds, scoring, bundles, CLI |
| `src/iterative-context/` | Python MCP/code-search backend used as a challenger interface |
| `docs/` | Public docs, reference docs, and research notes |
| `tooling/` | Repo lifecycle tools such as Repomix |
| `BUCK` | Root Buck gates: `//:check`, `//:check_full` |

More: [docs/components.md](docs/components.md)

## Development

SearchBench’s development workflow is intentionally graph-shaped:

```text
Nix gives you the environment.
Buck gives you the work graph.
Git hooks call Buck gates.
```

Common commands:

```bash
nix develop
buck2 test //:check
buck2 test //:check_full
buck2 test //docs:check
```

Docs preview:

```bash
npm run docs:dev
```

More: [docs/development.md](docs/development.md)

## Research thesis

SearchBench is exploring a broader claim:

> The model is not the whole system.  
> The interface changes what the agent is capable of.

Read the research note:

- [Agent interface research](docs/research/agent-interface-research.md)

## Links

- 📚 Docs: [becker63.github.io/searchbench-go](https://becker63.github.io/searchbench-go/)
- 🚀 Start here: [docs/start-here.md](docs/start-here.md)
- 🧭 Concepts: [docs/concepts.md](docs/concepts.md)
- 🧩 Components: [docs/components.md](docs/components.md)
- 🛠️ Development: [docs/development.md](docs/development.md)
- 🤖 Agent contract: [AGENTS.md](AGENTS.md)

## Module

```text
github.com/becker63/searchbench-go
```
