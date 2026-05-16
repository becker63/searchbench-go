# IC local path workspace seed provider

Complete **local file copy** integration for the SearchBench ↔ Iterative Context candidate lifecycle. This branch does not include Buck descriptor providers, `optimizable_backend.json`, or `buck_descriptor` config.

## What it does

```text
WorkspaceSeedProvider (local_path)
  → WorkspaceSeed
  → CandidateMaterializer (copy + exclude junk)
  → ICCandidateWorkspace
  → ValidateProposalInWorkspace (pipeline cwd = candidate root)
  → AcceptedICCandidate + ICLaunchSpec
  → MCP launch seam (cwd must match candidate root)
```

Invariant: the workspace that passes validation is the workspace whose MCP server launches.

## Why this path

- Simplest public integration: configure `runtime.workspaceSeed.localPath`.
- No Buck graph required for consumers.
- Identity: `provider=local_path`, `source=<path>`, `sha256=<tree digest>`.

## What it cannot express (vs Buck descriptor)

- Graph-addressable backend targets (`//src/iterative-context:optimizable_backend`).
- Declarative launcher/validator contracts in JSON for meta-harnesses.
- Future archive/git descriptor sources without new code.

## Config example

```pkl
runtime {
  workspaceSeed {
    provider = "local_path"
    localPath = "src/iterative-context"
  }
}
```

## Validation

```sh
cd src/searchbench-go
go test ./...
```

From repo root (after `nix develop`):

```sh
buck2 test //:check
buck2 test //:check_full
```
