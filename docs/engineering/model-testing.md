# SearchBench-Go Model Testing

## Purpose

SearchBench-Go has deterministic model-testing infrastructure under:

```text
internal/testing/
internal/testing/modeltest/
```

This infrastructure exists so agents can build model-backed execution code without:

- real provider calls
- API-key dependencies
- external network access in default tests
- paid model usage
- ad hoc mock systems in production packages

These helpers are test infrastructure only. They are not production architecture.

## Three tiers

SearchBench-Go uses three model-testing tiers.

### Tier 1: scripted Eino model

Use `internal/testing/modeltest.ScriptedModel` for most evaluator tests.

This helper implements Eino's `model.ToolCallingChatModel` and works directly with:

- `github.com/cloudwego/eino/components/model`
- `github.com/cloudwego/eino/schema`

It provides:

- deterministic `Generate` responses in order
- deterministic `Stream` chunks in order
- scripted model errors
- recorded calls for assertions
- recorded tool bindings via `WithTools`

This is the default path for:

- evaluator-loop tests
- multi-turn evaluator agent-loop tests
- multi-tool-call evaluator tests
- prompt-to-model integration tests
- malformed final-output tests
- empty-prediction tests
- failure-classification tests
- fake-tool orchestration tests

### Tier 2: fake OpenAI-compatible server

Use `internal/testing/modeltest.FakeOpenAIServer` only when the test needs to prove provider-boundary wiring.

Current scope:

- local-only `httptest` server
- `POST /v1/chat/completions`
- scripted responses by request order
- recorded method, path, body, and headers
- deterministic exhaustion error

This is appropriate for:

- Eino OpenAI adapter smoke tests
- BaseURL override tests
- request-shape recording tests

This is not a full OpenAI mock.

### Tier 3: live model tests

Live model tests are allowed only as explicit opt-in tests.

They must never run in the default suite.

Examples:

```text
SEARCHBENCH_RUN_LIVE_MODEL_TESTS=1 go test ./...
go test -tags=live ./...
```

If a test can spend money, it must be opt-in.

## Package inventory

Current package layout:

```text
internal/testing/doc.go
internal/testing/modeltest/doc.go
internal/testing/modeltest/fixtures.go
internal/testing/modeltest/scripted_model.go
internal/testing/modeltest/openai_server.go
internal/testing/modeltest/testdata/
```

### `fixtures.go`

Provides:

- `Fixture(name string) ([]byte, error)`
- `MustFixture(name string) []byte`

Use this for scrubbed provider-response fixtures stored in package `testdata`.

### `scripted_model.go`

Provides:

- `NewScriptedModel(responses ...ScriptedResponse) *ScriptedModel`
- `ScriptedResponse`
- `ScriptedCall`
- `(*ScriptedModel).Calls() []ScriptedCall`
- `ErrNoScriptedResponses`

`ScriptedResponse` supports:

- `Message *schema.Message`
- `Stream []*schema.Message`
- `Err error`

Important behavior:

- each `Generate` or `Stream` call consumes exactly one scripted response
- calls are recorded before response consumption
- `WithTools` returns a new view over the same scripted state
- tool bindings are recorded on each call
- exhaustion returns `ErrNoScriptedResponses`

### `openai_server.go`

Provides:

- `NewFakeOpenAIServer(responses ...FakeResponse) *FakeOpenAIServer`
- `(*FakeOpenAIServer).BaseURL() string`
- `(*FakeOpenAIServer).Requests() []RecordedRequest`
- `(*FakeOpenAIServer).Close()`
- `DecodeRecordedRequest`

`FakeResponse` supports:

- `Status`
- `Body`
- `Header`

`RecordedRequest` captures:

- `Method`
- `Path`
- `Body`
- `Header`

Important behavior:

- binds to local loopback only
- records every request before responding
- returns scripted responses in order
- returns a deterministic fixture-exhaustion error when responses run out
- does not proxy or fall through to any real provider

## Default test contract

Default tests for model-backed code must remain:

- offline
- deterministic
- cheap
- safe to run repeatedly

Default tests must not require:

- `OPENAI_API_KEY`
- `ANTHROPIC_API_KEY`
- `OPENROUTER_API_KEY`
- external network access
- live provider endpoints

When a new agent/executor test is possible with Tier 1, use Tier 1.

Do not jump to Tier 2 or Tier 3 unless the test is specifically about those boundaries.

## Guidance for the next two issues

### 1. `Define LCA localization domain schema from the Python SearchBench contract`

This issue should stay pure domain work.

It should:

- define typed domain models
- define path normalization
- define validation and parsing behavior
- use in-memory table tests only

It should not import:

- `internal/testing/modeltest`
- Eino
- provider adapters
- `httptest`

Reason:

The domain-schema issue is a prerequisite for execution code, not execution code itself. Its tests should prove purity and deterministic normalization, not model behavior.

What this issue may rely on from the testing strategy:

- the repository already has a strict rule that default model tests are offline
- future execution-layer tests already have a home and should not push model fixtures into `internal/domain`

What this issue should prepare for:

- a canonical prediction type that the Eino evaluator can return
- typed execution phases and failure kinds that evaluator tests can assert against
- stable path normalization shared by finalization tests later

### 2. `Prove the minimal Eino agent loop`

This issue should consume `internal/testing/modeltest` directly.

Default test strategy:

- use `ScriptedModel` for evaluator logic tests
- use local fake tools for deterministic tool execution
- use typed domain result assertions

Optional provider-boundary test:

- use `FakeOpenAIServer` only for one narrow Eino OpenAI adapter smoke test if needed

Recommended coverage shape:

- success final prediction via `ScriptedModel`
- success with multiple model turns and tool calls inside one evaluator run
- configured `SystemSpec.Runtime.MaxSteps` bounds evaluator turns through the runtime
- malformed final output via `ScriptedModel`
- empty predicted files via `ScriptedModel`
- evaluator/model error via `ScriptedModel`
- tool failure with deterministic local fake tool
- evaluator retry attempts that are distinct from model/tool turns inside one run
- optional tool-call-shaped scripted model responses if the runner uses them cleanly

The minimal Eino issue should not:

- create another test fixture package
- create a production model abstraction to hide Eino
- move these helpers into executor packages
- call real model APIs in default tests

## Recommended usage patterns

### Tier 1 evaluator test

Use this shape when the test is about runner behavior:

```go
model := modeltest.NewScriptedModel(
	modeltest.ScriptedResponse{
		Message: &schema.Message{
			Role:    schema.Assistant,
			Content: `{"predicted_files":["src/main.kt"]}`,
		},
	},
)
```

Then assert:

- runner result
- phase classification
- failure kind
- model call count
- prompt/message content

### Tier 2 adapter smoke test

Use this shape only when the test is about the OpenAI-compatible adapter boundary:

```go
server := modeltest.NewFakeOpenAIServer(modeltest.FakeResponse{
	Status: http.StatusOK,
	Body:   string(modeltest.MustFixture("chat_completion_success.json")),
})
defer server.Close()
```

Then configure Eino OpenAI with:

- fake API key
- `BaseURL: server.BaseURL()`
- fixed model name

Then assert:

- request path
- request method
- request body contains the expected high-signal fields

## Anti-patterns

Do not:

- put model fixtures in `internal/domain`, `internal/run`, `internal/score`, or `internal/compare`
- introduce a generic provider abstraction just for tests
- add real provider calls to default tests
- copy raw provider dumps with secrets or giant payloads
- assert brittle JSON formatting details
- make Tier 2 the main path for evaluator logic tests

## Why this matters

The next execution issues should be able to move quickly because:

- Eino types are already present in test infrastructure
- provider-boundary smoke testing already has a narrow local harness
- offline deterministic defaults are already enforced by example
- execution code can focus on prompting, tooling, finalization, and typed failures instead of rebuilding test scaffolding
