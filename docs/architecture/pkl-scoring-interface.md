# Pkl Scoring Interface

## Status

This document defines the visible Pkl-based objective scoring interface in SearchBench-Go.

SearchBench-Go now has a narrow executable scoring seam in `internal/adapters/scoring/pkl`:

- Go supplies `score.ScoreEvidenceDocument` inputs
- a Pkl objective file is evaluated with `pkl-go`
- Go validates the returned `score.ObjectiveResult`

This still does not add parent-run discovery, experiment execution, or optimizer/runtime configuration ownership to Pkl.

## Goals

SearchBench-Go already has the pure Go types needed to support a visible scoring layer:

- raw score evidence in `internal/pure/score.ScoreEvidenceDocument`
- typed objective results in `internal/pure/score.ObjectiveResult`
- immutable bundle persistence for `score.pkl` and optional `objective.json`

The missing piece is the contract between those two layers. That contract must make scoring math durable, reviewable, and file-native instead of hiding it behind Go reducer names or transform enums.

The intended flow is:

1. Go computes raw evidence from a comparison report.
2. Pkl reads explicit evidence inputs and computes named objective values.
3. Go validates the objective result.
4. Go persists the objective result into the immutable bundle.

## Non-Goals

This interface does not define:

- experiment manifests
- dataset configuration
- evaluator, backend, writer, or optimizer configuration in Pkl
- parent bundle discovery or lineage loading
- content-addressed references
- graph-distance or token-efficiency reducer changes
- MCP, IC, jCodeMunch, tree-sitter, materialization, or tracing integration

## Existing Pure Models

The current pure Go boundary is:

- `internal/pure/score/evidence.go`
  - `ScoreEvidenceDocument`
- `internal/pure/score/objective.go`
  - `ObjectiveResult`
  - `ObjectiveValue`
  - `ObjectiveEvidenceRef`
  - `ObjectiveBounds`
- `internal/pure/report/evidence.go`
  - `ProjectScoreEvidence(report.CandidateReport)`

Those types stay authoritative. Pkl is a future producer of objective values, not a replacement for Go-owned evidence or validation.

## Lifecycle

The scoring lifecycle is:

1. `report.CandidateReport` is projected into `score.ScoreEvidenceDocument`.
2. Artifact writing persists that evidence as `score.pkl`.
3. A scoring runner resolves explicit evidence references for:
   - the current run
   - an optional parent run
4. Pkl evaluates pure scoring math over those resolved evidence documents.
5. Pkl returns named values plus explicit final selection.
6. Go validates the result with `ObjectiveResult.Validate()`.
7. Artifact writing persists `objective.json` when validation succeeds.

In the current manifest-driven local run, this happens immediately after the
comparison report is projected into score evidence and before bundle
finalization. Invalid objective output or a missing objective file fails the run
before a completed bundle is produced.

The canonical resolved experiment semantics that supply objective path,
current-parent evidence refs, and output settings now live in
`internal/app/experiment`. The scorer still remains a narrow adapter-edge seam.

The scoring runner is intentionally narrow. It evaluates formulas over evidence. It does not own execution orchestration, bundle mutation, or system configuration.

The current implementation imports explicit `score.pkl` evidence modules directly, converts them to `Dynamic`, and evaluates the objective file against that Pkl-native shape.

## Evidence Inputs

Pkl should only read explicit, durable evidence inputs. The first-class source is bundled `score.pkl`, which already exposes field-addressable evidence such as:

- `current.metrics`
- `current.localizationDistance`
- `current.usage`
- `current.regressions`
- `current.invalidPredictions`
- `current.promotion_decision`

If parent-aware scoring is added later, the parent input must also be explicit. It must not be discovered implicitly from lineage or mutable global state.

A future runner should resolve inputs into a shape conceptually equivalent to:

- `current`
  - current run `ScoreEvidenceDocument`
- `parent`
  - optional parent `ScoreEvidenceDocument`
- `constants`
  - scorer-owned constants from the scoring file

Pkl must not read `report.json` tables or Go-specific reducer internals when `score.pkl` already exposes the needed evidence.

## Evidence References

Current and parent evidence references must remain explicit and typed in Go through `score.ObjectiveEvidenceRef`.

The scoring boundary should treat these refs as the reviewable declaration of what evidence was used. Typical names are:

- `current`
- `parent`

Each ref should identify the durable source by path and, when available later, a digest:

- `bundle_path`
- `score_path`
- `report_path`
- `sha256`

The important design point is that previous evidence is threaded explicitly
through these refs. It should become part of the Pkl value graph through
`current`, `parent`, `currentRef`, and `parentRef`, not through hidden Go
optimizer loop state.

This issue does not add hash-based ref discovery. It only fixes the contract
that a future implementation must honor.

## Pkl Output Contract

Pkl should emit named numeric values, not hidden reducer ids. The output contract maps directly onto `score.ObjectiveResult`.

Expected shape:

- `objective_id`
  - stable scorer identifier such as `localization-v1`
- `evidence_refs`
  - explicit refs for `current` and optional `parent`
- `values`
  - every named intermediate, penalty, and final value
- `final`
  - the name of the final selected value
- `bounds`
  - optional numeric guardrails for the final value

This keeps objective math legible. A reviewer can read `objective.json` and see:

- which evidence was used
- which named values were computed
- which value is final

without reverse-engineering Go code.

## Named Value Semantics

Pkl should expose every reviewable intermediate directly as a named value. For example:

- `currentLocalizationQuality`
- `parentLocalizationQuality`
- `improvementVsParent`
- `tokenEfficiency`
- `base`
- `regressionPenalty`
- `invalidPredictionPenalty`
- `final`

These names map 1:1 onto `score.ObjectiveValue.Name`.

Kinds map onto `score.ObjectiveValueKind`:

- `intermediate`
  - ordinary derived values
- `penalty`
  - multiplicative or subtractive penalties
- `final`
  - the explicit final score

The final score must remain explicit and must be referenced by name through `ObjectiveResult.Final`.

Visible objective files should stay compact. Shared Pkl helpers may construct
`ObjectiveValue` entries so the scoring file reads like objective math plus a
small export block, for example:

- `helpers.intermediate("base", base)`
- `helpers.penalty("regressionPenalty", regressionPenalty)`
- `helpers.finalValue("final", finalScore)`

Those helpers are only constructor conveniences. They do not replace the
authoritative final selector, which remains `final = "final"`.

## Example Formula Style

The visible scoring layer supports formulas of the form:

`currentLocalizationQuality = 1.0 - current.localizationDistance.goldHop.candidate / maxHop`

`tokenEfficiency = 1.0 - current.usage.totalTokens / tokenBudget`

`final = base * regressionPenalty * invalidPredictionPenalty`

The important property is not the specific math. It is that the formula is durable and human-readable in a file instead of being encoded as opaque reducer wiring in Go.

The current evidence document does not expose a canonical `localizationDistance.mean` field. A scorer must either:

- use the existing named evidence fields such as `goldHop` or `issueHop`, or
- add a new pure evidence field in a separate issue before depending on it in Pkl

## Go Validation Responsibilities

Go remains responsible for validating the result shape before persistence.

The current `ObjectiveResult.Validate()` contract already rejects:

- missing schema version
- missing objective id
- duplicate value names
- missing final value
- final value not present in `values`
- non-finite values
- duplicate or malformed evidence refs
- final values outside declared bounds

The current executable runner wraps those validation failures at the scoring boundary, but it still terminates by producing a validated `ObjectiveResult` or a typed failure.

## Bundle Integration

Immutable run bundles already persist:

- `resolved.json`
- `report.json`
- `score.pkl`
- optional `objective.json`
- `metadata.json`
- optional rendered report
- `COMPLETE`

- `score.pkl` remains the raw evidence input
- `objective.json` remains the validated scoring output

In the executable local path, `score.pkl` and `objective.json` have distinct
roles:

- `score.pkl`
  - the Pkl-readable evidence document projected from `report.json`
- `objective.json`
  - the evaluated and validated result of the configured `objective.pkl`

The scoring source file should also be copied into the bundle as a reviewable artifact so the exact visible math used for that run is immutable. A reasonable future shape is:

- `scoring/objective.pkl`

or

- `objective.pkl`

The current scorer does not yet copy the scoring source into bundles automatically. That remains a follow-up persistence concern.

## Failure Modes

The scorer should treat these as explicit failures, not silent fallbacks:

- Pkl runtime unavailable
- scoring file missing
- parse failure
- invalid evidence input shape
- missing required evidence refs
- invalid objective output structure
- duplicate named values
- malformed evidence refs
- non-finite final value
- final value outside bounds
- bundle serialization or write failure

Two rules are important:

1. Failed objective evaluation must not silently degrade into a default score.
2. Objective persistence must only happen after validation succeeds.

## Current Executable Scope

The manifest-driven scoring path currently supports:

- current-run evidence
- optional parent evidence when explicitly provided to the scoring adapter or
  manifest-driven run request seam
- local/fake comparison execution feeding `score.pkl`

The current score evidence may be materialized temporarily before final bundle
writing so the Pkl objective can import an explicit `score.pkl` module. That
materialization seam is intentionally small and reviewable:

- materialize current `score.pkl`
- produce the current `ObjectiveEvidenceRef`
- accept an optional explicit parent `ObjectiveEvidenceRef`
- pass both into Pkl evaluation
- clean up temporary current score files afterward

It intentionally does not yet implement:

- automatic parent bundle discovery
- lineage lookup
- content-addressed bundle ids
- typed Pkl evidence classes
- real backend execution

For future optimizer lineage, prefer:

- explicit artifact references
- Pkl `amends`
- immutable per-round bundles

before introducing any bespoke inheritance or content-addressed store.

## Dependency Direction

The dependency direction remains:

`score/domain/run`

`report projection`

`artifact serialization`

Pkl belongs between score evidence and objective result production. It does not replace those layers and must not become a second general runtime for SearchBench-Go.

## Implementation Boundary For Future Work

The current implementation adds that minimum seam:

1. resolve explicit evidence refs supplied by Go
2. load the scoring source file
3. run Pkl against score evidence inputs
4. map the output directly into `score.ObjectiveResult`
5. validate the result in Go

Bundle persistence of `objective.json` is already supported once a caller supplies the validated result. Automatic scoring-source copying remains future work.
