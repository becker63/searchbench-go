You are working in the SearchBench-Go repository.

Task: implement GitHub issue #19, titled “Define objective result models for visible future Pkl scoring calculations”.

First, read the project guidance:

1. Read AGENTS.md.
2. Read docs/engineering/agentic-development-flow.md.
3. Read docs/engineering/issue-style-guide.md.
4. Read the target issue with:

   gh issue view 19

Treat the GitHub issue body as the source of truth. If this prompt conflicts with the issue, follow the issue and mention the conflict in your final summary.

Core intent:

Define the pure Go model for future objective calculation results.

This issue prepares SearchBench-Go for future Pkl-visible scoring formulas, but it must not add Pkl, evaluate formulas, change bundle serialization, or implement objective math.

The future flow is:

Go computes raw score evidence
  -> score.ScoreEvidenceDocument

Future Pkl computes visible objective math
  -> objective result with named intermediate values

Go validates objective result
  -> future objective.json

This issue only implements the pure objective result model and validation.

Repository context:

Issue #18 introduced objective-ready score evidence in the pure score/report layer.

Inspect:

internal/score/evidence.go
internal/score/evidence_test.go
internal/report/evidence.go
internal/report/evidence_test.go

Use the same architectural direction:

- score owns pure score/result models
- report projects report-shaped data into score evidence
- artifact writes files later
- artifact must not define this model
- Pkl must not be introduced yet

Important architecture constraints:

- Do not implement Pkl.
- Do not add Pkl schema files.
- Do not invoke Pkl.
- Do not implement expression parsing.
- Do not implement objective formula execution in Go.
- Do not add hidden transform registries.
- Do not change scoring reducers.
- Do not change graph-distance scoring.
- Do not change token-efficiency scoring.
- Do not load parent runs.
- Do not implement content-addressed lineage.
- Do not change bundle writer behavior.
- Do not add objective.json serialization yet.
- Do not import artifact from score.
- Do not import Eino, MCP, LangSmith, materialization, tree-sitter, or provider SDKs from score.

Suggested package placement:

internal/score/
  objective.go
  objective_test.go

If validation errors become nontrivial, this is also acceptable:

internal/score/
  objective.go
  objective_error.go
  objective_test.go

Model guidance:

Create pure types roughly equivalent to:

ObjectiveResult
  schema_version
  objective_id
  evidence_refs
  values
  final
  bounds optional

ObjectiveValue
  name
  value
  kind
  unit optional
  description optional

ObjectiveValueKind
  intermediate
  penalty
  final

ObjectiveEvidenceRef
  name
  bundle_path optional
  score_path optional
  report_path optional
  sha256 optional

ObjectiveBounds
  min optional
  max optional

Use existing repo naming conventions where they are better.

The model should support a future Pkl-shaped result like:

currentLocalizationQuality = 0.82
parentLocalizationQuality = 0.74
improvementVsParent = 0.08
tokenEfficiency = 0.91
base = 0.77
regressionPenalty = 1.0
invalidPredictionPenalty = 1.0
final = 0.77

These are ordinary data values. Do not encode the formula that produced them.

Design preference:

Prefer making the final objective score explicit and reviewable.

A reasonable shape is:

- Values contains all named values.
- Final identifies the final value by name, or otherwise explicitly represents the final value.
- Validation ensures the final value exists and is finite.
- If bounds are present, validation enforces them.

Do not overbuild a generic expression engine.

Validation expectations:

Add validation that rejects:

- missing schema version
- missing objective id
- empty objective value names
- duplicate objective value names
- missing final value
- final value not present in values, if final is represented by name
- NaN
- positive infinity
- negative infinity
- duplicate evidence ref names
- empty evidence ref names
- malformed evidence refs
- final value outside declared bounds

Evidence refs:

ObjectiveEvidenceRef should be typed data, not stringly hidden state.

A valid evidence ref should have a non-empty name and at least one useful locator, such as score path, report path, bundle path, or sha256.

Do not resolve evidence refs in this issue.

Do not read files.

Do not validate that paths exist.

Tests to add:

1. valid future-Pkl-shaped objective result validates
2. valid result can include current and parent evidence refs
3. named intermediate values are preserved
4. penalty values are preserved
5. final value is explicit
6. duplicate objective value names are rejected
7. missing final value is rejected
8. final value not present in values is rejected, if final is name-based
9. NaN is rejected
10. positive infinity is rejected
11. negative infinity is rejected
12. declared min bound is enforced
13. declared max bound is enforced
14. duplicate evidence ref names are rejected
15. empty evidence ref names are rejected
16. malformed evidence refs are rejected
17. score package does not import Pkl, Eino, MCP, LangSmith, artifact, materialization, tree-sitter, or provider SDKs for this feature

Use only pure unit tests.

Do not require filesystem bundles.

Do not require network access.

Do not require model calls.

Implementation notes:

- Use ordinary Go errors unless the existing score package already has a clear typed validation error pattern.
- Keep validation readable and small.
- Prefer simple constructors or Validate methods over framework-like abstractions.
- Keep JSON tags stable and future objective.json-friendly.
- Do not add objective.json writing to internal/artifact yet.
- Do not change existing score evidence behavior from issue #18.

Before finishing:

- Run gofmt on changed Go files.
- Run go test ./...
- If tests fail for reasons unrelated to your change, report that clearly.
- Keep the diff bounded to issue #19.

Final response format:

Summarize:

- files changed
- objective models added
- validation behavior added
- tests added
- commands run
- any deviations from the issue
- any follow-up issues exposed

Do not claim that Pkl, objective formula execution, objective.json bundle serialization, lineage, parent loading, graph-distance scoring, or token-efficiency scoring were implemented.
