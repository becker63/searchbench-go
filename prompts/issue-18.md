You are working in the SearchBench-Go repository.

Task: implement the GitHub issue titled “Define objective-ready score evidence outside artifact serialization”.

First, read the project guidance:

1. Read AGENTS.md.
2. Read docs/engineering/agentic-development-flow.md.
3. Read docs/engineering/issue-style-guide.md.
4. Read the target issue with:

   gh issue view 18

Treat the GitHub issue body as the source of truth. If this prompt conflicts with the issue, follow the issue and mention the conflict in your final summary.

Core intent:

Move score evidence modeling out of internal/artifact and into the pure score/report model layer.

The bundle writer currently writes score.json, but internal/artifact should not define ScoreEvidence, MetricEvidence, or RoleCounts. Artifact code should serialize evidence; it should not own the scoring evidence contract.

This issue prepares the foundation for future Pkl-visible objective calculations, but must not add Pkl.

Important architecture constraints:

- Do not implement Pkl.
- Do not implement objective formulas.
- Do not implement objective result models.
- Do not implement lineage, parent run loading, or content addressing.
- Do not implement graph-distance scoring changes.
- Do not implement token-efficiency scoring changes.
- Do not introduce new reducers.
- Do not change evaluator, backend, MCP, IC, jCodeMunch, tree-sitter, or LangSmith behavior.
- Do not put score evidence types in internal/artifact.
- Do not move CandidateReport into score.
- Do not make score depend on artifact.
- Keep bundle behavior stable.

Desired dependency direction:

score/domain/run
  ↑
report projection
  ↑
artifact serialization

Not:

artifact defines scoring evidence

Repository context:

internal/artifact currently owns bundle serialization and writes:

- resolved.json
- report.json
- score.json
- metadata.json
- optional report.md / report.txt
- COMPLETE

Keep that behavior working.

Implementation guidance:

Create a pure score evidence model, likely in:

internal/score/evidence.go

Add tests in:

internal/score/evidence_test.go

If projection from CandidateReport belongs in report to avoid cycles, add something like:

internal/report/evidence.go
internal/report/evidence_test.go

A reasonable model shape is:

ScoreEvidenceDocument
  schema_version
  report_id
  systems
  run_counts
  failure_counts
  localization_distance
  usage
  regressions
  invalid_predictions
  metrics
  promotion_decision

RoleCounts
  baseline
  candidate

MetricEvidence
  metric
  direction
  baseline
  candidate
  delta
  improved
  regressed

UsageEvidence
  input_tokens
  output_tokens
  total_tokens
  cost_usd

RegressionEvidenceSummary
  count
  minor_count
  severe_count

Use existing project names and conventions where they are better than these suggestions.

The evidence document should be future Pkl-friendly. It should expose field-addressable evidence for later formulas such as:

current.localizationDistance.mean
current.usage.totalTokens
current.regressions.severeCount
current.invalidPredictions

Do not implement those formulas now.

Important behavior:

- CandidateReport should project into ScoreEvidenceDocument.
- Existing bundle writer should still write score.json.
- score.json should now be produced from the score/report-owned evidence model.
- Artifact code should consume the new evidence model rather than defining it.
- Existing deterministic bundle serialization tests should continue to pass.
- Missing optional evidence should be represented honestly.
- Do not pretend graph distance was computed if it was not.
- Do not pretend token usage was measured if it was unavailable.

Tests to add or update:

1. CandidateReport projects into a score evidence document.
2. metric comparisons appear in evidence with stable values.
3. metric directions are preserved.
4. improved/regressed flags are preserved or recomputed consistently.
5. baseline/candidate run counts are correct.
6. baseline/candidate failure counts are correct.
7. promotion decision is preserved.
8. regressions are preserved or summarized.
9. usage totals are included when runs have usage summaries.
10. deterministic JSON serialization still holds for score.json.
11. artifact package no longer defines score evidence types.
12. artifact package still writes score.json.
13. no Pkl imports are introduced.
14. no Eino, MCP, LangSmith, materialization, or tree-sitter imports are introduced into score evidence code.

Use fake/local report data only.

Do not hit the network.

Do not require real model calls.

Do not overbuild a generic expression engine, Pkl runtime, or objective calculator.

Before finishing:

- Run gofmt on changed Go files.
- Run go test ./...
- If tests fail for reasons unrelated to your change, report that clearly.
- Keep the diff bounded to this issue.

Final response format:

Summarize:

- files changed
- where score evidence now lives
- how CandidateReport projects into evidence
- how artifact serialization changed
- tests added or updated
- commands run
- any deviations from the issue
- any follow-up issues exposed

Do not claim Pkl, objective calculation, lineage, content-addressing, graph-distance implementation, real backend execution, or tracing are implemented.
