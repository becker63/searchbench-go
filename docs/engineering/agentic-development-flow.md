# Agentic Development Flow

## Purpose

This document describes the development loop used for SearchBench-Go when working with coding agents.

The project is not being built by handing an agent a vague goal and asking it to “make progress.” Instead, development is organized as a contract-driven loop:

    think with ChatGPT
    write a GitHub issue
    write an implementation prompt
    run an agent
    inspect the result
    pack/summarize the result
    feed new information back into the next issue

This workflow exists because SearchBench-Go is a migration, not a blank-slate greenfield project.

The goal is to preserve the hard-earned semantics from Python SearchBench while avoiding a mechanical port of the Python implementation shape.

## Core idea

The GitHub issue is the implementation contract.

The Codex/agent prompt is the execution wrapper around that contract.

The agent result is not the end of the loop. It becomes new evidence for the next design pass.

The human remains responsible for:

- architectural taste
- issue ordering
- scope control
- deciding what old semantics matter
- noticing dangling future work
- deciding when a result should become a new contract

The agent is responsible for:

- implementing bounded issues
- following repository guidance
- running tests
- preserving explicit constraints
- reporting deviations

## Development loop

The normal flow is:

    1. Discuss design with ChatGPT

    2. Convert the design into a GitHub issue

    3. Add explicit constraints:
        - goal
        - background
        - architecture rules
        - phase flow
        - non-goals
        - expected behavior
        - acceptance criteria
        - tests
        - review checklist

    4. Write a fresh agent prompt for that issue

    5. Tell the agent to:
        - read AGENTS.md
        - read relevant docs
        - read the GitHub issue with gh cli
        - treat the issue as source of truth
        - stop on conflicts

    6. Run the agent

    7. Pack or summarize the result

    8. Review what the result exposed

    9. Create follow-up issues, clarifications, or dangling cleanup notes

    10. Repeat

## Why this works

This workflow gives the project a feedback loop without letting the agent own the architecture.

Issues encode intent.

Prompts encode execution discipline.

Agent results expose reality.

Follow-up issues absorb what was learned.

This is especially useful for SearchBench-Go because many decisions are not purely local. A small implementation result may reveal:

- a missing domain type
- a bad package boundary
- a fake implementation becoming too architectural
- a testing seam that should exist first
- a future refactor that should not happen yet
- a constraint that needs to be made explicit in the next issue

## Issue-first development

Issues should be written before implementation whenever possible.

A good issue should make the desired implementation feel obvious without writing the implementation for the agent.

Use prose for intent.
Use phase pseudocode for lifecycle.
Use acceptance criteria for review.
Use non-goals to prevent architecture drift.

The issue should be specific enough that an agent can execute it, but not so specific that it cargo-cults code from the prompt.

## Prompt-after-issue development

The agent prompt should not replace the issue.

The prompt should tell the agent how to approach the issue.

A good prompt tells the agent to:

- read the target GitHub issue using `gh issue view`
- read AGENTS.md
- read project docs
- inspect relevant package docs
- prefer the issue body as source of truth
- stop if the issue conflicts with architecture guidance
- avoid implementing from memory
- run the expected tests
- summarize changes and deviations

The prompt may restate important constraints, but the GitHub issue remains the contract.

## Packed result feedback

After an agent implements an issue, the result should be packed or summarized and fed back into the next design pass.

This can include:

- changed files
- package structure
- tests added
- unexpected implementation choices
- missing abstractions
- naming problems
- places where the agent had to improvise
- places where the issue was underspecified

This is important because implementation teaches us things that design discussion alone cannot.

The next issue should use that information.

I use repomix for this. And just pack the whole repo into chatgpt with the paste buffer `wl-copy`.

## Dangling issue capture

Not every useful observation should become immediate work.

Some observations should become dangling issues.

A dangling issue is a parking-lot item for future refactor or audit work.

Example:

    audit and separate pure model code from effectful adapter code

This kind of issue is valuable because it records architectural pressure without derailing the current implementation sequence.

The ability to notice dangling activity is important.

It means the workflow is not only producing code. It is also surfacing project-shape information.

## Why dangling issues matter

During implementation, we often discover future cleanup pressure before it is time to act on it.

Examples:

- pure domain code may start drifting toward adapter concerns
- Eino execution types may leak into core models
- testing helpers may start shaping production architecture
- prompt inputs may become too coupled to executor packages
- CLI pipeline types may want to become shared run records
- old Python semantics may be preserved in the wrong layer

Stopping immediately to fix all of this would break sequencing.

Ignoring it would lose important architectural signal.

A dangling issue preserves the signal without changing the current task.

## Information flow

The workflow intentionally creates a chain of information:

    design discussion
        ↓
    issue contract
        ↓
    agent prompt
        ↓
    implementation result
        ↓
    packed summary / review
        ↓
    next issue or dangling issue

This means the project becomes more specified over time.

The agent does not merely consume the issue list.
The agent’s results help generate the next issue list.

## Ordering principle

Order issues by information exposure.

Prefer proving seams before composing systems.

A useful sequence is:

    1. define deterministic testing fixtures
    2. define domain schema
    3. prove minimal evaluator loop
    4. add retry and CLI pipeline validation
    5. prove MCP tool wiring
    6. add process/session lifecycle
    7. add policy installation
    8. run IC evaluator
    9. add scoring
    10. compose full run
    11. add tracing
    12. add writer/optimization loop
    13. perform package-boundary sweep

Do not ask an agent to build the full system before the seams exist.

## What the human should watch for

During review, look for:

- duplicate domain models
- adapter leakage into pure packages
- fake implementations becoming architecture
- unbounded command execution
- hidden real API calls
- tests that spend money
- vague failure handling
- missing phase names
- missing non-goals
- unnecessary framework creation
- agent-added abstractions that were not requested

When one of these appears, decide whether it is:

    immediate fix
    issue clarification
    follow-up issue
    dangling issue

## What the agent should not own

The agent should not own:

- issue ordering
- architectural layering
- vendor/tooling strategy
- broad refactor scope
- deciding which Python semantics matter
- deciding when to preserve or delete old implementation shape

The agent may suggest, but the human decides.

## Preferred artifact types

Use these artifacts deliberately:

    GitHub issue
      implementation contract

    Codex/agent prompt
      execution instructions

    packed repo/result summary
      evidence for next planning pass

    dangling issue
      future cleanup signal

    engineering doc
      stable project/process knowledge

## Final rule

The workflow is successful when every implementation result either:

- lands a bounded slice
- exposes the next missing contract
- creates a useful dangling issue
- clarifies a project boundary

The point is not just to get code written.

The point is to steadily convert hard-earned project understanding into contracts that agents can execute without owning the architecture.
