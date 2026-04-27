You are implementing GitHub issue Dangling: clarify bounded evaluator-agent execution semantic #8.

Before doing anything else, you MUST read the issue directly using the GitHub CLI:

gh issue view #8 --comments

Do not rely only on this prompt. The GitHub issue body is the implementation contract. If anything here conflicts with the issue, stop and report the conflict.

After reading the issue, also read:

- AGENTS.md
- docs/engineering/agentic-development-flow.md
- docs/engineering/issue-style-guide.md
- docs/engineering/model-testing.md
- docs/engineering/pure-center.md

---

GOAL

Clarify evaluator execution semantics.

This is NOT a major implementation task. It is a clarification task across:

- documentation
- comments
- tests

You are making the intended behavior of the evaluator explicit.

---

CORE IDEA

The evaluator is a bounded agent loop, NOT a single model call.

One evaluator Run(ctx, task) may include:

- multiple model turns
- multiple tool calls
- multiple tool results
- internal reasoning handled by Eino
- one final structured prediction

The evaluator must still:

- be bounded
- be deterministic at the harness level
- return exactly one final prediction
- remain isolated from writer/repair concerns

---

CRITICAL CONSTRAINT

The evaluator must NOT run any CLI validation or writer pipeline behavior.

Specifically, it must NOT run:

- go test
- gofmt
- templ generate
- ruff
- pyright
- pytest
- candidate validation
- writer repair loops
- pipeline classification

That logic belongs to writer/repair systems, not the evaluator.

---

IMPORTANT DISTINCTIONS

Make the following distinctions clear in docs/comments/tests:

Evaluator run:
- one attempt to solve one task

Evaluator model turn:
- one model invocation inside a run

Evaluator tool call:
- one tool invocation requested by the model

Evaluator retry attempt:
- a new evaluator run after failure (malformed output, empty prediction, recoverable error)

Writer pipeline attempt:
- candidate generation + CLI validation + repair loop

These are NOT the same and must not be conflated.

---

EXPECTED CHANGES

You should NOT rewrite the evaluator.

Instead, make targeted updates in:

- internal/executor/eino/evaluator.go
- internal/executor/eino/phases.go
- internal/executor/eino/retry.go
- internal/executor/eino/errors.go
- internal/executor/eino/evaluator_test.go
- internal/executor/eino/retry_test.go
- docs/engineering/* (where appropriate)

Add or refine:

1. Code comments explaining evaluator loop semantics
2. Documentation clarifying bounded agent behavior
3. Tests that assert the intended behavior

---

REQUIRED SEMANTIC CLARIFICATIONS

Make it clear that:

1. A single evaluator Run:
   - may involve multiple model turns
   - may involve multiple tool calls
   - is bounded by configuration
   - returns exactly one Prediction

2. Evaluator retries:
   - are separate from model/tool turns
   - represent a new attempt, not continuation

3. Evaluator execution:
   - does NOT include writer pipeline behavior
   - does NOT execute CLI validation

4. Eino:
   - owns the internal model/tool loop

5. The harness:
   - owns bounds, retries, and finalization

---

BOUNDS (DOCUMENT, DO NOT OVER-ENGINEER)

Document that evaluator execution should be bounded by things like:

- max model turns
- max tool calls
- max duration
- max output tokens
- context cancellation

Do NOT introduce a large new config system.

If config already exists, reference it.
If not, document expectations without overbuilding.

---

TEST REQUIREMENTS

Add or update tests to demonstrate:

1. Evaluator is allowed to make multiple tool calls within a single Run
2. Evaluator is allowed to have multiple model turns (mocked if needed)
3. Evaluator still returns exactly one final Prediction
4. Evaluator retries are distinct from internal model/tool turns
5. Evaluator does NOT invoke any pipeline or command execution
6. Tests remain deterministic and offline

You may use mocks/fakes to simulate multi-turn + tool-call behavior.

---

NON-GOALS

Do NOT implement:

- writer agent
- repair loop
- CLI validation pipeline
- MCP wiring
- new tool systems
- new state machine frameworks
- large configuration systems

This issue is about CLARITY, not expansion.

---

VALIDATION

Run:

go test ./...

Also:

go test ./internal/executor/eino

All tests must pass.

All tests must remain offline and deterministic.

---

FINAL RESPONSE FORMAT

Summarize:

- files changed
- what clarifications were added
- what tests were added or updated
- how evaluator semantics are now clearly defined
- any deviations from the issue contract
