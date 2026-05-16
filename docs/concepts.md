# Concepts

SearchBench uses one vocabulary in code, CLI output, Pkl manifests, bundles, and docs.

```text
Game → Round → Match → Evidence → Decision → NextChallenger
```

## Interface

An **interface** is anything that changes what the agent can see, query, validate, or act on during a game.

Examples: code-search tools, MCP servers, graph traversal and bounded lookahead, symbol/reference lookup, Pkl round configs, workspace seed providers, Buck-backed validation targets, docs/context packs, visualization projections.

Rounds compare **incumbent** and **challenger** interfaces (or policies that bundle them) on the same slice.

## Game

A **Game** wraps a **benchmark or task family**. It defines:

- dataset / match input shape
- tool and interface contract
- allowed outputs
- evidence to collect
- scoring and objective rules

SearchBench is a **generic harness** for turning task families into games. It is **not** only a code-localization benchmark.

**First stress-test game: code localization** — bug reports and repository context over a slice of instances. This game isolates: *can the agent find the right place to look?* It is used to evaluate **symbol/code-search interfaces with lookahead** (grep, symbol search, graph hops, MCP-backed search, configuration-driven tool policies).

**Possible future games** (not all implemented): SWE-bench-style issue resolution, proof-target selection, docs/context selection, MCP tool comparison, repository-operation tasks, config validation, release-candidate promotion.

## Dataset slice

A **dataset slice** is the bounded set of **matches** used for one round. It keeps comparison fair:

```text
same model (when held constant)
same match slice
same scoring objective
different interface or policy
```

## Round

A **Round** is one contest under a game: one **IncumbentPolicy** vs one **ChallengerPolicy** over a fixed dataset slice.

A round produces match executions, evidence, an objective result, a **Decision**, a durable **bundle**, and optionally a **NextChallenger** proposal. Completed rounds are immutable.

## Match

A **Match** is one benchmark instance in the slice (older docs may say “task”). Each match runs incumbent and challenger on the same input and records per-side outputs, usage, failures, and match-level evidence.

## IncumbentPolicy and ChallengerPolicy

- **IncumbentPolicy** — the interface or policy being defended (baseline, previous winner, reference tool).
- **ChallengerPolicy** — the candidate under test (prompt, code, retrieval, tools, bounds, provider settings).

## Evidence

**Evidence** is durable facts from a round: scores, regressions, usage, failures, artifact hashes, and domain summaries. Evidence supports a **decision**; a trace alone is debugging material, not release proof.

## Decision

A **Decision** applies the game’s objective and release rule to round evidence:

```text
PROMOTE | REVIEW | REJECT
```

## NextChallenger

The optimizer may propose a **NextChallenger** — a file-backed candidate for a future round. It does not change production by itself.

## Bundle

A **bundle** records what happened in a round: inputs, predictions, scores, validation evidence, decision, and candidate identity — the basis for reports and **visualization** (inspecting what the agent looked at, whether lookahead helped, why the challenger won or lost).

## Meta-harness (long term)

Autonomous loops (agents propose interface changes → SearchBench evaluates → bundles drive promotion) are an **advanced / internal** direction. The current system emphasizes comparable rounds and durable evidence before that loop is productized.

## Deeper reference

Extended naming and bundle layout: [reference/architecture-full.md](./reference/architecture-full.md).

Research thesis: [AGENT_INTERFACE_RESEARCH.md](https://github.com/becker63/searchbench-go/blob/main/AGENT_INTERFACE_RESEARCH.md).
