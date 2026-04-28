# Pkl Experiment Manifests

## Status

SearchBench-Go uses Pkl here as a human-authored experiment surface, not as the source of truth for runtime semantics.

The intended flow is:

1. a human writes `experiment.pkl`
2. `pkl-go` resolves the manifest into typed Go config structs via `internal/adapters/config/pkl`
3. Go validates SearchBench-specific cross-field rules
4. future code projects that validated config into pure/app/adapter inputs

This does not make Pkl drive full SearchBench execution yet.

## Boundary

Go owns:

- pure domain models
- task identity
- run lifecycle
- evaluator execution
- report and score models
- backend and pipeline semantics
- SearchBench-specific validation

Pkl owns:

- a typed, readable manifest surface
- defaults
- simple local constraints
- manifest composition

Pkl is a surface. Go remains the source of truth.

## Manifest Shape

The first manifest shape is intentionally small:

- `name`
- `mode`
- `dataset`
- `systems`
- `evaluator`
- optional `writer`
- `scoring`
- `outputConfig`

`outputConfig` is spelled that way because `pkl-go` code generation cannot use a top-level module property literally named `output` without colliding with Pkl's built-in module `output` property.

The manifest answers:

- what run is this
- what dataset and split are used
- what systems are being compared
- what evaluator model and bounds apply
- whether writer mode is enabled
- what writer pipeline, if any, is configured
- which scoring objective file is referenced
- what output and tracing preferences are requested

The local example is intentionally self-contained under one experiment folder:

- `configs/experiments/local-ic-vs-jcodemunch/experiment.pkl`
- `configs/experiments/local-ic-vs-jcodemunch/scoring/localization-objective.pkl`
- `configs/experiments/local-ic-vs-jcodemunch/policies/candidate_policy.py`
- `configs/experiments/local-ic-vs-jcodemunch/artifacts/runs/`

That keeps the manifest, local scoring file, local policy artifact, and local bundle root together instead of spreading them across the repo with `../` path traversal.

## Evaluator And Writer Separation

The evaluator owns bounded model/tool execution only.

The writer owns candidate-attempt behavior and optional CLI validation pipeline configuration.

That separation is explicit in the manifest shape:

- `evaluator` has model, bounds, and retry settings
- `writer` optionally has model, attempts, and `pipeline`

There is deliberately no `pipeline` block under `evaluator`.

## Scoring Boundary

`scoring.objective` points to the separate visible scoring file.

The experiment manifest does not define objective math.

The scoring objective file answers:

- how bundled `score.pkl` evidence is turned into named objective values
- which value is final

That scoring behavior remains separate from experiment configuration.

## Current Scope

This first manifest surface proves:

- SearchBench can carry typed Pkl experiment files
- `pkl-go` can resolve them into Go structs
- Go can validate SearchBench-specific invariants afterward
- the manifest surface stays in `internal/adapters/config/pkl` instead of leaking Pkl runtime concerns into pure packages

It does not yet implement:

- full execution from manifests
- objective Pkl execution
- backend session construction
- policy installation
- writer optimization execution
- lineage or parent-run resolution
