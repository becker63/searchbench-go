# Expanding the Pure Center

**Docs hub:** [Documentation index](../README.md) · [Package boundaries](../architecture/package-boundaries.md) · [Architecture](../architecture/architecture.md)

SearchBench-Go keeps its stable product model in deterministic packages and pushes world-touching behavior to adapter edges.

The product model is:

```text
Game -> Round -> Match -> Evidence -> Decision -> NextChallenger
```

## Pure Center

The pure center contains typed concepts that should be readable without knowing Eino, MCP, provider SDKs, LangSmith, filesystem layout, or CLI behavior.

Current pure packages include:

```text
internal/pure/game
internal/pure/round
internal/pure/match
internal/pure/domain
internal/pure/policy
internal/pure/execution
internal/pure/score
internal/pure/report
internal/pure/optimizer
internal/pure/codegraph
internal/pure/usage
```

Evaluator and optimizer prompts render from `internal/agents/*/prompt`; they intentionally stay out of `internal/pure` so the pure layer never pulls Eino.

These packages answer stable questions:

```text
What game is being played?
What round is being evaluated?
What match is one dataset item?
What policy executed the match?
What evidence was produced?
What decision was recorded?
What next challenger was proposed?
```

## Adapter Edge

Effects still exist, but they should not own SearchBench vocabulary.

Effectful concerns belong at the edge:

```text
Pkl config loading
Pkl objective execution
Eino-backed model execution
filesystem bundle writing
subprocess pipeline execution
future MCP/provider/tracing integrations
```

The dependency direction is:

```text
surface / app / adapters
    ->
ports
    ->
pure
```

The pure center must not import adapters, surfaces, provider SDKs, tracing SDKs, or subprocess execution.

## Why This Matters

A pure center lets most behavior be tested without API keys, network, live model calls, repo materialization, tracing backends, or mutable filesystem state.

Effects should be converted into typed data quickly:

```text
external command output -> pipeline.StepResult
model/tool observation   -> execution.ExecutedRun
round report             -> score.RoundEvidenceDocument
round evidence           -> score.ObjectiveResult
```

Once effects are converted, scoring, reporting, decisions, prompt rendering, and bundle metadata stay deterministic.

## Current Story

A newcomer should be able to form this story from pure/app code:

```text
A Game defines review rules.
A Round compares an incumbent policy and a challenger policy.
A Match is one dataset item inside that round.
Each policy produces a match execution.
Score and report code produce round evidence.
A Decision records whether the challenger advances.
Optimizer code may propose a NextChallenger.
```

Dataset adapters may still name upstream records with the source dataset's own vocabulary, but SearchBench-facing packages should project those records into matches before they cross inward.

## Boundary Pressure

Watch for these signs that a boundary sweep is needed:

```text
adapter DTOs appearing in pure packages
provider details leaking into execution records
prompt inputs importing executor packages
filesystem paths becoming core identity
tracing SDK types appearing in reports or evidence
parallel report/score/match models inside adapters
```

Not every boundary issue should be fixed immediately. Capture future cleanup pressure as a follow-up issue when fixing it would derail the current change.

## Practical Rule

When adding code, ask whether the new concept is stable SearchBench meaning or an effectful mechanism.

Stable meaning belongs near the center:

```text
Game
Round
Match
Policy
Execution
Evidence
Decision
NextChallenger
ObjectiveResult
PipelineClassification
```

Effectful mechanisms belong at the edge:

```text
Eino model call
Pkl runtime evaluation
MCP client session
subprocess execution
repo checkout
LangSmith span
OpenAI request
filesystem write
```

Keep expanding the part of the system that can be understood, tested, and changed without the world being present.
