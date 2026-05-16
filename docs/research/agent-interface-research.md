# SearchBench: Evaluating Agent Interfaces, Not Just Agents

> **Research note** — not the operational docs index. Start at [../index.md](../index.md).

SearchBench is an evaluation harness for studying how different repository interfaces change the behavior of coding agents.

The central claim is simple:

> The model is not the whole system.
> The interface changes what the agent is capable of.

Most coding-agent benchmarks ask:

> Which model writes better code?

SearchBench asks a different question:

> Which environment makes an agent behave like a better engineer?

That means evaluating not only models, but the tools, graphs, validation surfaces, artifacts, and feedback loops surrounding them.

---

## Thesis

Agent performance is not only a property of the model.

It is distributed across:

```text
model
prompt
tool surface
repo structure
build graph
code graph
validation loop
artifact format
human inspection surface
````

A weak interface can make a strong model wasteful, confused, or overconfident.

A strong interface can make the same model more reliable, efficient, and inspectable.

SearchBench exists to measure that difference.

---

## Core Research Question

Given the same model, same repository, and same task:

```text
How does the available interface change the agent's behavior?
```

For example:

```text
same model
same repo
same task
different tool surface
different outcome
```

This lets us test whether agents improve when the repository exposes better operational structure.

---

## Interface Families

SearchBench can compare at least three broad interface families.

### 1. Raw Repository Interface

The agent receives normal repo access:

```text
files
grep/search
shell commands
README prose
ad hoc scripts
```

This is the default world most coding agents operate in.

The agent has to infer:

```text
what files matter
what commands exist
what tests prove the change
what generated files need updating
when it is done
```

This interface is realistic, but it forces the model to recover project policy from scattered prose and convention.

---

### 2. Code Intelligence Interface

The agent receives code-structure tools:

```text
symbols
references
call graph
file graph
semantic search
bounded lookahead
```

Examples:

```text
CodeQL
LSP
jCodeMunch
Iterative Context
static code graphs
```

This interface helps answer:

```text
Where should I look?
What code is related?
What calls what?
What files are likely relevant?
```

This is the current core SearchBench direction: measuring whether better code-search and graph-exploration tools improve localization and repair behavior.

---

### 3. Work Graph Interface

The agent receives repository-operation tools:

```text
Buck2 targets
build graph queries
test suites
proof targets
generated artifact dependencies
legal operations
```

Examples:

```text
buck2 targets
buck2 uquery
buck2 test //:check
buck2 test //:check_full
config-bundle targets
release-candidate targets
```

This interface helps answer:

```text
What am I allowed to do?
What does this action depend on?
What proves this change?
Which validation target is appropriate?
What is too expensive or too live for normal checks?
When am I done?
```

This may be as important as code search.

Code graphs help the agent understand code.

Work graphs help the agent understand engineering lifecycle.

---

## Key Distinction

```text
Code graph:
  lookahead over meaning

Work graph:
  lookahead over action
```

A code graph helps an agent find relevant files.

A work graph helps an agent choose the right operation and proof.

Many agent failures are not failures of syntax or code generation. They are lifecycle failures:

```text
ran the wrong checks
skipped generated files
edited the wrong layer
over-tested
under-tested
used live/manual targets in deterministic gates
did not know what counted as done
```

Buck-like systems can expose those lifecycle rules as graph structure instead of prose.

---

## Buck2 as an Agent Interface

In this framing, Buck2 is not only a build system.

It is an agent-facing action graph.

A Buck target is a named legal move:

```text
//:check
//:check_full
//src/searchbench-go:check
//src/iterative-context:check_full
//configs/rounds/optimize-ic:round_validate
//configs/rounds/optimize-ic:ic_workspace_smoke
```

Instead of asking an agent to infer commands like:

```sh
go test ./...
pytest
ruff check
basedpyright
pkl eval
repomix
```

the repository can expose:

```text
these are the legal operations
these are their dependencies
these are their costs
these are their artifacts
these are their proof obligations
```

That turns the repo from a pile of scripts into a structured operational environment.

---

## Structured Agent Operations

A future SearchBench experiment could avoid asking the agent to write raw Buck/Starlark.

Instead, the agent could emit structured operations:

```json
{
  "operation": "extend_test_suite",
  "package": "",
  "suite": "check_full",
  "add_tests": [
    "//configs/rounds/optimize-ic:round_validate"
  ],
  "reason": "deterministic config validation belongs in the full gate"
}
```

A deterministic renderer would turn this into a Buck file edit.

Then Buck validates the graph.

Then SearchBench records whether the agent chose the right operation and proof.

This is the important pattern:

```text
agent emits semantic intent
system renders sanctioned repo operation
Buck validates the work graph
tests validate behavior
bundle records evidence
```

The agent does not need arbitrary shell access to be useful.

It needs a good action language.

---

## Possible Experiment Design

SearchBench can run the same task under different tool surfaces.

### Baseline

```text
files
grep
shell
README prose
```

### Candidate A: Code Graph

```text
files
grep
shell
code graph tools
symbol lookup
references
bounded lookahead
```

### Candidate B: Work Graph

```text
files
grep
shell
Buck target listing
Buck query
Buck target execution
structured Buck operations
```

### Candidate C: Hybrid

```text
code graph
work graph
bundle evidence
release report
```

The model stays constant.

The repo stays constant.

The task stays constant.

Only the interface changes.

---

## Example Tasks

SearchBench could evaluate tasks like:

```text
Add a config bundle and expose the correct validation target.

Determine which target proves a change to a round manifest.

Add a generated-file check to the full deterministic gate.

Fix a failing Buck target with the smallest patch.

Add an IC workspace smoke target without putting live/provider-backed work into //:check.

Given a changed file set, choose the minimal proof targets.

Promote a release candidate only if the correct evidence bundle passes.
```

These tasks measure agentic engineering behavior, not just code editing.

---

## Metrics

SearchBench should score more than pass/fail.

Potential metrics:

```text
correct patch
correct target selected
minimal proof selected
invalid commands attempted
irrelevant commands attempted
tokens spent
wall-clock time
number of retries
files touched
unrelated files touched
over-testing
under-testing
lifecycle policy violations
whether the agent stopped with a valid proof
```

A key metric could be:

```text
proof distance
```

That means:

```text
How far was the agent's chosen validation path from the minimal correct proof path?
```

For example, if the correct proof is:

```text
//configs/rounds/optimize-ic:round_validate
```

but the agent runs:

```text
go test ./...
pytest
nix flake check
buck2 test //:check_full
```

the task may pass, but the agent has shown weak operational understanding.

SearchBench can measure that.

---

## Hypothetical Result

A useful SearchBench result might look like:

```text
Same model, same repo, same task.

Raw repo interface:
  42k tokens
  17 tool calls
  6 irrelevant commands
  wrong validation target

Code graph interface:
  24k tokens
  9 tool calls
  found correct files faster
  still unsure what to run

Work graph interface:
  18k tokens
  6 tool calls
  selected correct proof target
  stopped cleanly

Hybrid interface:
  14k tokens
  5 tool calls
  correct patch
  correct proof
  clean evidence bundle
```

This would support the claim that interface design changes effective agent capability.

---

## Relationship to SearchBench's Existing Direction

SearchBench already studies code-localization interfaces:

```text
baseline retrieval
vs
candidate graph exploration
```

The Buck/work-graph direction extends the same idea from code search to engineering lifecycle.

```text
Iterative Context:
  better code-search lookahead

Buck2:
  better action/proof lookahead

Bundles:
  better evidence and release memory

Visualization:
  better human inspection
```

Together, these form a broader research product:

```text
SearchBench evaluates agent environments.
```

Not just agents.

---

## Product Thesis

SearchBench should not only answer:

```text
Which model is best?
```

It should answer:

```text
Which interface makes this model behave better?
```

This is the deeper product category.

SearchBench can compare:

```text
models
prompts
tools
code graphs
work graphs
artifact bundles
validation loops
visualization surfaces
```

under a shared evaluation harness.

The result is a way to study agentic engineering systems as systems.

---

## Philosophical Claim

The underlying philosophy is:

> Tools matter because they change what cognition is cheap, expensive, visible, or impossible.

For agents, this is especially important.

A model operating over raw files and shell commands must infer too much.

A model operating inside a well-designed environment can spend more of its effort on the actual engineering problem.

The goal is not to make the model smarter in isolation.

The goal is to design an environment where the model's intelligence is usable, bounded, inspectable, and correctable.

---

## Short Version

```text
CodeQL makes code queryable.

Buck makes work queryable.

SearchBench can measure whether queryable work makes agents better engineers.
```

Or:

```text
The benchmark is not just the model.

The benchmark is the model inside an environment.
```
