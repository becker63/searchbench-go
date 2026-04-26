package modeltest

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
)

const chatCompletionsPath = "/v1/chat/completions"

// FakeResponse defines one scripted HTTP response.
type FakeResponse struct {
	Status int
	Body   string
	Header map[string]string
}

// RecordedRequest captures one received HTTP request.
type RecordedRequest struct {
	Method string
	Path   string
	Body   []byte
	Header http.Header
}

// FakeOpenAIServer is a tiny local OpenAI-compatible fixture server.
type FakeOpenAIServer struct {
	Server *httptest.Server

	mu        sync.Mutex
	requests  []RecordedRequest
	responses []FakeResponse
}

// NewFakeOpenAIServer starts a local fixture server for chat completions.
func NewFakeOpenAIServer(responses ...FakeResponse) *FakeOpenAIServer {
	server := &FakeOpenAIServer{
		responses: append([]FakeResponse(nil), responses...),
	}

	mux := http.NewServeMux()
	mux.HandleFunc(chatCompletionsPath, server.handleChatCompletions)

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	unstarted := httptest.NewUnstartedServer(mux)
	unstarted.Listener = listener
	unstarted.Start()
	server.Server = unstarted
	return server
}

// BaseURL returns the OpenAI-compatible base URL for Eino OpenAI adapters.
func (s *FakeOpenAIServer) BaseURL() string {
	return s.Server.URL + "/v1"
}

// Close shuts down the local test server.
func (s *FakeOpenAIServer) Close() {
	if s != nil && s.Server != nil {
		s.Server.Close()
	}
}

// Requests returns a copy of all recorded requests.
func (s *FakeOpenAIServer) Requests() []RecordedRequest {
	s.mu.Lock()
	defer s.mu.Unlock()

	requests := make([]RecordedRequest, len(s.requests))
	for i, request := range s.requests {
		requests[i] = RecordedRequest{
			Method: request.Method,
			Path:   request.Path,
			Body:   slices.Clone(request.Body),
			Header: request.Header.Clone(),
		}
	}
	return requests
}

func (s *FakeOpenAIServer) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":{"message":"failed to read request body"}}`, http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	s.requests = append(s.requests, RecordedRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Body:   slices.Clone(body),
		Header: r.Header.Clone(),
	})

	response, ok := s.dequeueLocked()
	s.mu.Unlock()
	if !ok {
		writeJSON(w, http.StatusInternalServerError, `{"error":{"message":"no scripted fake response remaining","type":"fixture_exhausted"}}`, nil)
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, `{"error":{"message":"method not allowed"}}`, nil)
		return
	}

	writeJSON(w, response.Status, response.Body, response.Header)
}

func (s *FakeOpenAIServer) dequeueLocked() (FakeResponse, bool) {
	if len(s.responses) == 0 {
		return FakeResponse{}, false
	}

	response := s.responses[0]
	s.responses = s.responses[1:]
	return response, true
}

func writeJSON(w http.ResponseWriter, status int, body string, header map[string]string) {
	for key, value := range header {
		w.Header().Set(key, value)
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}

// DecodeRecordedRequest unmarshals a recorded JSON request body into dst.
func DecodeRecordedRequest[T any](request RecordedRequest, dst *T) error {
	return json.Unmarshal(request.Body, dst)
}
