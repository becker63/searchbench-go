You are working in the searchbench-go repository.

  The target GitHub issue is #3:

      Define deterministic model fixture testing for Eino/OpenAI-compatible calls

  Before making any changes, read the project guidance and architecture docs:

  1. AGENTS.md
  2. docs/engineering/issue-style-guide.md
  3. integration-shape.md
  4. todo.md
  5. Existing package docs, especially:
     - internal/domain/doc.go
     - internal/run/doc.go
     - internal/compare/doc.go
     - internal/backend/doc.go
     - internal/score/doc.go

  Also use the GitHub CLI to find and read the full target GitHub issue.

  Run:

      gh issue list --state open --search "Define deterministic model fixture testing for
  Eino/OpenAI-compatible calls"
      gh issue view <ISSUE_NUMBER> --comments --json title,body,comments,labels,state,url

  If multiple issues match, choose the open issue whose title exactly matches:

      Define deterministic model fixture testing for Eino/OpenAI-compatible calls

  Treat the GitHub issue body as the source of truth for the implementation contract.

  If no issue matches, stop and report that the target issue could not be found. Do not
  implement from memory.

  If this prompt conflicts with the GitHub issue, prefer the GitHub issue unless it
  clearly violates AGENTS.md or the repository architecture docs. In that case, stop and
  explain the conflict before editing.

  Do not start editing until you understand the existing package boundaries.

  Task: implement the issue “Define deterministic model fixture testing for Eino/OpenAI-
  compatible calls”.

  Goal:
  Create deterministic, reusable testing utilities for model-backed SearchBench-Go code
  without using real paid model APIs in the default test suite.

  This issue exists to prevent tests from accidentally depending on real OpenAI/
  OpenRouter/Anthropic calls, API keys, network access, or paid model usage.

  Core decisions to preserve:

  - SearchBench-Go uses three model testing tiers.
  - Tier 1 is an in-process scripted Eino-compatible model helper.
  - Tier 2 is a tiny fake OpenAI-compatible HTTP server using Go stdlib httptest.
  - Tier 3 is live model testing, opt-in only.
  - Default tests must never call real APIs.
  - These helpers are test infrastructure, not production architecture.
  - Do not introduce production model/provider abstractions unless the existing code
  already requires one.

  Architecture constraints:

  - Do not put these helpers in core domain/run/score/report/compare packages.
  - Do not leak test fixture types into production code.
  - Do not create a generic model framework.
  - Do not use Swagger, OpenAPI-generated mocks, Gin, Echo, chi, or another web
  framework.
  - Do not use map[string]any unless it is truly unavoidable and adapter/test-local.
  - Prefer typed structs and json.RawMessage at narrow fixture boundaries.
  - Keep everything boring, explicit, and small.

  Preferred package placement:

      internal/testing/modeltest/
        scripted_model.go
        scripted_model_test.go
        openai_server.go
        openai_server_test.go
        fixtures.go

  If the repo already has a stronger test utility convention, follow it, but do not place
  this in production domain packages.

  Implement Tier 1: scripted model helper.

  The helper should:

  - Return deterministic scripted responses in order.
  - Record calls for assertions.
  - Support model errors.
  - Be small and shaped around the Eino interfaces actually used or expected.
  - Avoid over-abstracting.
  - If exact Eino interfaces are not yet available in the repo, create the smallest local
  test helper that can later be adapted, and document the seam clearly.

  Conceptual behavior:

      ScriptedModel
        responses
        calls
        Generate(ctx, messages, options...)
        Stream(ctx, messages, options...) if practical

  Required capabilities:

  - successful final text/JSON response
  - malformed final response
  - empty response
  - model error
  - call recording

  Implement Tier 2: fake OpenAI-compatible server.

  Use only:

  - net/http
  - net/http/httptest
  - http.ServeMux or a tiny custom handler

  The fake server should:

  - Be fixture-driven.
  - Return scripted responses by request order.
  - Record method, path, body, and headers.
  - Return deterministic error if no scripted response remains.
  - Support success and provider-error fixtures.
  - Bind only locally through httptest.
  - Never proxy to a real provider.

  Suggested shape:

      type FakeOpenAIServer struct {
          Server    *httptest.Server
          Requests  []RecordedRequest
          Responses []FakeResponse
      }

      type FakeResponse struct {
          Status int
          Body   string
          Header map[string]string
      }

      type RecordedRequest struct {
          Method string
          Path   string
          Body   []byte
          Header http.Header
      }

  Adjust names and details to fit the project’s style.

  The fake server should support the endpoint family actually needed by tests. If no Eino
  adapter is wired yet, support Chat Completions as the initial narrow shape:

      POST /v1/chat/completions

  Do not implement a full OpenAI API mock.

  Fixture policy:

  - Fixture files should be small and scrubbed.
  - Do not commit real prompts, API keys, secrets, large traces, or raw provider dumps.
  - Fixture names should describe behavior.
  - Use testdata under the modeltest package.

  Suggested fixtures:

      internal/testing/modeltest/testdata/
        chat_completion_success.json
        chat_completion_error.json

  Only add more fixtures if needed by tests.

  Request assertion policy:

  Good assertions:

  - request method is POST
  - request path is the configured endpoint
  - request body is recorded
  - response order is deterministic

  Avoid brittle assertions:

  - exact JSON field ordering
  - full prompt snapshots
  - incidental provider fields
  - broad OpenAI API completeness

  Tests to add:

  1. scripted model returns responses in order
  2. scripted model records calls
  3. scripted model can return a model error
  4. fake OpenAI-compatible server returns fixture success response
  5. fake OpenAI-compatible server returns fixture error response
  6. fake server records request method and path
  7. fake server records request body
  8. fake server returns deterministic error when scripted responses are exhausted
  9. default tests do not require API keys
  10. default tests do not call a real external endpoint

  Important non-goals:

  Do not implement:

  - minimal Eino evaluator loop
  - MCP
  - Iterative Context
  - jCodeMunch
  - repo materialization
  - policy artifacts
  - policy injection
  - graph-distance scoring
  - writer agent
  - optimization loop
  - LangSmith tracing
  - GitHub issue automation
  - live model tests enabled by default
  - a full OpenAI API mock
  - Swagger/OpenAPI mock generation
  - third-party HTTP framework
  - production model abstraction

  Acceptance criteria:

  - There is a documented three-tier model testing strategy.
  - Default tests do not require API keys.
  - Default tests do not call real model APIs.
  - Default tests do not require external network access.
  - A Tier 1 scripted model helper exists.
  - The scripted model returns deterministic responses.
  - The scripted model records calls for assertions.
  - A Tier 2 fake OpenAI-compatible httptest server exists.
  - The fake server uses net/http and net/http/httptest.
  - The fake server does not use Swagger/OpenAPI-generated code.
  - The fake server does not use a third-party web framework.
  - The fake server returns fixture responses by request order.
  - The fake server records method, path, headers, and body.
  - The fake server supports at least one success response and one provider error
  response.
  - The fake server returns deterministic error when no scripted response remains.
  - Fixture files are small and scrubbed.
  - Any live model test path is opt-in only if added at all.
  - Testing utilities do not leak into production domain packages.

  Be especially careful not to invent a parallel model system. This repository already
  has strong core nouns. The testing helper should be a narrow utility for later
  executor/evaluator tests.

  Before handing off, run:

      gofmt -w .
      go test ./...
      go mod tidy

  Then summarize:

  - files changed
  - package created
  - tests added
  - any deliberate deviations from the issue
  - whether any Eino interface mismatch required a smaller local seam
