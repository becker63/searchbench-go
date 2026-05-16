# Candidate workspaces

How SearchBench isolates, validates, launches, and records optimizable backend candidates.

## The invariant

**The workspace that passes validation is the workspace whose MCP server launches.**

SearchBench does not validate one tree and launch another. It does not mutate the original backend source in place. It **materializes an isolated candidate workspace**, validates proposals there, and launches MCP from **that same copy**.

## Why two providers?

SearchBench has **two provider implementations** and **one sandboxing lifecycle**. They are not competing models — they are two ways to resolve a **source seed** for the same path.

| Provider | Audience | What it optimizes for | Requires Buck? |
| --- | --- | --- | --- |
| `local_path` | Public users, local examples, simple tests | Simplest path from directory → candidate workspace | **No** |
| `buck_descriptor` | Internal infra, meta-harness, release/provenance experiments | Stable Buck label + descriptor-backed backend identity | **Yes** |

**`local_path`** exists because public users and simple examples should not need Buck. If a backend is a directory on disk, the round points at that path; SearchBench copies it into a candidate workspace, validates there, and launches MCP from the copy. This is the default public mental model.

**`buck_descriptor`** exists because internal SearchBench development and future meta-harness workflows need **graph-addressable** backend identity. A label such as `//src/iterative-context:optimizable_backend` names a repo capability. The JSON descriptor can carry launcher, validator, runtime admin, and source metadata for optimization and release workflows.

**`buck_descriptor` is not more sandboxed than `local_path`.** Sandboxing comes from candidate workspace materialization and validation, not from Buck.

Both providers **converge before validation**:

```text
WorkspaceSeedProvider
  → WorkspaceSeed
  → same materializer
  → ICCandidateWorkspace
  → same validation pipeline
  → same MCP launch path
  → same bundle / evidence model
```

## Lifecycle

```text
WorkspaceSeedProvider → WorkspaceSeed → ICCandidateWorkspace
  → ValidateProposalInWorkspace → AcceptedICCandidate → MCP launch → bundle / evidence
```

| Step | Meaning |
| --- | --- |
| **WorkspaceSeedProvider** | `local_path` or `buck_descriptor` — resolves source into a `WorkspaceSeed` |
| **WorkspaceSeed** | Stable source identity + root path to copy from |
| **ICCandidateWorkspace** | Mutable isolated copy for one candidate attempt |
| **ValidateProposalInWorkspace** | Pipeline runs with cwd = candidate workspace root |
| **AcceptedICCandidate** | Validated policy + launch spec bound to that workspace |
| **Bundle / evidence** | Records seed identity, candidate workspace identity, policy identity |

## How to choose

**Use `local_path` when:**

- You are writing public examples or testing locally
- A backend directory path is enough
- You want the smallest mental model
- You do not want Buck as a user requirement

**Use `buck_descriptor` when:**

- You are inside the SearchBench monorepo
- The backend should be addressed as a Buck target
- A meta-harness needs graph identity
- You want descriptor-backed launcher / validator / runtime-admin metadata
- You are experimenting with release provenance or repo-graph planning

## Provider: `local_path` (public / default)

**Pkl excerpt:**

```pkl
runtime {
  workspaceSeed {
    provider = "local_path"
    localPath = "src/iterative-context"
  }
}
```

**Meaning:** Resolve directory → copy to temp `ICCandidateWorkspace` → validate and launch from the **copy**, not the original checkout path as the mutable surface.

`validateIterativeContextProposal` uses this provider by default.

## Provider: `buck_descriptor` (internal)

**Pkl excerpt:**

```pkl
runtime {
  workspaceSeed {
    provider = "buck_descriptor"
    buckDescriptorTarget = "//src/iterative-context:optimizable_backend"
  }
}
```

**Descriptor file:** `src/iterative-context/optimizable_backend.json`

```json
{
  "source": {
    "kind": "local_path",
    "path": "src/iterative-context"
  },
  "launcher": {
    "kind": "mcp_stdio",
    "cwd_mode": "candidate_workspace"
  },
  "candidate_validator": {
    "kind": "ic_policy_pipeline"
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

Buck buys **graph identity and descriptor structure**, not archive snapshots or source manifests (deferred). MCP still launches from the **materialized candidate workspace**, not from the raw repo checkout.

## Identity (evidence)

| Identity | Distinguishes |
| --- | --- |
| **WorkspaceSeedIdentity** | Same backend source across materializations |
| **ICCandidateWorkspace.ID** | One concrete mutable instance per attempt |
| **Policy identity** | Staged policy path, id, hash, symbol (`score_fn`, …) |
| **Runtime identity** | Workspace + seed + policy + active score verification |

Evidence can answer: “same seed, different candidate attempt.”

## Backend descriptor vs repo checks

| Layer | Role |
| --- | --- |
| **Backend descriptor** | Candidate-facing runtime / optimization contract |
| **Repo Buck targets** | Contributor, CI, and meta-harness validation |

Backend descriptors must **not** include `repo_checks`. Repo checks stay on:

```text
//:check
//:check_full
//src/iterative-context:check
//src/iterative-context:check_full
```

## Decision record

- **`local_path`** remains the **public/default** provider.
- **`buck_descriptor`** remains the **internal/meta-harness** provider.
- **`git`** and **`archive`** are reserved in Pkl and **not** implemented.
- Archive snapshots and source manifests remain **deferred**.

## Reference

| Topic | Location |
| --- | --- |
| Pkl schema | `configs/schema/SearchBenchRound.pkl` |
| Config validation | `src/searchbench-go/internal/adapters/config/pkl/workspace_seed.go` |
| Local path provider | `src/searchbench-go/internal/adapters/workspace/localpath/` |
| Buck descriptor provider | `src/searchbench-go/internal/adapters/workspace/buckdescriptor/` |
| Candidate materializer | `src/searchbench-go/internal/adapters/workspace/materialize/` |
| Optimizer validation | `src/searchbench-go/internal/agents/optimizer/policy/candidate_pipeline.go` |
| IC descriptor | `src/iterative-context/optimizable_backend.json` |
| IC MCP server | `src/iterative-context/src/iterative_context/server.py` |
| IC policy checks | `src/iterative-context/src/iterative_context/validate_policy.py` |
| Optimizer validation reference | [optimizer-policy-validation.md](./reference/optimizer-policy-validation.md) |

Legacy: [workspace-seeds.md](./workspace-seeds.md) redirects here.
