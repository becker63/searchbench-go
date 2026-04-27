You are implementing GitHub issue Implement harness-owned token usage accounting as an Eino callback #15.

Before doing anything else, you MUST read the issue directly using the GitHub CLI:

gh issue view #15 --comments

Do not rely only on this prompt. The GitHub issue body is the implementation contract. If anything here conflicts with the issue, stop and report the conflict.

After reading the issue, also read:

- AGENTS.md
- docs/engineering/agentic-development-flow.md
- docs/engineering/issue-style-guide.md
- docs/engineering/model-testing.md
- docs/engineering/pure-center.md

---

GOAL

Implement harness-owned token usage accounting as a concrete Eino callback.

This builds on the existing callback boundary.

After this is complete:

- evaluator runs produce token usage records
- usage is owned by the harness (NOT tracing)
- usage is attached to evaluator results
- usage works without LangSmith/Lanesmith

---

CORE IDEA

Token usage is:

- collected via an Eino callback
- normalized into harness-owned records
- attached to evaluator results
- independent of tracing systems

Tracing is a sink.
Usage records are the source of truth.

---

ARCHITECTURAL RULE

SearchBench-Go owns:

- normalized usage records
- usage summaries
- scoring inputs (future)
- token accounting logic
- fallback estimation

Eino owns:

- execution
- callback lifecycle
- provider usage metadata

Tracing systems own:

- display/export only

You MUST NOT:

- read usage from tracing
- depend on LangSmith/Lanesmith
- leak provider-specific types into domain models
- scatter token fields across evaluator code

---

REQUIRED IMPLEMENTATION

1. Usage callback (execution layer)

Inside existing callback boundary:

internal/implementation/executor/eino/callbacks/

Add:

- usage.go
- usage_test.go

This callback must:

- observe model input/output
- capture provider-reported usage if present
- pass raw data to a collector

It must NOT:

- own domain models
- perform scoring
- depend on tracing

---

2. Usage domain (harness-owned)

Create a usage package:

internal/usage/

With:

- record.go      (per-call usage record)
- summary.go     (aggregated run summary)
- collector.go   (collects and normalizes usage)
- tokenizer.go   (local estimation)

This layer must:

- be provider-neutral
- not depend on Eino types
- not depend on tracing
- define canonical usage structures

---

3. Usage record requirements

Each model call should produce a record with:

- model/provider (if known)
- input tokens (reported and/or estimated)
- output tokens (reported and/or estimated)
- total tokens (if available)
- source:
  - reported
  - estimated
  - mixed
  - unavailable

Do NOT expose raw provider blobs.

---

4. Usage collector

The collector must:

- accept callback events
- aggregate per-call records
- normalize usage
- handle missing provider data
- fall back to local estimation
- track failures (tokenizer failure, missing data, etc.)

Collector must be reusable later.

---

5. Evaluator integration

Update evaluator flow:

RunEvaluator(ctx, task)

  phase: prepare_callbacks
    include usage callback

  phase: prepare_usage_accounting
    create usage collector
    wire collector into callback

    if failure:
      return usage_accounting_setup_failed

  phase: run_evaluator
    callbacks record usage per model call

  phase: finalize_usage
    normalize usage records
    attach summary to evaluator result

  phase: finalize_prediction
    attach prediction + usage

Evaluator must:

- not depend on tracing
- not compute tokens itself
- only attach finalized usage results

---

6. Failure handling

You must distinguish:

- callback_setup_failed
- usage_accounting_setup_failed
- evaluator_failed
- tokenizer_unavailable
- estimation_failed
- incomplete_usage
- prediction_finalization_failed

Rules:

- missing provider usage → fallback to estimate
- tokenizer failure → record incomplete usage
- usage failure → does NOT invalidate evaluator success
- setup failure → fail closed

---

TEST REQUIREMENTS

Add tests proving:

1. usage works without tracing enabled
2. callback setup is cold (no model calls)
3. provider usage → reported record
4. missing provider usage → estimated record
5. tokenizer failure → incomplete usage (not zero)
6. usage records attach to evaluator result
7. usage summary is available on result
8. callback setup failure → typed failure
9. usage setup failure → typed failure
10. usage package does NOT import tracing
11. usage package does NOT depend on Eino types
12. usage + tracing callbacks compose cleanly

Tests must be:

- deterministic
- offline
- behavior-focused

Avoid asserting exact token counts unless tokenizer is fixed.

---

STRICT NON-GOALS

Do NOT implement:

- token-efficiency scoring
- graph scoring
- LangSmith/Lanesmith integration
- cost/billing logic
- CLI reporting
- writer/optimizer usage
- dataset aggregation
- custom callback framework

Do NOT:

- read usage from tracing
- make tracing required
- move usage into tracing adapters

---

DESIGN CONSTRAINTS

- usage must be harness-owned
- callback must be execution-layer only
- usage records must be normalized
- evaluator must work without usage (if disabled)
- no global state
- no hidden coupling

---

VALIDATION

Run:

go test ./...

Also:

go test ./internal/implementation/executor/eino ./internal/usage

All tests must pass.

All tests must remain offline and deterministic.

---

FINAL RESPONSE FORMAT

Summarize:

- files added/changed
- usage callback design
- usage record + summary structure
- how evaluator integrates usage
- how fallback estimation works
- what tests prove correctness
- confirmation that tracing is not required
- confirmation that no non-goals were implemented
- any deviations from the issue contract
