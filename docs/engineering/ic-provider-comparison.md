# IC provider comparison (integration result)

Comparison branches `compare/ic-localpath-provider-complete` and `compare/ic-buck-descriptor-provider-complete` informed this integration. The union lives on `integrate/ic-workspace-seed-providers`.

## Summary

| Dimension | `local_path` | `buck_descriptor` |
| --- | --- | --- |
| implementation complexity | Low | Medium (+ descriptor schema) |
| LOC / new concepts | `localpath` + shared materializer | `buckdescriptor` + submodule target |
| runtime overhead | One tree copy | Descriptor load + same copy (Level 2) |
| debuggability | High (plain paths) | Medium (descriptor + Buck label) |
| failure modes | Missing path, copy I/O | Malformed descriptor, Buck unavailable |
| workspace purity | Same excludes (`.git`, caches, …) | Same |
| evidence identity | path + digest | Buck label + digest |
| public ergonomics | Best | Repo-internal |
| meta-harness legibility | Weaker | Stronger (graph-addressable) |
| archive pressure | None | Deferred explicitly |

## Decision

1. **LocalPathProvider is sufficient for now** — default optimizer validation uses it.
2. **BuckBackendDescriptorProvider is worth adopting internally** — keep behind `buck_descriptor` config.
3. **Archive snapshots are deferred** — Level 2 `local_path` inside descriptor only.
4. **Final integration keeps both** behind `ports/workspace.SeedProvider`.

## Smoke test

`internal/adapters/workspace/compare/compare_test.go` runs both providers through materialization and checks key IC files exist in each candidate workspace.
