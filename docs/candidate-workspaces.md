# Candidate workspaces

How SearchBench isolates, validates, launches, and records optimizable backend candidates.

## The invariant

**The workspace that passes validation is the workspace whose MCP server launches.**

SearchBench does not validate one tree and launch another. It does not mutate the original backend source in place. It **materializes an isolated candidate workspace**, validates proposals there, and launches MCP from **that same copy**.

## Lifecycle

```text
WorkspaceSeedProvider
  → WorkspaceSeed
  → ICCandidateWorkspace (materializer)
  → ValidateProposalInWorkspace
  → AcceptedICCandidate
  → MCP launch
  → bundle / evidence
```

| Step | Meaning |
| --- | --- |
| **WorkspaceSeedProvider** | Resolves a backend source into a stable **seed** (`local_path` or `buck_descriptor`, …). |
| **WorkspaceSeed** | Stable source identity + root path to copy from. |
| **ICCandidateWorkspace** | Mutable isolated copy used for one candidate attempt. |
| **ValidateProposalInWorkspace** | Pipeline runs with cwd = candidate workspace root. |
| **AcceptedICCandidate** | Validated policy + launch spec bound to that workspace. |
| **Bundle / evidence** | Records seed identity, candidate workspace identity, policy identity, validation steps. |

Today this path is used heavily for **Iterative Context** selection-policy optimization; the same sandboxing model applies wherever a backend is materialized before validation and launch.

## Why this exists

- The optimizer needs a **safe place to stage** candidate policies.
- Validation must run against the **staged candidate**, not the pristine source tree.
- Runtime must launch the **exact tree** that passed validation.
- Evidence must distinguish **same seed, different candidate attempt**.

## Identity model

| Identity | Role |
| --- | --- |
| **WorkspaceSeedIdentity** | Stable across materializations: provider, source label/path, tree digest. |
| **ICCandidateWorkspace.ID** | One concrete mutable instance per materialization (e.g. seed id + temp dir basename). |
| **Policy identity** | Staged policy path, id, hash, symbol (`score_fn`, …). |
| **Runtime identity** | Workspace + seed + policy + active score verification for MCP. |

Evidence can answer: “same backend seed, new candidate attempt.”

## Provider variants

Two ways to produce the **seed** for the same lifecycle (not the main concept — sandboxing is):

| Provider | Use when | What it buys |
| --- | --- | --- |
| **`local_path`** | Public/default, local dev, tests without Buck | Simplest path: directory path is enough |
| **`buck_descriptor`** | Internal/meta-harness, graph-addressed backend | Stable Buck label + JSON descriptor (launcher, validator, admin) |

`git` and `archive` are reserved in Pkl and **not** implemented.

### `local_path` (public / default)

Use when a directory path is enough.

```pkl
runtime {
  workspaceSeed {
    provider = "local_path"
    localPath = "src/iterative-context"
  }
}
```

The provider resolves that path into a `WorkspaceSeed`. The materializer copies it into an isolated `ICCandidateWorkspace`. Validation and MCP launch use **the copy**, not the original checkout path as the mutable surface.

`validateIterativeContextProposal` uses this provider by default.

### `buck_descriptor` (internal)

Use when the backend should be addressed as a **repo graph** capability.

```pkl
runtime {
  workspaceSeed {
    provider = "buck_descriptor"
    buckDescriptorTarget = "//src/iterative-context:optimizable_backend"
  }
}
```

**Level 2 semantics (current):**

```text
buck2 build //src/iterative-context:optimizable_backend
  → descriptor target exists in the Buck graph

BuckBackendDescriptorProvider
  → loads optimizable_backend.json from the repo checkout
  → descriptor source.kind = local_path
  → same materializer copies that path into ICCandidateWorkspace
```

Buck buys **graph identity and descriptor structure**, not archive snapshots (deferred).

## Responsibility split

| Layer | Owns |
| --- | --- |
| **Pkl** | Provider intent in the round manifest |
| **Provider** | Seed resolution |
| **Materializer** | Isolated candidate workspace on disk |
| **Validation pipeline** | Candidate acceptance |
| **Runtime** | MCP launch from accepted candidate |
| **Bundle / evidence** | What happened |
| **Buck** | Repo operations and graph-addressable descriptors |

**Buck is not required** for public users who only use `local_path`.

## Backend descriptor vs repo checks

| Layer | Role |
| --- | --- |
| **Backend descriptor** | Candidate-facing runtime / optimization contract |
| **Repo Buck targets** | Contributor, CI, meta-harness validation |

Backend descriptors must **not** include `repo_checks`. Repo checks stay on:

```text
//:check
//:check_full
//src/iterative-context:check
//src/iterative-context:check_full
```

## Decision record

- **`local_path`** is sufficient and is the **public/default** path.
- **`buck_descriptor`** is kept for internal infra / meta-harness workflows.
- **Archive snapshots**, git providers, and source manifests remain **deferred**.

## Reference

| Topic | Location |
| --- | --- |
| Pkl schema | `configs/schema/SearchBenchRound.pkl` |
| Config validation | `src/searchbench-go/internal/adapters/config/pkl/workspace_seed.go` |
| Materialization | `src/searchbench-go/internal/adapters/workspace/` |
| Candidate validation | `src/searchbench-go/internal/agents/optimizer/policy/candidate_pipeline.go` |
| Optimizer pipeline steps | [reference/optimizer-policy-validation.md](./reference/optimizer-policy-validation.md) |

Legacy doc name: [workspace-seeds.md](./workspace-seeds.md) redirects here.
