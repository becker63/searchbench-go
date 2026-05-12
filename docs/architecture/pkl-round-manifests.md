# Pkl Round Manifests

**Related docs:** [Documentation index](../README.md) · [Architecture](./architecture.md) · [Package boundaries](./package-boundaries.md)

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

The current round manifest surface is continuation-backed:

- `name`
- `round.id`
- `round.incumbent`, when defining a round from scratch
- `round.challenger`
- `round.matches`, when defining a round from scratch
- `round.evaluator`, when defining a round from scratch
- `round.scoring`, when defining a round from scratch

From-scratch rounds usually amend a game schema module and use helper
constructors:

```pkl
amends "../../schema/games/code-localization.pkl"
import "../../schema/games/code-localization-helpers.pkl" as game

name = "example-local-ic-vs-jcodemunch-round-001"

round = (game.defineFromScratch("round-001")) {
  incumbent = game.jcodemunch()
  challenger = (game.iterativeContext("policies/challenger_policy.py")) {
    selectionPolicy {
      id = "challenger-policy-round-001"
    }
  }
  matches = game.lca("py", "dev", 5)
  scoring = game.objective("scoring/localization-objective.pkl")
  evaluator = game.fakeEvaluator()
}
```

Continued rounds usually amend the previous bundle's `continuation.pkl` and
only patch the new round id plus the next challenger:

```pkl
amends "../local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001/continuation.pkl"
import "../../schema/games/code-localization-helpers.pkl" as game

name = "example-optimize-ic-round-002"

round {
  id = "round-002"

  challenger {
    generate {
      optimizer = game.fakeOptimizer()
      artifactName = "next_challenger_policy.round-002.py"
    }
  }
}
```

The checked-in examples are:

- `configs/rounds/local-ic-vs-jcodemunch/round.pkl`
- `configs/rounds/local-ic-vs-jcodemunch/scoring/localization-objective.pkl`
- `configs/rounds/local-ic-vs-jcodemunch/policies/challenger_policy.py`
- `configs/rounds/local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/`
- `configs/rounds/optimize-ic/round.pkl`
- `configs/rounds/optimize-ic/artifacts/games/code-localization/rounds/`

That keeps the from-scratch example, the continued generated-challenger example,
and their bundle roots in obvious places instead of spreading them across the
repo with extra example directories.

## Evaluator And Next-Challenger Separation

The evaluator owns bounded model/tool execution only.

The optimizer owns next-challenger proposal behavior inside a normal round.

That separation is explicit in the round surface:

- `round.evaluator` owns evaluation behavior
- `round.challenger.generate.optimizer` owns challenger materialization

There is deliberately no return to `mode = "optimization"` or a separate
top-level optimization manifest shape.

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

## Continuation Artifacts

Every completed round bundle writes both:

- `continuation.json`
- `continuation.pkl`

The split is intentional:

- `continuation.json` is machine authority
- `continuation.pkl` is amendable human lineage

Go still validates continuation by requiring an explicit completed bundle with:

- `COMPLETE`
- `continuation.json`

`continuation.pkl` exists so the next human-authored round can preserve Pkl
`amends` semantics without turning Pkl into the execution engine.

The preferred authoring direction is:

- write a completed round bundle
- amend that bundle's `continuation.pkl`
- let Go resolve and validate the parent bundle through `continuation.json`

This keeps lineage explicit and bundle-scoped. It does not rely on hidden
"latest round" discovery or a separate optimizer workflow graph.

## Manifest-Relative Paths

These manifest fields resolve relative to the directory containing
`round.pkl` unless they are already absolute:

- `round.scoring.objective`
- `round.incumbent.selectionPolicy.path`
- `round.challenger.selectionPolicy.path`

CLI-only overrides are separate from manifest-relative resolution. For example,
`searchbench round run --bundle-root ...` resolves from the caller environment rather
than from the manifest directory.

When a continued manifest amends a parent bundle's `continuation.pkl`, inherited
defaults may contain absolute file paths. That is expected: the bundle-local
continuation surface is readable lineage, while Go still verifies the explicit
bundle continuation via `continuation.json`.

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
- `continuation.json`
  - machine-stable continuation record validated by Go
- `continuation.pkl`
  - amendable continuation surface for the next human-authored round
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
