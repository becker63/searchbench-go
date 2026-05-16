# Concepts

```text
Game → Round → Match → Evidence → Decision → NextChallenger
```

Each term below anchors on a checked-in file or artifact.

## Game

**Example:** `code-localization` — `configs/schema/games/code-localization.pkl`

**Meaning:** Benchmark task family: match shape, tool contract, evidence model, scoring hooks.

## Dataset slice

**Example:** `matches = game.lca("py", "dev", 5)` in `configs/rounds/local-ic-vs-jcodemunch/round.pkl`

**Meaning:** Bounded set of **matches** (here, five rows from the LCA bug-localization dataset).

## Round

**Example:** `configs/rounds/local-ic-vs-jcodemunch/round.pkl`

**Meaning:** One incumbent vs one challenger on the same slice → evidence, objective, decision, bundle.

**Continuation example:** `configs/rounds/optimize-ic/round.pkl` (amends prior `continuation.pkl`).

## Match

**Example:** one row from the LCA slice (resolved in `round-report.json` under the bundle)

**Meaning:** Single benchmark instance — incumbent and challenger each run against it.

## Interface

**Examples:**

| Role | Interface | Policy / backend |
| --- | --- | --- |
| Incumbent | jCodeMunch | `game.jcodemunch()` |
| Challenger | Iterative Context | `game.iterativeContext("policies/challenger_policy.py")` |

**Meaning:** What the agent can query — MCP servers, search tools, selection policies.

## Objective

**Example:** `configs/rounds/local-ic-vs-jcodemunch/scoring/localization-objective.pkl`

**Output:** `objective.json` in the bundle (e.g. `configs/rounds/local-ic-vs-jcodemunch/artifacts/.../round-001/objective.json`)

**Meaning:** Pkl scoring math over evidence; Go builds evidence and validates the result.

## Evidence

**Example:** `evidence.pkl` in the round bundle

**Meaning:** Durable facts for scoring (distances, usage, regressions) — not raw traces alone.

## Bundle

**Example:** `configs/rounds/local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001/`

**Meaning:** Immutable round record: inputs, reports, evidence, objective, decision, continuation.

See [reference/bundles.md](./reference/bundles.md).

## Decision

**Example:** `decision.json` — `"decision": "PROMOTE_CHALLENGER"`

**Meaning:** Harness verdict on challenger vs incumbent for this round.

## Candidate workspace

**Example:** `localPath = "src/iterative-context"` in a round `runtime.workspaceSeed` block

**Code:** `src/searchbench-go/internal/adapters/workspace/`

**Meaning:** Isolated copy of a backend; validation and MCP launch use the same tree. See [candidate-workspaces.md](./candidate-workspaces.md).

## NextChallenger

**Example:** `configs/rounds/optimize-ic/artifacts/.../round-002/policies/next_challenger_policy.round-002.py`

**Meaning:** Optimizer output for a future round; does not change production by itself.

## Research

Broader thesis: [research/agent-interface-research.md](./research/agent-interface-research.md).
