# Pkl Experiment Manifests

## Status

SearchBench-Go uses Pkl here as a human-authored experiment surface, not as the source of truth for runtime semantics.

The intended flow is:

1. a human writes `experiment.pkl`
2. `pkl-go` resolves the manifest into typed Go config structs via `internal/adapters/config/pkl`
3. Go validates SearchBench-specific cross-field rules
4. `internal/app/localrun` resolves the validated manifest into an execution plan with stable absolute paths
5. the current local/fake compare path executes from that resolved plan
6. a durable bundle is written with resolved inputs, report artifacts, score evidence, and objective output

This now makes the current local/fake SearchBench path executable from a Pkl
manifest. It still does not make Pkl drive full SearchBench execution semantics
or future backend integrations.

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

## Manifest Immutability

Experiment manifests are immutable run inputs. SearchBench-Go should not mutate
an existing `experiment.pkl` after a run completes.

If a later optimization round needs different configuration or explicit prior
evidence, that should be expressed as a new Pkl module and a new artifact
bundle, not by rewriting the previous manifest in place.

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

## Running A Manifest

The current executable path is:

```text
searchbench run --manifest configs/experiments/local-ic-vs-jcodemunch/experiment.pkl
```

That command currently uses the safe local/fake comparison seam in
`internal/app/localrun`. It does not call real IC MCP, jCodeMunch MCP, or
provider APIs in tests.

## Lineage Direction

The current direction for optimization lineage is:

- optimization history should be artifact history
- previous evidence should be explicit in the manifest/scoring seam
- previous evidence should become part of the Pkl value graph
- Pkl `amends` plus direct artifact/module references is the preferred first
  mechanism for round-to-round lineage

This means a future optimizer round should more naturally look like:

- a new Pkl module
- optionally `amends` a shared schema or prior-round shape
- explicitly references prior `score.pkl` or related artifacts
- produces a new immutable bundle

It should not rely on hidden Go optimizer state to thread previous score
evidence, and it should not start with a bespoke content-addressed lookup
system.

## Manifest-Relative Paths

These manifest fields resolve relative to the directory containing
`experiment.pkl` unless they are already absolute:

- `scoring.objective`
- `systems.baseline.policy.path`, when present later
- `systems.candidate.policy.path`
- `outputConfig.bundleRoot`

CLI-only overrides are separate from manifest-relative resolution. For example,
`searchbench run --bundle-root ...` resolves from the caller environment rather
than from the manifest directory.

## Produced Artifacts

The manifest-driven local run writes an immutable bundle under the resolved
bundle root:

- `resolved.json`
  - resolved manifest projection used for execution
- `report.json`
  - structured comparison report
- `score.pkl`
  - Pkl-readable score evidence projected from `report.json`
- `objective.json`
  - validated result of evaluating the configured objective against `score.pkl`
- `metadata.json`
  - deterministic artifact inventory
- `report.txt` or `report.md`
  - optional human-readable rendering

## Current Scope

The current executable path intentionally remains local/fake-only for runtime
execution. It does not yet implement:

- real IC MCP execution
- real jCodeMunch MCP execution
- repo materialization
- tree-sitter indexing
- writer optimization execution
- lineage or parent-run discovery by hash
- tracing export

Content hashes may still remain useful later as integrity and reproducibility
metadata, but they are not the primary lineage mechanism in this first
implementation direction.
