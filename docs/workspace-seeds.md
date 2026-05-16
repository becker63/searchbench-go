# Workspace seed providers

Iterative Context (IC) optimization copies a **workspace seed** into an isolated **candidate workspace**, validates proposals there, then launches MCP from that same tree. The invariant: **the workspace that passes validation is the workspace whose MCP server launches.**

```text
WorkspaceSeedProvider
  → WorkspaceSeed
  → ICCandidateWorkspace (materializer)
  → ValidateProposalInWorkspace
  → AcceptedICCandidate + MCP launch
  → bundle/evidence (seed + policy identity)
```

## Seed identity vs candidate workspace identity

- **`WorkspaceSeedIdentity`** — stable across materializations from the same source (provider, source label, tree digest).
- **`ICCandidateWorkspace.ID`** — one concrete mutable instance per materialization (for example `seed-id` + temp dir basename). Evidence can distinguish “same seed, different attempt.”

## Responsibility split

| Layer | Declares |
| --- | --- |
| **Pkl** (`configs/schema/SearchBenchRound.pkl`) | Provider intent: `local_path` or `buck_descriptor` |
| **Buck** | Legal repo operations (`//src/iterative-context:optimizable_backend`, `//:check`, …) |
| **Bundle / evidence** | What happened: seed identity, validation steps, runtime identity |

Buck is **not** required for public users. Pkl stays provider-neutral at the schema level.

### Pkl schema

```pkl
typealias WorkspaceSeedProvider = "local_path" | "buck_descriptor" | "git" | "archive"

class Runtime {
  workspaceSeed: WorkspaceSeedConfig?
}

class WorkspaceSeedConfig {
  provider: WorkspaceSeedProvider = "local_path"
  localPath: String?
  buckDescriptorTarget: String?
}
```

Go validation: `internal/adapters/config/pkl/workspace_seed.go` — `local_path` requires `localPath`; `buck_descriptor` requires `buckDescriptorTarget`; `git` / `archive` are reserved and rejected.

## `local_path` (public / default)

**When to use:** public examples, local dev, tests without Buck, simple backends where a directory path is enough.

**Benefits:** simplest mental model, no Buck for consumers, easiest failure modes.

```pkl
runtime {
  workspaceSeed {
    provider = "local_path"
    localPath = "src/iterative-context"
  }
}
```

```go
seed, err := localpath.Provider{Source: "src/iterative-context"}.PrepareSeed(ctx)
```

`validateIterativeContextProposal` uses this path by default.

## `buck_descriptor` (internal / meta-harness)

**When to use:** internal IC optimization, graph-addressable backend identity, config tied to Buck labels, release/evidence provenance experiments.

**Benefits:** stable Buck label (`//src/iterative-context:optimizable_backend`), candidate-facing JSON descriptor (launcher, validator, runtime admin).

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

```pkl
runtime {
  workspaceSeed {
    provider = "buck_descriptor"
    buckDescriptorTarget = "//src/iterative-context:optimizable_backend"
  }
}
```

```go
seed, err := buckdescriptor.Provider{
    DescriptorTarget: "//src/iterative-context:optimizable_backend",
}.PrepareSeed(ctx)
```

## Backend descriptor vs repo checks

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

## Decision record

- **LocalPathProvider** is sufficient and should be the **public/default** path.
- **BuckBackendDescriptorProvider** is worth keeping as an **internal** infra/meta-harness path.
- **Archive snapshots** remain deferred.
- **Deferred providers:** `git`, `archive`, source manifest snapshots (schema reserved only).

## Related

- Optimizer validation pipeline: [reference/optimizer-policy-validation.md](./reference/optimizer-policy-validation.md)
- Package entry: `internal/adapters/workspace/`, `internal/agents/optimizer/policy/candidate_pipeline.go`
