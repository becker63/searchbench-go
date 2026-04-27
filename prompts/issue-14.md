You are implementing GitHub issue Define Eino callbacks as an explicit execution boundary.

Before doing anything else, you MUST read the issue directly using the GitHub CLI:

gh issue view #14 --comments

Do not rely only on this prompt. The GitHub issue body is the implementation contract. If anything here conflicts with the issue, stop and report the conflict.

After reading the issue, also read:

- AGENTS.md
- docs/engineering/agentic-development-flow.md
- docs/engineering/issue-style-guide.md
- docs/engineering/model-testing.md
- docs/engineering/pure-center.md

---

GOAL

Introduce an explicit Eino callback boundary in the repository.

This is a STRUCTURAL + TESTING task, not a feature implementation.

You are:

- defining where callbacks live
- defining how they are composed into evaluator execution
- proving the seam works with a fake/test callback

You are NOT implementing real instrumentation.

---

CORE IDEA

Callbacks are an execution-layer boundary owned by Eino.

SearchBench-Go must make this boundary:

- explicit
- visible in directory structure
- composable
- independent from evaluator business logic

The evaluator should receive callbacks, not define them internally.

---

ARCHITECTURAL RULE

SearchBench-Go owns:

- repository structure
- evaluator lifecycle boundaries
- callback composition
- typed results and failures

Eino owns:

- callback interfaces
- callback lifecycle
- invocation during execution

Callbacks must:

- live near Eino execution code
- NOT live in domain/scoring packages
- NOT be scattered through evaluator logic

---

REQUIRED STRUCTURE

Create a clear callback boundary under the Eino executor.

Preferred shape:

internal/implementation/executor/eino/
  evaluator.go

internal/implementation/executor/eino/callbacks/
  callbacks.go
  fake_test_callback.go
  callbacks_test.go

Exact names can vary slightly, but:

- the boundary must be obvious
- callbacks must be isolated
- evaluator imports from this boundary

---

EVALUATOR FLOW (UPDATED)

RunEvaluator(ctx, task)

  phase: prepare_callbacks
    construct callback set
    include fake/test callback when configured
    MUST be cold (no model calls)

    if failure:
      return callback_setup_failed

  phase: run_evaluator
    execute through Eino with callbacks

    if failure:
      return evaluator_failed

  phase: finalize_prediction
    parse and validate prediction

  phase: complete
    return result

Callbacks must ONLY observe execution, not control it.

---

WHAT TO IMPLEMENT

1. Callback composition surface

Create a minimal function like:

- BuildCallbacks(config) -> []EinoCallback

This should:

- assemble callbacks
- allow test injection
- remain simple (no framework)

2. Fake/test callback

Implement a fake callback that can record:

- constructed
- attached
- observed model start/end
- observed tool call (if available)

This must:

- be test-oriented
- not become a production abstraction
- not introduce domain types

3. Wire into evaluator

- evaluator must call callback builder in prepare_callbacks phase
- evaluator must pass callbacks into Eino execution
- evaluator must NOT contain callback logic inline

4. Typed failure

Add or reuse a failure type:

- callback_setup_failed

Ensure evaluator returns it when callback construction fails.

---

TEST REQUIREMENTS

Add tests that prove:

1. Callback construction is cold
   - no model/tool execution during setup

2. Evaluator can run with callbacks attached

3. Fake callback observes execution events

4. Callback setup failure returns callback_setup_failed

5. Evaluator runs correctly with:
   - no callbacks
   - fake callbacks enabled

6. Callback package:
   - does NOT import scoring packages
   - does NOT import tracing systems
   - does NOT pull in heavy domain logic

7. Fake callback:
   - remains test-scoped or clearly marked as fixture

Tests must be:

- deterministic
- offline
- behavior-focused (not implementation-detail fragile)

---

STRICT NON-GOALS

Do NOT implement:

- token counting
- token usage tracking
- LangSmith or Lanesmith
- tracing systems
- scoring
- cost tracking
- usage records
- MCP-specific callbacks
- writer/optimizer callbacks
- custom callback frameworks

Do NOT:

- move domain models into callbacks
- introduce callback-owned data models
- require callbacks for evaluator execution

---

DESIGN CONSTRAINTS

- Callback construction must be COLD
- Callbacks must be COMPOSABLE peers
- No callback should own another
- Evaluator must work WITHOUT callbacks
- No global state
- No hidden wiring

Avoid:

- over-engineering
- introducing interfaces that mirror Eino
- building a framework instead of a seam

---

FAILURE BEHAVIOR

You must distinguish:

- callback_setup_failed
- evaluator_failed
- prediction_finalization_failed
- unexpected_internal_failure

If callback setup fails:

- evaluator must fail closed
- must not proceed to execution

Fake callback failures must be visible in tests.

---

VALIDATION

Run:

go test ./...

Also:

go test ./internal/implementation/executor/eino

All tests must pass.

All tests must remain offline and deterministic.

---

FINAL RESPONSE FORMAT

Summarize:

- files added/changed
- where the callback boundary lives
- how callbacks are composed
- how evaluator integrates callbacks
- what tests prove the seam works
- how cold construction is enforced
- confirmation that no non-goals were implemented
- any deviations from the issue contract
