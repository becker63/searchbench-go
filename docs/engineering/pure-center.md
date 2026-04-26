# Expanding the Pure Center

## Purpose

This note captures an important architectural direction behind SearchBench-Go:

    use exploratory systems to discover the real model,
    then preserve that model in a pure, typed center,
    while pushing effects to explicit adapter edges.

This is not only a SearchBench-Go idea. It is a broader engineering taste that can transfer across languages, infrastructure projects, build systems, developer tools, and AI evaluation harnesses.

## The core idea

A lot of good engineering tooling expands the part of a system that is:

- typed
- deterministic
- inspectable
- reproducible
- testable without the world
- readable without knowing every adapter
- portable across languages and runtimes

Call that part the pure center.

The pure center is where the project’s real nouns live.

In SearchBench-Go, that center is things like:

    domain
    run
    score
    report
    codegraph

These packages describe the stable model:

    What is a task?
    What is a system?
    What is a prediction?
    What happened during a run?
    How do we score it?
    How do we report it?
    What code universe are we reasoning over?

Those concepts should be understandable without knowing Eino, MCP, OpenAI, LangSmith, repo checkout logic, subprocess details, or the old Python implementation.

## The adapter edge

Effects still exist.

They just should not own the model.

The effectful edges are things like:

    Eino execution
    MCP sessions
    OpenAI/OpenRouter provider calls
    LangSmith tracing
    CLI subprocess validation
    repo materialization
    filesystem writes
    network calls
    test fixture harnesses

These are adapter-shaped concerns.

They are allowed to be messy, concrete, and runtime-specific, but they should depend inward on the pure center.

The pure center should not depend outward on them.

A useful dependency direction is:

    adapters -> pure model

not:

    pure model -> adapters

## Why Python mattered

The pure center could not have been designed correctly up front.

The original Python SearchBench was the discovery medium.

Python made it cheap to explore, refactor, glue things together, and learn from pressure. It allowed the project to discover the real nouns through use:

    task
    prediction
    run
    score
    report
    policy
    backend
    pipeline step
    failure classification
    dataset row
    graph distance
    prompt input

The Python version was messy in places because it had to be. It was doing semantic discovery.

It answered questions like:

    What keeps showing up?
    What do tests keep needing?
    What gets passed everywhere?
    What becomes painful when it is a dict?
    What failures need names?
    What concepts survive refactors?
    What is just adapter glue?
    What should be scored deterministically?
    What should never be model-visible?

The Go rewrite is valuable because those answers now exist.

Python found the shape.
Go is crystallizing it.

## Preserve semantics, not implementation shape

The goal of the Go rewrite is not to mechanically port Python SearchBench.

The goal is to preserve the hard-earned semantics while deleting accidental structure.

Preserve:

    typed task identity
    file-localization predictions
    run lifecycle
    deterministic scoring inputs
    pipeline step results
    structured failure classification
    writer/evaluator separation
    prompt boundaries
    policy identity
    bounded retry behavior

Do not blindly preserve:

    Python-specific class structure
    Pydantic implementation details
    dict-shaped data passing
    Langfuse-specific tracing shape
    custom state machine implementation
    worktree mutation as policy selection
    custom backend adapter machinery when Eino/MCP can buy it

A migration is successful when the stable model survives and the accidental scaffolding disappears.

## Why this improves readability

A pure center makes the project easier for newcomers and for future me.

A newcomer can start with:

    domain
    run
    score
    report
    codegraph

and understand the project before learning the adapters.

They can form the basic story:

    A task exists.
    A system attempts it.
    The system produces a prediction.
    A run records what happened.
    A scorer evaluates the run.
    A report explains the comparison.
    A code graph provides deterministic localization structure.

That is much easier than asking them to begin with model APIs, MCP, tracing, subprocesses, or benchmark sync.

The pure center gives the system a readable spine.

## Why this improves testing

The larger the pure center is, the more of the system can be tested without the world.

Good tests should often need no:

    API keys
    network
    live model calls
    repository checkout
    tracing backend
    subprocess
    filesystem mutation
    provider SDK behavior

This is why SearchBench-Go uses deterministic model fixtures and fake provider boundaries.

It lets execution code be tested through typed contracts instead of paid side effects.

The same principle applies to CLI validation:

    external command output -> StepResult
    StepResult -> PipelineClassification
    PipelineClassification -> bounded feedback
    feedback -> retry context

Once effects are converted into typed data, the rest of the system becomes easier to reason about.

## Tooling that expands the pure center

This architectural taste shows up in many tools.

Nix expands the pure part of environments and builds.

Buck2 expands the pure part of build graphs and dependency reasoning.

Kubernetes expands the declarative model of infrastructure state.

CRDs let teams define typed domain models inside a control plane.

Go types expand compile-time architectural constraints.

Pydantic expands runtime validation and schema clarity.

templ expands typed prompt rendering.

tree-sitter expands source code into structured data.

CodeQL expands code search into queryable semantics.

The shared pattern is:

    turn messy operational reality into typed, inspectable, reproducible models

That is the broader engineering direction.

## SearchBench-Go as an example

Agent systems are usually effect-heavy:

    prompt goes in
    model does something
    tools may be called
    text comes out
    maybe tests ran
    maybe cost exploded
    maybe tracing captured something

SearchBench-Go tries to carve out a larger pure center:

    TaskSpec
    SystemSpec
    PolicyArtifact
    Prediction
    Run
    ScoreSet
    CandidateReport
    StepResult
    PipelineClassification
    PromptInput
    CodeGraph

The model/provider/tooling world is still there, but it is pushed to the edge:

    Eino executes model calls.
    MCP exposes tools.
    CLI runner executes commands.
    LangSmith records traces.
    Repo materialization checks out code.
    Provider clients talk to remote APIs.

The center remains legible.

## Dangling boundary pressure

As adapters grow, they exert pressure on the pure center.

That pressure should be watched deliberately.

Examples:

    Eino types leaking into domain models
    provider details leaking into run records
    test fixtures shaping production abstractions
    CLI subprocess details leaking into scoring
    prompt inputs depending on executor packages
    backend-specific fields appearing in predictions
    tracing SDKs appearing in pure packages

These are signs that a boundary sweep may be needed.

Not every boundary issue should be fixed immediately. Some should become dangling issues: future cleanup notes that preserve architectural signal without derailing current implementation.

The ability to notice dangling activity is important.

It means the process is not only producing code. It is also detecting where the architecture wants to be cleaned up later.

## Development process

The current process supports this direction:

    discuss design
    write GitHub issue as implementation contract
    write agent prompt as execution wrapper
    run coding agent
    inspect result
    pack result
    feed discoveries into the next issue
    create dangling cleanup notes when needed

The issue is the contract.

The prompt is the execution protocol.

The agent result is evidence.

The next issue absorbs what was learned.

This process lets the pure center grow from implementation pressure rather than from abstract speculation.

## Practical rule

When adding a new feature, ask:

    Is this a stable concept or an effectful mechanism?

If it is a stable concept, it may belong near the center.

Examples:

    task
    prediction
    run phase
    score
    report
    pipeline step result
    failure kind

If it is an effectful mechanism, it probably belongs at the edge.

Examples:

    Eino model call
    MCP client session
    subprocess execution
    repo checkout
    LangSmith span
    OpenAI request
    filesystem write

Then ask:

    Can the effect be converted into typed data quickly?

If yes, the rest of the system can remain pure.

## Final principle

Use exploratory tools to discover the model.

Use typed systems to preserve it.

Use adapters to touch the world.

Keep expanding the part of the system that can be understood, tested, and changed without needing the world to be present.
