# Buck / work-graph as agent interface (research)

**Status:** research direction, not a shipped SearchBench feature.

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

## Pattern (aspirational)

```text
agent emits semantic intent (structured operation)
  → renderer edits sanctioned repo graph (e.g. BUCK)
  → Buck validates dependencies
  → tests prove behavior
  → bundle records evidence
```

SearchBench today uses Buck for **contributor gates** (`//:check`, `//:check_full`, `//docs:check`) and IC descriptor targets — not yet for agent-emitted graph edits.

## Relation to SearchBench

Same thesis as [agent-interface-research.md](./agent-interface-research.md): compare **interfaces**, not only models. Buck/work-graph is one candidate interface family alongside code-search tools and MCP surfaces.

External **meta-harness** (worktrees, batch issues, merge orchestration) lives outside this repository; see root [AGENTS.md](../../AGENTS.md).
