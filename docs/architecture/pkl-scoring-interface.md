# Pkl Scoring Interface

## Status

This document defines the future interface for visible Pkl-based objective scoring in SearchBench-Go.

It is intentionally design-only. It does not add executable Pkl evaluation, bundle loading, parent resolution, experiment manifests, or runtime configuration.

## Goals

SearchBench-Go already has the pure Go types needed to support a visible scoring layer:

- raw score evidence in `internal/score.ScoreEvidenceDocument`
- typed objective results in `internal/score.ObjectiveResult`
- immutable bundle persistence for `score.json` and optional `objective.json`

The missing piece is the contract between those two layers. That contract must make scoring math durable, reviewable, and file-native instead of hiding it behind Go reducer names or transform enums.

The intended flow is:

1. Go computes raw evidence from a comparison report.
2. Pkl reads explicit evidence inputs and computes named objective values.
3. Go validates the objective result.
4. Go persists the objective result into the immutable bundle.

## Non-Goals

This interface does not define:

- executable Pkl integration
- experiment manifests
- dataset configuration
- evaluator, backend, writer, or optimizer configuration in Pkl
- parent bundle discovery or lineage loading
- content-addressed references
- graph-distance or token-efficiency reducer changes
- MCP, IC, jCodeMunch, tree-sitter, materialization, or tracing integration

## Existing Pure Models

The current pure Go boundary is:

- `internal/score/evidence.go`
  - `ScoreEvidenceDocument`
- `internal/score/objective.go`
  - `ObjectiveResult`
  - `ObjectiveValue`
  - `ObjectiveEvidenceRef`
  - `ObjectiveBounds`
- `internal/report/evidence.go`
  - `ProjectScoreEvidence(report.CandidateReport)`

Those types stay authoritative. Pkl is a future producer of objective values, not a replacement for Go-owned evidence or validation.

## Lifecycle

The future scoring lifecycle is:

1. `report.CandidateReport` is projected into `score.ScoreEvidenceDocument`.
2. Artifact writing persists that evidence as `score.json`.
3. A future scoring runner resolves explicit evidence references for:
   - the current run
   - an optional parent run
4. Pkl evaluates pure scoring math over those resolved evidence documents.
5. Pkl returns named values plus explicit final selection.
6. Go validates the result with `ObjectiveResult.Validate()`.
7. Artifact writing persists `objective.json` when validation succeeds.

The scoring runner is intentionally narrow. It evaluates formulas over evidence. It does not own execution orchestration, bundle mutation, or system configuration.

## Evidence Inputs

Pkl should only read explicit, durable evidence inputs. The first-class source is bundled `score.json`, which already exposes field-addressable evidence such as:

- `current.metrics`
- `current.localizationDistance`
- `current.usage`
- `current.regressions`
- `current.invalidPredictions`
- `current.promotionDecision`

If parent-aware scoring is added later, the parent input must also be explicit. It must not be discovered implicitly from lineage or mutable global state.

A future runner should resolve inputs into a shape conceptually equivalent to:

- `current`
  - current run `ScoreEvidenceDocument`
- `parent`
  - optional parent `ScoreEvidenceDocument`
- `constants`
  - scorer-owned constants from the scoring file

Pkl must not read `report.json` tables or Go-specific reducer internals when `score.json` already exposes the needed evidence.

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

This issue does not add ref resolution. It only fixes the contract that a future implementation must honor.

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

## Example Formula Style

The visible scoring layer should support formulas of the form:

`currentLocalizationQuality = 1.0 - min(current.localizationDistance.goldHop.candidate, constants.maxHop) / constants.maxHop`

`tokenEfficiency = 1.0 - min(current.usage.totalTokens, constants.tokenBudget) / constants.tokenBudget`

`final = base * regressionPenalty * invalidPredictionPenalty`

The important property is not the specific math. It is that the formula is durable and human-readable in a file instead of being encoded as opaque reducer wiring in Go.

The current evidence document does not expose a canonical `localizationDistance.mean` field. A future scorer must either:

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

Future executable Pkl support should add runner-level failures around that validation, but it must still terminate by producing a validated `ObjectiveResult` or a typed failure.

## Bundle Integration

Immutable run bundles already persist:

- `resolved.json`
- `report.json`
- `score.json`
- optional `objective.json`
- `metadata.json`
- optional rendered report
- `COMPLETE`

When future executable scoring is added:

- `score.json` remains the raw evidence input
- `objective.json` remains the validated scoring output

The scoring source file should also be copied into the bundle as a reviewable artifact so the exact visible math used for that run is immutable. A reasonable future shape is:

- `scoring/objective.pkl`

or

- `objective.pkl`

This issue does not implement that artifact. It only makes the requirement explicit so executable support does not hide scoring source outside the bundle.

## Failure Modes

The future runner should treat these as explicit failures, not silent fallbacks:

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

## Dependency Direction

The dependency direction remains:

`score/domain/run`

`report projection`

`artifact serialization`

Pkl belongs between score evidence and objective result production. It does not replace those layers and must not become a second general runtime for SearchBench-Go.

## Implementation Boundary For Future Work

A future implementation should add only the minimum executable seam needed:

1. resolve explicit evidence refs
2. load the scoring source file
3. run Pkl against score evidence inputs
4. map named output values into `score.ObjectiveResult`
5. validate the result in Go
6. persist `objective.json` and copied scoring source into the bundle

That implementation should remain fixture-driven and scoring-specific. It should not expand into a generic manifest or plugin framework unless a later issue explicitly requires it.
