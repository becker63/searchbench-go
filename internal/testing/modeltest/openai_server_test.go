package modeltest

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	openai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

func TestFakeOpenAIServerReturnsFixtureSuccessResponse(t *testing.T) {
	t.Parallel()

	server := NewFakeOpenAIServer(FakeResponse{
		Status: http.StatusOK,
		Body:   string(MustFixture("chat_completion_success.json")),
	})
	defer server.Close()

	response, err := http.Post(server.Server.URL+chatCompletionsPath, "application/json", strings.NewReader(`{"model":"test"}`))
	if err != nil {
		t.Fatalf("http.Post() error = %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if got, want := response.StatusCode, http.StatusOK; got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}
	if got, want := strings.TrimSpace(string(body)), strings.TrimSpace(string(MustFixture("chat_completion_success.json"))); got != want {
		t.Fatalf("body = %s, want %s", got, want)
	}
}

func TestFakeOpenAIServerReturnsFixtureErrorResponse(t *testing.T) {
	t.Parallel()

	server := NewFakeOpenAIServer(FakeResponse{
		Status: http.StatusTooManyRequests,
		Body:   string(MustFixture("chat_completion_error.json")),
	})
	defer server.Close()

	response, err := http.Post(server.Server.URL+chatCompletionsPath, "application/json", strings.NewReader(`{"model":"test"}`))
	if err != nil {
		t.Fatalf("http.Post() error = %v", err)
	}
	defer response.Body.Close()

	if got, want := response.StatusCode, http.StatusTooManyRequests; got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}
}

func TestFakeOpenAIServerRecordsRequestMethodPathAndBody(t *testing.T) {
	t.Parallel()

	server := NewFakeOpenAIServer(FakeResponse{
		Status: http.StatusOK,
		Body:   string(MustFixture("chat_completion_success.json")),
	})
	defer server.Close()

	requestBody := `{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hello"}]}`
	response, err := http.Post(server.Server.URL+chatCompletionsPath, "application/json", strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("http.Post() error = %v", err)
	}
	response.Body.Close()

	requests := server.Requests()
	if len(requests) != 1 {
		t.Fatalf("len(Requests()) = %d, want 1", len(requests))
	}
	if got, want := requests[0].Method, http.MethodPost; got != want {
		t.Fatalf("request method = %q, want %q", got, want)
	}
	if got, want := requests[0].Path, chatCompletionsPath; got != want {
		t.Fatalf("request path = %q, want %q", got, want)
	}
	if got, want := strings.TrimSpace(string(requests[0].Body)), requestBody; got != want {
		t.Fatalf("request body = %s, want %s", got, want)
	}
}

func TestFakeOpenAIServerReturnsDeterministicErrorWhenResponsesExhausted(t *testing.T) {
	t.Parallel()

	server := NewFakeOpenAIServer()
	defer server.Close()

	response, err := http.Post(server.Server.URL+chatCompletionsPath, "application/json", strings.NewReader(`{"model":"test"}`))
	if err != nil {
		t.Fatalf("http.Post() error = %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if got, want := response.StatusCode, http.StatusInternalServerError; got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}
	if got, want := strings.TrimSpace(string(body)), `{"error":{"message":"no scripted fake response remaining","type":"fixture_exhausted"}}`; got != want {
		t.Fatalf("body = %s, want %s", got, want)
	}
}

func TestDefaultTestsDoNotRequireAPIKeysOrExternalEndpoint(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_BASE_URL", "")

	server := NewFakeOpenAIServer(FakeResponse{
		Status: http.StatusOK,
		Body:   string(MustFixture("chat_completion_success.json")),
	})
	defer server.Close()

	chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		APIKey:  "test-key",
		BaseURL: server.BaseURL(),
		Model:   "gpt-4o-mini",
	})
	if err != nil {
		t.Fatalf("NewChatModel() error = %v", err)
	}

	message, err := chatModel.Generate(context.Background(), []*schema.Message{
		{Role: schema.User, Content: "Say hello"},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if got, want := message.Content, `{"files":["pkg/example.go"],"reasoning":"fixture"}`; got != want {
		t.Fatalf("message content = %q, want %q", got, want)
	}

	requests := server.Requests()
	if len(requests) != 1 {
		t.Fatalf("len(Requests()) = %d, want 1", len(requests))
	}
	if got, want := requests[0].Path, chatCompletionsPath; got != want {
		t.Fatalf("request path = %q, want %q", got, want)
	}

	host, _, err := net.SplitHostPort(strings.TrimPrefix(server.Server.URL, "http://"))
	if err != nil {
		t.Fatalf("SplitHostPort() error = %v", err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		t.Fatalf("server host = %q, want loopback address", host)
	}

	var requestPayload struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(requests[0].Body, &requestPayload); err != nil {
		t.Fatalf("Unmarshal(request) error = %v", err)
	}
	if got, want := requestPayload.Model, "gpt-4o-mini"; got != want {
		t.Fatalf("request model = %q, want %q", got, want)
	}
	if got, want := len(requestPayload.Messages), 1; got != want {
		t.Fatalf("len(request messages) = %d, want %d", got, want)
	}
}
