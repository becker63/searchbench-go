You are working in the SearchBench-Go repository.

Task: implement the GitHub issue titled “Implement immutable run bundles and deterministic bundle serialization”.

First, read the project guidance:

1. Read AGENTS.md.
2. Read docs/engineering/agentic-development-flow.md.
3. Read docs/engineering/issue-style-guide.md.
4. Read the target issue with:

   gh issue view 17

Treat the GitHub issue body as the source of truth. If this prompt conflicts with the issue, follow the issue and mention the conflict in your final summary.

Core intent:

Implement the first local immutable run bundle substrate for SearchBench-Go.

A bundle is a file-native artifact directory produced after a baseline/candidate comparison. It should serialize:

- resolved.json
- report.json
- score.json
- metadata.json
- optional human-readable report output
- a completion marker written last

This issue must not add Pkl. Pkl should arrive later, after bundles exist.

Important architecture constraints:

- Do not implement Pkl.
- Do not implement optimizer lineage.
- Do not implement content-addressed previous-run references.
- Do not implement IC, jCodeMunch, MCP, repo materialization, tree-sitter, or LangSmith.
- Do not add a CLI command unless the issue explicitly requires it.
- Do not redefine score semantics.
- Do not move report, score, or run domain models into the bundle package.
- Do not make report rendering depend on terminal styling details.
- Do not leak raw policy source into report-safe bundle outputs.

Suggested package:

Prefer internal/artifact unless the existing code strongly suggests a better name.

A reasonable shape is:

internal/artifact/
  bundle.go
  writer.go
  metadata.go
  serialize.go
  errors.go
  writer_test.go

Design guidance:

- Bundle writing is effectful filesystem work.
- Bundle models should be small and provider-neutral.
- Bundle code may consume report.CandidateReport, score/report/run/domain values, and caller-provided resolved input.
- Bundle code should not own scoring, reporting, Eino execution, tracing, or CLI behavior.
- Use staging writes so partial failures do not look complete.
- Write the completion marker last.
- Fail if the final bundle already exists and is complete.
- Prefer injected clocks and caller-provided IDs in tests.
- Keep serialization deterministic enough for future hashing.

Implementation expectations:

Create types roughly equivalent to:

- BundleRequest
- BundleRef
- BundleFile
- BundleMetadata
- BundleError or typed failure kind

The exact names can differ if the repo already has better naming conventions.

The writer should have one obvious entrypoint, such as:

WriteBundle(ctx, request) (BundleRef, error)

The request should include:

- root path
- bundle id
- resolved input
- candidate report
- optional rendered report string or optional renderer
- created time or clock

The output directory should initially look like:

artifacts/runs/<bundle-id>/
  resolved.json
  report.json
  report.md or report.txt when provided
  score.json
  metadata.json
  COMPLETE

Do not overbuild this into a full artifact store yet.

Score evidence:

Create score.json from report-derived evidence only.

It should preserve useful comparison evidence, such as:

- metric comparisons
- metric names
- baseline values
- candidate values
- deltas
- regressions
- promotion decision
- run counts
- failure counts

Do not invent a new scoring reducer.

Determinism:

Tests should prove fixed input and fixed time produce identical serialized bytes.

Use stable JSON field names and stable indentation.

Avoid absolute paths in serialized content unless they are explicitly bundle metadata.

Failure behavior:

Represent typed failures for at least:

- bundle_validation_failed
- bundle_filesystem_failed
- bundle_serialization_failed
- bundle_already_exists
- bundle_finalize_failed

Use ordinary Go errors. Preserve wrapping where useful.

Tests:

Add tests that prove:

1. a fake CandidateReport can be serialized into a bundle
2. resolved.json is written
3. report.json is written
4. score.json is written
5. metadata.json is written
6. COMPLETE is written last
7. repeated serialization with fixed input and fixed clock produces identical bytes
8. existing completed bundle fails with bundle_already_exists
9. serialization failure does not create a completed bundle
10. metadata lists every generated artifact
11. report-safe outputs do not include raw policy source
12. the bundle package does not import Pkl, Eino, MCP, LangSmith, materialization, or tree-sitter packages

Use fake/local report data only. Do not hit the network. Do not require real model calls.

Before finishing:

- Run gofmt on changed Go files.
- Run go test ./...
- If tests fail for reasons unrelated to your change, report that clearly.
- Keep the diff bounded to this issue.

Final response format:

Summarize:

- files changed
- bundle directory shape implemented
- tests added
- commands run
- any deviations from the issue
- any follow-up issues exposed

Do not claim Pkl, lineage, content-addressing, real backend execution, or tracing are implemented.
