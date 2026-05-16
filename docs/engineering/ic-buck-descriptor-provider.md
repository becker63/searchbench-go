# IC Buck descriptor workspace seed provider

Complete **Buck-backed optimizable backend descriptor** integration for the SearchBench ↔ Iterative Context candidate lifecycle. This branch does not ship `LocalPathProvider` as the primary public path.

## Level 2 semantics

```text
buck2 build //src/iterative-context:optimizable_backend
  → validates the descriptor target exists in the Buck graph

BuckBackendDescriptorProvider
  → loads src/iterative-context/optimizable_backend.json from the repo checkout
  → source.kind = local_path (descriptor-internal)
  → materializer copies that path into ICCandidateWorkspace
```

Archive snapshots and source manifests are **deferred**. The provider does not pretend they exist.

## What Buck buys

- Graph-addressable backend: `//src/iterative-context:optimizable_backend`
- Candidate-facing JSON contract: launcher, validator steps, runtime admin
- Stable evidence identity: `provider=buck_descriptor`, `source=<Buck label>`

## Repo checks boundary

Backend descriptors must **not** include `repo_checks`. Contributor validation stays on ordinary targets:

```text
//:check
//:check_full
//src/iterative-context:check
//src/iterative-context:check_full
```

## Config example

```pkl
runtime {
  workspaceSeed {
    provider = "buck_descriptor"
    buckDescriptorTarget = "//src/iterative-context:optimizable_backend"
  }
}
```

## Validation

```sh
cd src/searchbench-go
go test ./...

nix develop -c buck2 build //src/iterative-context:optimizable_backend
nix develop -c buck2 test //:check
nix develop -c buck2 test //:check_full
```
