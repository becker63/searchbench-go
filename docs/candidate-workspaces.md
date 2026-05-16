# Candidate workspaces

How SearchBench isolates, validates, launches, and records optimizable backend candidates.

## The invariant

**The workspace that passes validation is the workspace whose MCP server launches.**

SearchBench does not validate one tree and launch another. It **materializes an isolated candidate workspace**, validates proposals there, and launches MCP from **that same copy**.

## Lifecycle

```text
WorkspaceSeedProvider → WorkspaceSeed → ICCandidateWorkspace
  → ValidateProposalInWorkspace → AcceptedICCandidate → MCP launch → bundle / evidence
```

| Step | Code |
| --- | --- |
| Seed resolution | `src/searchbench-go/internal/adapters/workspace/localpath/`, `.../buckdescriptor/` |
| Materialize copy | `src/searchbench-go/internal/adapters/workspace/materialize/` |
| Validate proposal | `src/searchbench-go/internal/agents/optimizer/policy/candidate_pipeline.go` |
| Backend contract | `src/iterative-context/optimizable_backend.json` |
| MCP server | `src/iterative-context/src/iterative_context/server.py` |
| Policy checks | `src/iterative-context/src/iterative_context/validate_policy.py` |

## Provider: `local_path` (public / default)

**Pkl excerpt** (in a round that sets `runtime.workspaceSeed`):

```pkl
runtime {
  workspaceSeed {
    provider = "local_path"
    localPath = "src/iterative-context"
  }
}
```

**Meaning:** Resolve directory → copy to temp `ICCandidateWorkspace` → validate and launch from the copy.

**Buck not required** for this path.

## Provider: `buck_descriptor` (internal)

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
  "source": { "kind": "local_path", "path": "src/iterative-context" },
  "launcher": { "kind": "mcp_stdio", "cwd_mode": "candidate_workspace", ... },
  "candidate_validator": { "kind": "ic_policy_pipeline", "steps": [...] }
}
```

**Meaning:** Buck labels the descriptor; provider loads JSON from checkout; same materializer copies `local_path` into a candidate workspace. Archive snapshots deferred.

## Identity (evidence)

| Identity | Distinguishes |
| --- | --- |
| **WorkspaceSeedIdentity** | Same backend source across attempts |
| **ICCandidateWorkspace.ID** | One materialization (per attempt) |
| **Policy identity** | Staged policy path, hash, `score_fn` symbol |

Evidence can answer: “same seed, different candidate attempt.”

## Backend descriptor vs repo checks

| Layer | Example |
| --- | --- |
| Backend descriptor | `optimizable_backend.json` — runtime / validation contract |
| Repo Buck targets | `//src/iterative-context:check_full`, `//:check_full` |

Descriptors must **not** embed `repo_checks`.

## Decision record

- **`local_path`** — public/default.
- **`buck_descriptor`** — internal / meta-harness graph identity.
- **git / archive** providers — reserved, not implemented.

## Reference

| Topic | Location |
| --- | --- |
| Pkl schema | `configs/schema/SearchBenchRound.pkl` |
| Config validation | `src/searchbench-go/internal/adapters/config/pkl/workspace_seed.go` |
| Optimizer steps | [optimizer-policy-validation.md](./reference/optimizer-policy-validation.md) |

Legacy: [workspace-seeds.md](./workspace-seeds.md) redirects here.
