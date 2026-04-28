# SearchBench-Go Issue Style Guide

## Purpose

SearchBench-Go issues are implementation contracts.

They are not vague task reminders, and they are not full design documents. They should give an agent or human implementer enough context to complete a bounded piece of work without rediscovering the architecture or copying old Python implementation shape.

Good issues should preserve:

- project intent
- architectural boundaries
- lifecycle semantics
- acceptance criteria
- non-goals
- failure behavior

Good issues should avoid:

- over-prescribing exact implementation code
- hiding important constraints in prose
- broad “port this subsystem” requests
- allowing agents to invent architecture
- leaking old Python structure into Go unnecessarily

## Core principle

Each issue should answer:

    What are we building?
    Why does it exist?
    What existing project decision does it rely on?
    What must be preserved from old SearchBench?
    What must not be preserved?
    What packages may be touched?
    What are the observable success conditions?
    What failure cases must be handled?

## Recommended issue structure

Use this shape unless the issue is intentionally tiny.

    # Title

    ## Goal

    ## Background

    ## Architecture rule

    ## Phase flow

    ## Scope

    ## Non-goals

    ## Expected behavior

    ## Failure behavior

    ## Acceptance criteria

    ## Suggested package placement

    ## Test expectations

    ## Review checklist

    ## Follow-up issues unlocked by this

Not every issue needs every section, but important implementation issues should usually include most of them.

## Titles

Titles should be imperative and specific.

Good:

    Prove Eino official MCP tool integration with a minimal MCP server

    Implement harness-owned policy installation for IC MCP sessions

    Make graph-distance localization the canonical scoring interface

Bad:

    Add MCP

    Fix backend

    Port scoring

    Improve evaluator

## Goal

The goal should state the smallest useful outcome.

Good:

    Implement the smallest possible SearchBench-Go evaluator loop using Eino, without MCP, Iterative Context, jCodeMunch, repo materialization, policy injection, or graph scoring.

Bad:

    Implement the evaluator.

The goal should make scope obvious.

## Background

Background should explain why this issue exists and what prior decision it depends on.

It should not become a full essay.

Good background answers:

- Why now?
- What previous SearchBench lesson applies?
- What are we deliberately simplifying?
- What should future issues be able to rely on after this lands?

## Architecture rule

Every nontrivial issue should include architectural constraints.

Use this section to prevent agent drift.

Example:

    SearchBench-Go owns:
    - domain models
    - prompt inputs
    - final prediction normalization
    - execution result typing

    Eino owns:
    - agent/model execution
    - tool calling
    - intermediate reasoning/tool loop behavior

    Do not introduce:
    - custom ReAct loops
    - custom state machines
    - backend abstractions
    - MCP-specific code in this issue
    - policy machinery

This section should be blunt. Agents benefit from strong negative constraints.

## Phase pseudocode

Use phase pseudocode to specify lifecycle behavior without over-prescribing exact Go implementation.

Phase pseudocode is not literal Go code. It defines:

- required phase ordering
- side effects
- external calls
- failure behavior
- retry/accept/exhaust decisions
- outcome shape

Example:

    Run(ctx, task)
      create execution result accumulator

      phase: render_prompt
          build typed evaluator prompt input
          render .templ XML-style prompt
          if render failed:
              return prompt_render_failed

      phase: run_evaluator
          run Eino evaluator agent
          expose allowed tools
          if evaluator failed:
              return evaluator_failed
          if tool failed:
              return tool_call_failed

      phase: finalize_prediction
          parse final structured JSON response
          validate predicted_files
          if malformed:
              return finalization_failed
          if predicted_files empty:
              return invalid_prediction

      phase: complete
          return successful LocalizationExecutionResult

This is a middle layer between prose and code.

Use it when lifecycle matters.

## Phase naming

Prefer stable, lower_snake_case phase names.

Examples:

    render_prompt
    run_evaluator
    finalize_prediction
    generate_candidate
    run_pipeline
    classify_pipeline
    accept_candidate
    prepare_retry
    exhausted

Phase names should appear in code, logs, results, or failure values when useful.

Do not introduce a state machine library just because phases exist.

## Preserve semantics, not old implementation shape

When referencing old Python SearchBench, distinguish between what should survive and what should be deleted.

Use this format:

    Preserve:
    - writer/evaluator isolation
    - explicit policy identity
    - typed outcomes
    - bounded repair attempts
    - external pipeline failure classification

    Do not preserve:
    - Python state machine framework
    - Langfuse-specific implementation shape
    - Python dict-shaped models
    - worktree mutation as policy selection
    - custom MCP adapter machinery if Eino already buys it

This helps agents avoid mechanical ports.

## Scope

Scope should list what the issue may implement.

Example:

    Create a minimal evaluator path that uses:
    - one fake/local tool
    - one evaluator prompt rendered from a .templ XML-style prompt document
    - one structured final prediction
    - one typed execution result
    - one lightweight phase-oriented runner

Scope should be narrow enough that the PR can be reviewed against the issue.

## Non-goals

Non-goals are mandatory for agent-facing issues.

Use them aggressively.

Example:

    This issue must not implement:
    - MCP
    - Iterative Context
    - jCodeMunch
    - repo materialization
    - policy artifacts
    - policy injection
    - graph-distance scoring
    - writer agent
    - optimization loop
    - LangSmith tracing
    - GitHub issue automation
    - custom state machine library

Non-goals prevent “helpful” overbuilding.

## Prompt decisions

Prompt-related issues should preserve the project’s prompt boundary.

Current decision:

    SearchBench-Go uses .templ files to render XML-style prompt documents.

    The .templ file is a contract boundary.
    Prompt structure lives in .templ.
    Prompt data shape lives in typed Go input structs.
    Render functions produce strings for Eino/model usage.
    Render tests lock expected prompt sections.

Expected shape:

    internal/pure/prompts/evaluator/
      input.go
      prompt.templ
      render.go
      prompt_test.go

Prompt documents may use XML-style sections:

    <searchbench_prompt>
    <role>
    <task>
    <issue>
    <available_tools>
    <constraints>
    <output_contract>

Final model outputs are not XML. Final outputs should be strict JSON or schema-constrained domain results.

## Fake implementation policy

Fake implementations are allowed when they prove a seam.

They must not become fake architecture.

Good fake:

    A test-local fake tool that proves Eino can call tools and finalization works.

Bad fake:

    A fake backend package that shapes production backend abstractions before MCP/IC/JCM are real.

Use this language when relevant:

    This issue may use fake/local tools only to prove the execution seam.

    Fake tools must remain test-local or clearly marked as fixtures. They must not become production backend abstractions.

    Later issues should replace the fake execution path with real MCP-backed tools, but the fake should remain as a low-cost regression test for the executor/finalizer layer.

## Failure behavior

Issues should say how failures become typed results.

Good:

    The result should distinguish:
    - prompt render failure
    - evaluator/model failure
    - tool failure
    - malformed final output
    - empty predicted_files
    - unexpected internal failure

Bad:

    Handle errors.

For external tools, prefer step results:

    StepResult
      name
      command
      exit_code
      stdout_tail
      stderr_tail
      duration
      passed

For classification, prefer structured categories:

    type_errors
    lint_errors
    test_failures
    generation_failures
    infrastructure_failures

## Acceptance criteria

Acceptance criteria should be checkboxes.

They should be observable.

Good:

    - [ ] The evaluator runner has one explicit Run-style entrypoint.
    - [ ] Runner construction is cold; no model/tool execution happens during construction.
    - [ ] Failures record the phase where they occurred.
    - [ ] The evaluator prompt is rendered from a .templ XML-style prompt document.
    - [ ] predicted_files is required and cannot be empty on success.
    - [ ] This is implemented with ordinary Go control flow, not a state machine library.

Bad:

    - [ ] Works well.
    - [ ] Is clean.
    - [ ] Uses good architecture.

## Suggested package placement

Package placement should guide without freezing exact implementation.

Use language like:

    Prefer using the existing project structure. If no better location already exists, a reasonable shape is:

        internal/pure/prompts/evaluator/
          input.go
          prompt.templ
          render.go
          prompt_test.go

        internal/implementation/executor/eino/
          evaluator.go
          finalizer.go
          errors.go
          phases.go

Then include constraints:

    Do not move core domain types into the Eino package.
    Eino should be an implementation detail.
    Prompt templates are a contract boundary and should remain easy to inspect independently.

## Test expectations

Tests should prove behavior and boundaries, not incidental implementation details.

Good test expectations:

    Add tests that prove:
    1. evaluator prompt renders from typed input
    2. rendered prompt contains expected XML-style sections
    3. runner construction is cold
    4. Run executes named phases in order
    5. fake tool failure becomes tool_call_failed
    6. empty predicted_files is rejected
    7. failures include phase and failure kind

Avoid tests that lock in private helper names, exact prompt whitespace, or incidental data structure layout.

## Review checklist

Every issue should include a reviewer checklist that reflects architectural taste.

Example:

    A reviewer should verify:

    - Did the implementation avoid rebuilding a framework?
    - Are domain types kept pure?
    - Is Eino isolated to execution-layer code?
    - Does the prompt use .templ as a typed contract boundary?
    - Are phase failures typed and distinguishable?
    - Are side effects isolated to execution phases?
    - Are CLI/tracing/backend concerns kept outside the core lifecycle?
    - Did this issue avoid premature MCP/backend/policy/scoring work?

Review checklists should be specific to the issue, not generic.

## Follow-up issues unlocked by this

Each issue should say what it enables next.

Example:

    After this issue lands, future agents should be able to build:
    - Eino MCP tool wiring
    - MCP server process manager
    - jCodeMunch external MCP baseline
    - Iterative Context policy installation
    - IC evaluator runtime
    - graph-distance scoring

This helps preserve the migration sequence.

## Issue ordering principle

Order issues by information exposure, not just final product importance.

Prefer this sequence:

    1. prove minimal local execution
    2. prove MCP tool wiring
    3. prove process/session lifecycle
    4. wire external baseline
    5. define policy artifacts
    6. install/verify IC policy
    7. run IC evaluator
    8. score graph-distance localization
    9. materialize repos at exact SHA
    10. compose full run
    11. add tracing
    12. add writer/optimization loop
    13. add CLI polish
    14. add GitHub automation

Do not ask agents to compose a full system before the seams are proven.

## Model-facing issue quality

Write issues as if the implementer is capable but literal.

Agents are good at:

- filling typed slots
- following explicit constraints
- implementing bounded workflows
- reacting to tests
- preserving named invariants

Agents are weaker at:

- inferring hidden architecture
- knowing which old code to preserve
- resisting over-abstraction
- understanding unstated non-goals
- stopping when scope expands

So issues should be explicit about boundaries.

## Bad issue pattern

Avoid:

    Port the Python SearchBench backend system to Go.

Problems:

- too broad
- unclear preservation target
- encourages mechanical copying
- no package boundaries
- no failure behavior
- no non-goals
- hard to review

## Good issue pattern

Prefer:

    Implement harness-owned policy installation for IC MCP sessions

    Goal:
      Ensure every IC evaluator run uses an explicit PolicyArtifact.

    Preserve:
      writer/evaluator isolation
      policy hash attribution
      fail-closed evaluation

    Do not preserve:
      worktree mutation
      evaluator-selected policy
      Python backend adapter shape

    Phase flow:
      start_mcp_session
      install_policy
      verify_policy
      expose_evaluator_tools
      close_session

    Acceptance:
      - evaluator never sees install_policy
      - missing policy fails closed
      - active policy hash is verified
      - result records policy_id and policy_hash

## Final rule

A good SearchBench-Go issue should make the desired implementation feel obvious without writing the implementation for the agent.

Use prose for intent.
Use phase pseudocode for lifecycle.
Use acceptance criteria for review.
Use non-goals to prevent architecture drift.
