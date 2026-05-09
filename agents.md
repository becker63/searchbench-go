# AGENTS.md

SearchBench-Go is an evaluation harness for agentic code-search systems.

The project compares a baseline retrieval system against a candidate retrieval system, executes both against the same task slice, scores the results, and produces durable evidence bundles suitable for promotion/regression decisions.

This repository intentionally separates:
- pure evaluation/scoring models
- orchestration/application flow
- effectful integrations (LLMs, MCP, tracing, filesystem, providers)

The architecture is designed to keep the evaluation model inspectable and reproducible.

---

# Start Here

If you are trying to understand the system quickly, read these files in order:

## 1. Complete End-to-End Flow

These files demonstrate the entire intended lifecycle of the system.

- `internal/app/locale2e/run.go`
- `internal/app/locale2e/run_test.go`

This is the clearest expression of the current system model.

The E2E flow currently models:

1. loading manifests
2. running evaluation
3. producing scored bundles
4. invoking optimizer flow
5. generating optimization outputs

The E2E tests are intentionally more important than the CLI surface.

---

## 2. Evaluation Flow

- `internal/app/evaluation/run.go`
- `internal/domain/evaluation/...`

Evaluation is responsible for:
- executing baseline/candidate systems
- collecting predictions/results
- producing scoring inputs
- rendering reports/evidence bundles

Evaluation should remain structurally deterministic and inspectable.

---

## 3. Optimizer Flow

- `internal/app/optimizer/run.go`
- `internal/domain/optimizer/...`

The optimizer consumes evaluation outputs and proposes improvements.

The optimizer is intentionally modeled separately from evaluation:
- evaluator = bounded execution/scoring
- optimizer = proposal generation

---

## 4. PKL Configuration + Lineage

- `pkl/`
- `examples/`
- `schema/`

PKL manifests are the canonical interface of the system.

The repository prefers explicit lineage and inspectable configuration over hidden orchestration state.

The `amends` model is important:
- runs inherit from prior runs
- scoring can reference parent outputs
- evaluation history becomes structurally inspectable

---

# Architectural Principles

## Prefer Explicit Models

If a concept matters to evaluation semantics, it should usually exist as:
- a typed domain structure
- a manifest field
- a scoring intermediary
- or a bundle artifact

Avoid hiding important evaluation state in:
- traces
- callbacks
- implicit runtime behavior
- provider-specific abstractions

---

## Pure vs Effectful Separation

The repository intentionally separates:
- pure domain modeling
- orchestration/application flow
- infrastructure integrations

Pure layers should not depend on:
- filesystem access
- network calls
- tracing SDKs
- provider SDKs
- CLI frameworks

Effectful adapters belong near the edges.

---

## Bundles Are First-Class Outputs

A run is not just a trace.

A run produces:
- reports
- objectives
- resolved manifests
- metadata
- scoring projections
- lineage information

The bundle is the durable artifact.

---

## The System Is About Promotion Decisions

The core question is not:
> "Did the agent run?"

The question is:
> "Should this candidate replace the current system?"

Everything in the architecture should support:
- comparison
- regression detection
- inspectability
- reproducibility
- promotion confidence

---

# Repository Navigation

## Application Layer

Application orchestration lives in:

- `internal/app/...`

These packages coordinate:
- loading manifests
- wiring dependencies
- executing flows
- emitting bundles

Application packages may depend on effectful integrations.

---

## Domain Layer

Core models live in:

- `internal/domain/...`

These packages should remain:
- deterministic
- typed
- inspectable
- provider-agnostic

Avoid introducing:
- filesystem access
- tracing logic
- provider SDKs
- CLI behavior

into the domain layer.

---

## Integrations

Integrations live near the edges of the repository.

Examples:
- MCP clients
- LangSmith callbacks
- filesystem materialization
- model providers
- repository cloning
- bundle persistence

The architecture intentionally treats these as adapters.

---

# Current Development Priorities

The current repository phase is:

## Transition From Modeling → Real Integrations

The high-level architecture and evaluation model are largely established.

Current work focuses on:
- real MCP integration
- real provider execution
- LangSmith tracing integration
- repository materialization
- end-to-end execution against real datasets

without collapsing architectural boundaries.

---

# Guidance For Agents

When making changes:

## Prefer Extending Existing Models

Avoid introducing parallel abstractions unless necessary.

Search for:
- existing domain types
- reducers
- manifest structures
- bundle projections

before adding new concepts.

---

## Preserve Structural Legibility

This repository values:
- explicitness
- inspectability
- typed flows
- deterministic structure

over:
- hidden magic
- framework-heavy abstractions
- implicit state propagation

---

## Keep Integrations Contained

When adding:
- tracing
- providers
- MCP execution
- filesystem behavior
- runtime orchestration

keep those concerns isolated from pure evaluation/scoring logic.

---

## Favor Readable Flows

The repository prefers:
- obvious orchestration
- explicit lifecycle stages
- inspectable transformations

over excessive abstraction.

If a new reader cannot follow the lifecycle from:
- manifest
→ evaluation
→ scoring
→ bundle
→ optimizer

the abstraction is probably too indirect.

---

# Mental Model

SearchBench-Go is closer to:
- a release engineering system
- an evaluation control plane
- a reproducible experiment harness

than a traditional agent framework.

The unit of concern is not a single trace.

The unit of concern is:
- a candidate system
- evaluated against a baseline
- producing durable evidence
- suitable for promotion decisions.
