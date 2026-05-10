# Pkl Round Manifests

## Status

SearchBench-Go uses Pkl here as a human-authored round surface, not as the source of truth for runtime semantics.

The intended flow is:

1. a human writes `round.pkl`
2. `pkl-go` resolves the manifest into typed Go config structs via `internal/adapters/config/pkl`
3. Go validates SearchBench-specific cross-field rules
4. `internal/app/evaluation` resolves the validated manifest into the canonical app-level round plan
5. an execution strategy such as `internal/app/round` executes that resolved plan
6. a durable bundle is written with resolved inputs, report artifacts, round evidence, and objective output

This now makes the current local/fake SearchBench path executable from a Pkl
manifest. It still does not make Pkl drive full SearchBench execution semantics
or future backend integrations.

Pkl is the canonical public round interface. Generated Pkl structs remain
adapter-edge data and are projected into an app-owned resolved round model
before execution starts.

## Boundary

Go owns:

- pure domain models
- match identity
- execution lifecycle
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

Round manifests are immutable round inputs. SearchBench-Go should not mutate
an existing `round.pkl` after a run completes.

If a later optimization round needs different configuration or explicit prior
evidence, that should be expressed as a new Pkl module and a new artifact
bundle, not by rewriting the previous manifest in place.

## Manifest Shape

The first manifest shape is intentionally small:

- `name`
- `mode`
- `dataset`
- `policies`
- `evaluator`
- `scoring`

`outputConfig` is spelled that way because `pkl-go` code generation cannot use a top-level module property literally named `output` without colliding with Pkl's built-in module `output` property.

The manifest answers:

- what round is this
- what dataset and split are used
- what policies are being compared
- what evaluator model and bounds apply
- which scoring objective file is referenced
- what output and tracing preferences are requested

The local example is intentionally self-contained under one round folder:

- `configs/rounds/local-ic-vs-jcodemunch/round.pkl`
- `configs/rounds/local-ic-vs-jcodemunch/scoring/localization-objective.pkl`
- `configs/rounds/local-ic-vs-jcodemunch/policies/challenger_policy.py`
- `configs/rounds/local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/`

That keeps the manifest, local scoring file, local policy artifact, and local bundle root together instead of spreading them across the repo with `../` path traversal.

## Evaluator And Next-Challenger Separation

The evaluator owns bounded model/tool execution only.

The optimizer owns next-challenger proposal behavior.

That separation is explicit in the manifest shape:

- `evaluator` has model, bounds, and retry settings
- `optimizer` has model, bounds, and tool policy

There is deliberately no `pipeline` block under `evaluator`.

## Scoring Boundary

`scoring.objective` points to the separate visible scoring file.

The round manifest does not define objective math.

The scoring objective file answers:

- how bundled `evidence.pkl` evidence is turned into named objective values
- which value is final

That scoring behavior remains separate from round configuration.

## Current Scope

This first manifest surface proves:

- SearchBench can carry typed Pkl round files
- `pkl-go` can resolve them into Go structs
- Go can validate SearchBench-specific invariants afterward
- the manifest surface stays in `internal/adapters/config/pkl` instead of leaking Pkl runtime concerns into pure packages

## Running A Manifest

The current executable path is:

```text
searchbench round run --manifest configs/rounds/local-ic-vs-jcodemunch/round.pkl
```

That command currently uses the safe local/fake comparison seam in
`internal/app/round`, but `round` is only one execution strategy. The
canonical round semantics now live in `internal/app/evaluation`. It does
not call real IC MCP, jCodeMunch MCP, or provider APIs in tests.

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
- explicitly references prior `evidence.pkl` or related artifacts
- produces a new immutable bundle

It should not rely on hidden Go optimizer state to thread previous score
evidence, and it should not start with a bespoke content-addressed lookup
system.

## Manifest-Relative Paths

These manifest fields resolve relative to the directory containing
`round.pkl` unless they are already absolute:

- `scoring.objective`
- `policies.incumbent.policy.path`, when present later
- `policies.challenger.policy.path`
- `outputConfig.bundleRoot`

CLI-only overrides are separate from manifest-relative resolution. For example,
`searchbench round run --bundle-root ...` resolves from the caller environment rather
than from the manifest directory.

## Produced Artifacts

The manifest-driven local round writes an immutable bundle under the resolved
bundle root:

- `resolved-round.json`
  - resolved manifest projection used for execution
- `round-report.json`
  - structured comparison report
- `evidence.pkl`
  - Pkl-readable round evidence projected from `round-report.json`
- `objective.json`
  - validated result of evaluating the configured objective against `evidence.pkl`
- `metadata.json`
  - deterministic artifact inventory
- `round-report.txt`
  - optional human-readable rendering

## Current Scope

The current executable path intentionally remains local/fake-only for runtime
execution. It does not yet implement:

- real IC MCP execution
- real jCodeMunch MCP execution
- repo materialization
- tree-sitter indexing
- real next-challenger execution
- lineage or parent-round discovery by hash
- tracing export

Content hashes may still remain useful later as integrity and reproducibility
metadata, but they are not the primary lineage mechanism in this first
implementation direction.
