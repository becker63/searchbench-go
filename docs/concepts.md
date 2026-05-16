# Concepts

SearchBench uses one product vocabulary everywhere: code, CLI output, Pkl manifests, bundles, and docs.

```text
Game → Round → Match → Evidence → Decision → NextChallenger
```

## Game

A **Game** is the domain contract: what is being evaluated, what counts as success, and how humans review failures.

A game owns match schema, evidence shape, scoring inputs, review panes, and win/loss semantics. It is not a single execution — it is the ruleset.

The first implemented game is **Code Localization** (repository as board, bug/issue as match input, localization as goal).

## Round

A **Round** is one contest under a game: one **IncumbentPolicy** vs one **ChallengerPolicy** over a fixed match slice.

A round produces match executions, evidence, an objective result, a **Decision**, a durable bundle, and optionally a **NextChallenger** proposal. Completed rounds are immutable.

## Match

A **Match** is one dataset item inside a round (older docs may say “task”). Each match runs both policies on the same input and collects per-side outputs, usage, failures, and match-level evidence.

## IncumbentPolicy and ChallengerPolicy

- **IncumbentPolicy** — the policy being defended (production baseline, previous winner, or reference system).
- **ChallengerPolicy** — the policy under test (prompt, code, retrieval, tools, bounds, provider settings).

## Evidence

**Evidence** is the durable facts from a round: scores, regressions, usage, failures, artifact hashes, and domain-specific summaries. Evidence supports a decision; a trace alone is debugging material, not release proof.

## Decision

A **Decision** applies the game’s objective and release rule to round evidence:

```text
PROMOTE | REVIEW | REJECT
```

## NextChallenger

The optimizer proposes a **NextChallenger** — a file-backed candidate for a future round. It does not mutate production directly; a later round may adopt it as the challenger.

## Deeper reference

Extended naming, bundle layout, and migration notes: [reference/architecture-full.md](./reference/architecture-full.md).
