# Concepts

```text
Game → Round → Match → Evidence → Decision → NextChallenger
```

## Interface

What the agent can see, query, validate, or act on: code-search tools, MCP servers, graph lookahead, Pkl configs, workspace seeds, Buck validation targets.

## Game

Wrapper around a **benchmark task family**: match shape, tool contract, outputs, evidence model, scoring rules.

SearchBench is a **generic harness**, not only code localization. **First stress-test game:** code localization — *can the agent find the right place to look?* via symbol/code-search interfaces with lookahead.

Future directions (not all built): SWE-bench-style resolution, proof-target selection, docs/context selection, MCP tool comparison.

## Dataset slice

Bounded **matches** for one round — same slice, same objective, different incumbent vs challenger.

## Round

**IncumbentPolicy** vs **ChallengerPolicy** on that slice → executions, **Evidence**, **Decision** (`PROMOTE` | `REVIEW` | `REJECT`), durable **bundle**, optional **NextChallenger**.

## Match

One benchmark instance in the slice.

## Evidence

Durable facts for decisions (scores, usage, failures, hashes) — not raw traces alone.

## Bundle

Immutable record of a round: inputs, outputs, validation, decision, candidate identity — basis for inspection and reports.

## NextChallenger

Optimizer proposal for a future round; does not change production by itself.

## Research

Broader thesis: [research/agent-interface-research.md](./research/agent-interface-research.md).
