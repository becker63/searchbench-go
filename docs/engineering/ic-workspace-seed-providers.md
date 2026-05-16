# IC workspace seed providers

`WorkspaceSeedProvider` is the core abstraction. `LocalPathProvider` and `BuckBackendDescriptorProvider` are competing implementations that feed the **same** candidate workspace lifecycle.

## Shared lifecycle

```text
WorkspaceSeedProvider
  → WorkspaceSeed
  → ICCandidateWorkspace (materializer)
  → ValidateProposalInWorkspace (candidate validation)
  → AcceptedICCandidate + MCP launch
  → bundle/evidence (seed + policy identity)
```

Invariant: **the workspace that passes validation is the workspace whose MCP server launches.**

## `local_path`

### When to use

- Public users and examples
- Quick local development and debugging
- Tests that should not require Buck
- Simple backends where a directory path is enough

### What it buys

- Simplest mental model
- No Buck dependency for consumers
- Easiest failure modes to reason about
- Best public ergonomics

### Config

```pkl
runtime {
  workspaceSeed {
    provider = "local_path"
    localPath = "src/iterative-context"
  }
}
```

### Go

```go
seed, err := localpath.Provider{Source: "src/iterative-context"}.PrepareSeed(ctx)
```

## `buck_descriptor`

### When to use

- Internal SearchBench-Go / IC optimization
- Meta-harnesses and graph-addressable backend identity
- Config bundle targets tied to Buck labels
- Release/evidence provenance experiments

### What it buys

- Stable Buck label as backend identity (`//src/iterative-context:optimizable_backend`)
- Candidate-facing JSON descriptor (launcher, validator, runtime admin)
- Better repo-graph legibility for infra agents

### Level 2 semantics (current)

```text
buck2 build //src/iterative-context:optimizable_backend
  → validates the descriptor target in the Buck graph

BuckBackendDescriptorProvider
  → loads src/iterative-context/optimizable_backend.json from the repo checkout
  → source.kind = local_path inside the descriptor
  → same materializer copies that path into ICCandidateWorkspace
```

Archive snapshots and source manifests are **deferred**.

### Config

```pkl
runtime {
  workspaceSeed {
    provider = "buck_descriptor"
    buckDescriptorTarget = "//src/iterative-context:optimizable_backend"
  }
}
```

### Go

```go
seed, err := buckdescriptor.Provider{
    DescriptorTarget: "//src/iterative-context:optimizable_backend",
}.PrepareSeed(ctx)
```

## Boundary: backend descriptor vs repo checks

| Layer | Role |
| --- | --- |
| Backend descriptor | Candidate-facing runtime/optimization contract |
| Repo Buck targets | Contributor / CI / meta-harness validation |

Backend descriptors must **not** include `repo_checks`. Repo checks stay on:

```text
//:check
//:check_full
//src/iterative-context:check
//src/iterative-context:check_full
```

## Decision summary

- **LocalPathProvider** is sufficient and should be the **public/default** path (`validateIterativeContextProposal` uses it today).
- **BuckBackendDescriptorProvider** is worth keeping as an **internal** infra/meta-harness path.
- **Archive snapshots** remain deferred.

## Validation

```sh
cd src/searchbench-go
go test ./internal/adapters/workspace/... ./internal/agents/optimizer/policy/...

nix develop -c buck2 build //src/iterative-context:optimizable_backend
nix develop -c buck2 test //:check
```
