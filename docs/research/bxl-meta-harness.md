# Buck / work-graph as agent interface

## Status

**Partially implemented** — initial BXL planners ship under `tooling/bxl/` ([issue #94](https://github.com/becker63/searchbench-go/issues/94)). They emit static JSON only. Meta-harness orchestration (worktrees, batch issues, merge policy) remains out of repo.

## Idea

Many agent failures are lifecycle failures (wrong checks, wrong layer, unclear “done”). A **work graph** (Buck2 targets as named legal moves) can expose lifecycle rules as structure instead of prose.

In this framing, Buck is an **agent-facing action graph**:

```text
//:check
//:check_full
//src/searchbench-go:check
//src/iterative-context:check_full
```

instead of asking the agent to invent `go test`, `pytest`, `ruff`, etc. from README tribal knowledge.

BXL adds **lookahead over action**: given a backend label or a list of changed files, planners suggest proof targets and comparison structure before any live eval runs.

```text
Code graph:  lookahead over meaning
Work graph:  lookahead over action   ← tooling/bxl (first slice)
```

## What shipped (#94)

| Planner | JSON `kind` |
|---------|-------------|
| `target_summary` | `searchbench.bxl_target_summary.v1` |
| `resolve_backend` | `searchbench.backend_resolution.v1` |
| `list_backends` | `searchbench.backend_inventory.v1` |
| `proof_plan` | `searchbench.proof_plan.v1` |
| `affected_plan` | `searchbench.affected_plan.v1` |
| `evaluation_matrix` | `searchbench.evaluation_matrix.v1` |

Reference: [buck-work-graph.md](../reference/buck-work-graph.md). Commands: [tooling/bxl/README.md](../../tooling/bxl/README.md).

## Pattern (target)

```text
agent emits semantic intent (structured operation)
  → renderer edits sanctioned repo graph (e.g. BUCK)
  → Buck validates dependencies
  → tests prove behavior
  → bundle records evidence
```

SearchBench today uses Buck for **contributor gates** and IC descriptor targets. BXL planners are the first step toward recording **which proof path the optimizer considered** alongside RoundBundle evidence.

## Relation to SearchBench

Same thesis as [agent-interface-research.md](./agent-interface-research.md): compare **interfaces**, not only models. Buck/work-graph is one candidate interface family alongside code-search tools and MCP surfaces.

External **meta-harness** (worktrees, batch issues, merge orchestration) should consume BXL JSON plans and still execute evaluations via Go SearchBench; see root [AGENTS.md](../../AGENTS.md).

## Non-goals (unchanged)

BXL must not run optimizer loops, provider APIs, live MCP eval, Git mutation, or replace Pkl/Go for public users.
