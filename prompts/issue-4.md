You are working in the searchbench-go repository.

The target GitHub issue is:

    Prove the minimal Eino agent loop

Before making any changes, use the GitHub CLI to find and read the full target issue.

Run:

    gh issue list --state open --search "Prove the minimal Eino agent loop"
    gh issue view <ISSUE_NUMBER> --comments --json title,body,comments,labels,state,url

If multiple issues match, choose the open issue whose title exactly matches:

    Prove the minimal Eino agent loop

If no issue matches, stop and report that the target issue could not be found. Do not implement from memory.

Treat the GitHub issue body as the source of truth for the implementation contract.

If this prompt conflicts with the GitHub issue, prefer the GitHub issue unless it clearly violates AGENTS.md/agents.md or the repository architecture docs. In that case, stop and explain the conflict before editing.

Before editing, read the project guidance and architecture docs:

1. AGENTS.md or agents.md, whichever exists
2. docs/engineering/issue-style-guide.md
3. docs/engineering/model-testing.md
4. integration-shape.md
5. todo.md
6. Existing package docs, especially:
   - internal/domain/doc.go
   - internal/run/doc.go
   - internal/compare/doc.go
   - internal/backend/doc.go
   - internal/score/doc.go
   - internal/testing/doc.go
   - internal/testing/modeltest/doc.go

Also inspect the existing implementation from prior issues:

- internal/testing/modeltest/
- internal/domain/localization.go
- internal/domain/localization_test.go
- internal/domain/prediction.go
- internal/domain/task.go
- internal/domain/repo.go
- internal/run/phases.go
- internal/run/failure.go
- internal/run/record.go

Do not start editing until you understand the existing package boundaries.

Important dependency context:
The deterministic model fixture testing issue is already done.

Use:

    internal/testing/modeltest.ScriptedModel

for default evaluator tests.

Use:

    internal/testing/modeltest.FakeOpenAIServer

only for an optional narrow Tier 2 provider-boundary smoke test.

Do not create a second model fixture system.

Important LCA domain context:
The LCA localization domain-schema issue is already done.

Before designing any evaluator input/result types, inspect and reuse the existing Hugging Face localization domain models in:

    internal/domain/localization.go
    internal/domain/localization_test.go
    internal/domain/prediction.go
    internal/domain/task.go
    internal/domain/repo.go

The evaluator should consume the existing LCA/domain model instead of inventing a duplicate localization task schema.

Use existing domain types such as:

- domain.LCATask
- domain.LCATaskIdentity
- domain.LCAContext
- domain.LCAGold
- domain.LCAHFRow
- domain.TaskSpec
- domain.TaskInput
- domain.TaskOracle
- domain.BenchmarkLCA
- domain.RepoSnapshot
- domain.RepoRelPath
- domain.Prediction
- domain.CanonicalizePath
- domain.CanonicalizePaths

The LCA schema encodes the Hugging Face localization contract:

- repo_owner
- repo_name
- base_sha
- issue_title
- issue_body
- changed_files
- optional issue_url / pull_url / diff metadata

The evaluator prompt should use model-visible issue/context fields from LCATask or TaskSpec.

Gold labels must not be included in model-visible prompt content.

Do not expose changed_files, TaskOracle.GoldFiles, LCAGold, or other oracle/scorer-only data to the evaluator prompt.

The finalizer should normalize predicted paths using existing domain path normalization helpers.

Do not reimplement:

- LCATaskIdentity
- LCAContext
- LCAGold
- LCATask
- LCAHFRow
- CanonicalizePath
- CanonicalizePaths

If these types need minor additions to support this issue, keep changes small and justify them in the final summary.

Task:
Implement the issue “Prove the minimal Eino agent loop”.

Goal:
Implement the smallest possible SearchBench-Go evaluator loop using Eino, without MCP, Iterative Context, jCodeMunch, repo materialization, policy injection, graph scoring, writer agents, or optimization.

This issue proves that SearchBench-Go can:

- run an evaluator-style agent through Eino
- expose one deterministic local tool
- render a typed XML-style prompt document
- receive a structured file-localization prediction
- classify evaluator failures
- return a typed execution result through the existing domain/run model

Architecture rule:
SearchBench-Go owns:

- task/domain models
- prompt inputs
- prompt rendering boundaries
- final prediction normalization
- execution result typing
- error/failure typing
- lightweight lifecycle phase classification

Eino owns:

- model execution
- tool calling
- intermediate model/tool loop behavior

The modeltest package owns:

- scripted model responses
- fake provider responses
- request recording for provider-boundary tests
- deterministic test behavior

Do not introduce:

- custom ReAct loops
- custom state machines
- MCP-specific code
- backend adapter abstractions
- policy machinery
- writer/repair machinery
- scoring
- live model calls in default tests
- new model testing infrastructure
- duplicate LCA task/domain models

Preserve the old SearchBench lifecycle discipline:

- explicit phases
- typed outcomes
- side effects isolated to execution phases
- clear failure classification
- cold construction
- one explicit Run-style entrypoint

But do not port the Python state machine.

Use ordinary Go control flow.

Expected phase flow:

    Run(ctx, task)
      create execution result accumulator

      phase: render_prompt
          build typed evaluator prompt input from existing LCA/domain task data
          render .templ XML-style prompt
          ensure oracle/gold fields are not included in prompt text
          if render failed:
              return prompt_render_failed

      phase: run_evaluator
          run Eino evaluator using Tier 1 scripted model fixture in tests
          expose one fake local tool
          if evaluator/model failed:
              return evaluator_failed
          if fake tool failed:
              return tool_call_failed

      phase: finalize_prediction
          parse final structured JSON response
          validate predicted_files/files
          normalize paths with domain.CanonicalizePaths
          if malformed:
              return finalization_failed
          if predicted files empty:
              return invalid_prediction

      phase: complete
          return successful execution result

Phase names should appear in code, result metadata, logs, or typed failure values so later issues can build on the same lifecycle language.

Minimal phases for this issue:

    render_prompt
    run_evaluator
    finalize_prediction
    complete

Failure kinds for this issue:

    prompt_render_failed
    evaluator_failed
    tool_call_failed
    finalization_failed
    invalid_prediction

Use names that fit existing run/domain conventions if equivalent constants already exist.

Prompt rendering:
SearchBench-Go uses templ files to render XML-style prompt documents.

Implement an evaluator prompt package shaped like:

    internal/prompts/evaluator/
      input.go
      prompt.templ
      render.go
      prompt_test.go

If the repo has a better prompt convention, follow it, but keep prompts separate from executor logic.

The .templ file is a contract boundary.

Prompt structure lives in .templ.
Prompt data shape lives in typed Go input structs.
The renderer produces a plain string for Eino/model usage.

Prompt documents may use XML-style sections such as:

    <searchbench_prompt>
    <role>
    <task>
    <issue>
    <repo>
    <available_tools>
    <constraints>
    <output_contract>

Dynamic issue/repo/user content should use escaped templ expressions.

Do not use raw rendering unless the content is trusted and the test explains why.

Final model outputs are not XML. Final outputs should be strict JSON.

Evaluator prompt input:
Create a typed input struct, for example:

    Input
      TaskID
      RepoName
      RepoSHA
      IssueTitle
      IssueBody
      AllowedTools
      OutputSchemaJSON
      Constraints

Prefer deriving this input from existing domain.LCATask or domain.TaskSpec.

Gold files must not be included in the prompt input.

Add a test that would fail if the rendered evaluator prompt includes:

- changed_files
- gold_files
- oracle
- a known gold path from the fixture task

Finalization:
The evaluator should parse strict final JSON into the existing domain.Prediction or a very small adapter-local finalization DTO that immediately converts into domain.Prediction.

The canonical evaluator output is file localization.

Accept JSON shaped around either:

    {"predicted_files":["src/file.go"],"reasoning":"..."}

or, if the existing domain model strongly prefers it:

    {"files":["src/file.go"],"reasoning":"..."}

Pick one output contract and test it.

Normalize predicted paths using existing domain.CanonicalizePaths.

Return domain.Prediction with normalized domain.RepoRelPath files.

Do not add scoring fields.

Do not add graph distance.

Do not add benchmark hit rate.

Do not add symbol ranking.

Execution result:
Prefer integrating with the existing run lifecycle types instead of inventing a parallel runner model.

Inspect:

    internal/run/

If there is already a suitable result/failure envelope, use it.

If a small evaluator-specific result type is necessary, keep it inside the executor/eino package and ensure it can later project into run.ExecutedRun or run.RunFailure.

The result must distinguish:

- successful prediction
- prompt render failure
- evaluator/model failure
- tool failure
- malformed final output
- empty predicted files
- unexpected internal failure

Failures should include:

- phase
- kind
- message
- cause where useful
- recoverable where useful if already consistent with repo style

Package placement:
Prefer:

    internal/prompts/evaluator/
      input.go
      prompt.templ
      render.go
      prompt_test.go

    internal/executor/eino/
      evaluator.go
      finalizer.go
      errors.go
      phases.go
      evaluator_test.go

If the repo already has a stronger implementation/adapters split, follow it.

Do not move core domain types into the Eino package.

Eino should remain an implementation detail.

Do not put prompt templates inside the executor package unless the existing project structure strongly prefers that.

Testing:
Default tests must be offline, deterministic, and safe to run repeatedly.

They must not require:

- API keys
- network access
- real OpenAI/OpenRouter/Anthropic calls
- paid model usage

Use Tier 1 scripted model fixtures from:

    internal/testing/modeltest

Required test scenarios:

1. evaluator prompt renders from typed input
2. rendered prompt contains expected XML-style sections
3. rendered prompt includes model-visible LCA issue/context fields
4. rendered prompt does not include gold changed_files / oracle data
5. evaluator runner construction is cold
6. evaluator Run records or returns named phases in order
7. evaluator can use one deterministic fake local tool
8. success final prediction produces a successful result with normalized predicted files
9. empty predicted files is rejected
10. malformed final output is rejected
11. scripted model error is surfaced as evaluator_failed
12. fake tool failure is surfaced as tool_call_failed
13. failures include phase and failure kind
14. default tests do not require API keys or real network access

The fake tool may be extremely small.

Examples:

- fake_resolve
- fake_search_file
- inspect_fake_repo
- lookup_symbol

The fake tool should return deterministic fixture data.

It does not need to inspect a real repository.

Fake implementation policy:
Fake tools are allowed only to prove the evaluator execution seam.

Fake tools must remain test-local or clearly marked as fixtures.

They must not become production backend abstractions.

Do not introduce a fake backend package.

Do not shape future MCP/IC/JCM architecture around the fake tool.

Optional Tier 2:
A Tier 2 fake OpenAI-compatible server smoke test is optional.

Only add it if it is cheap and proves a specific provider-boundary fact, such as:

- Eino OpenAI-compatible adapter can be configured with fake BaseURL
- fake OpenAI-compatible server receives the rendered evaluator prompt

If included, it must use:

    internal/testing/modeltest.FakeOpenAIServer

It must not call a real provider.

Do not make Tier 2 the default path for evaluator logic tests.

Important non-goals:
Do not implement:

- MCP
- Iterative Context
- jCodeMunch
- repo materialization
- policy artifacts
- policy injection
- graph-distance scoring
- writer agent
- optimization loop
- writer repair loop
- external validation pipeline
- LangSmith tracing
- GitHub issue automation
- full CLI command surface
- custom state machine library
- retry machinery
- real model API calls in default tests
- new model testing infrastructure
- duplicate LCA schema
- duplicate domain prediction model unless unavoidable
- Hugging Face dataset loading
- dataset caching

Acceptance criteria:

- There is a minimal Eino-based evaluator execution path.
- The evaluator runner has one explicit Run-style entrypoint.
- Runner construction is cold; no model/tool execution happens during construction.
- The implementation uses named phases such as render_prompt, run_evaluator, finalize_prediction, and complete.
- Failures record the phase where they occurred.
- The result distinguishes prompt render failure, evaluator failure, tool failure, finalization failure, and invalid prediction.
- This is implemented with ordinary Go control flow, not a state machine library.
- The evaluator prompt is rendered from a .templ XML-style prompt document.
- The prompt has a typed Go input struct.
- The prompt input is derived from existing LCA/domain task types.
- The rendered prompt contains clear XML-style sections such as <role>, <task>, <issue>, and <output_contract>.
- The rendered prompt includes model-visible issue/context fields.
- The rendered prompt does not include gold changed_files / oracle data.
- There is a render test for the evaluator prompt.
- The evaluator can call one deterministic local fake tool.
- The fake tool is test-local or clearly marked as a fixture.
- The evaluator returns domain.Prediction or an existing equivalent.
- Predicted files are required and cannot be empty on success.
- Predicted files are normalized with existing domain path normalization helpers.
- Invalid/malformed evaluator output produces a typed failure.
- Prompt render failure produces a typed failure.
- Tool failure produces a typed failure.
- Default tests use Tier 1 scripted model fixtures from the prior testing issue.
- Default tests do not require API keys.
- Default tests do not call real model APIs.
- Default tests do not require external network access.
- Any Tier 2 test uses the fake OpenAI-compatible httptest server from the prior issue.
- The implementation does not introduce MCP-specific code.
- The implementation does not introduce backend adapter abstractions.
- The implementation does not introduce a custom agent loop/state machine.
- The implementation does not duplicate the LCA schema from issue #3.
- Tests cover the happy path and at least one invalid-output path.
- Package placement follows the existing SearchBench-Go architecture.

Implementation guidance:
Keep this boring and small.

Do not overbuild.

Do not create a general evaluator framework.

Do not create a generic provider abstraction.

Do not create a large DSL.

Do not attempt real MCP/tool runtime integration.

The goal is the first executable evaluator slice, not the final evaluator architecture.

Before handing off, run:

    gofmt -w .
    go test ./...
    go mod tidy
    go test ./...

Then summarize:

- issue number and title implemented
- files changed
- prompt package added
- executor/eino package added or modified
- existing LCA/domain types reused
- tests added
- how Tier 1 modeltest helpers were used
- whether any optional Tier 2 test was added
- any deliberate deviations from the issue
- confirmation that no real model/provider/network calls were added
- confirmation that no MCP/backend/policy/scoring work was added
- confirmation that no duplicate LCA schema was introduced
