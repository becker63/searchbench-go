package modeltest

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestScriptedModelReturnsResponsesInOrder(t *testing.T) {
	t.Parallel()

	model := NewScriptedModel(
		ScriptedResponse{
			Message: &schema.Message{Role: schema.Assistant, Content: `{"files":["pkg/first.go"]}`},
		},
		ScriptedResponse{
			Message: &schema.Message{Role: schema.Assistant, Content: `{"files":["pkg/second.go"]}`},
		},
	)

	first, err := model.Generate(context.Background(), []*schema.Message{
		{Role: schema.User, Content: "first"},
	})
	if err != nil {
		t.Fatalf("first Generate() error = %v", err)
	}
	second, err := model.Generate(context.Background(), []*schema.Message{
		{Role: schema.User, Content: "second"},
	})
	if err != nil {
		t.Fatalf("second Generate() error = %v", err)
	}

	if got, want := first.Content, `{"files":["pkg/first.go"]}`; got != want {
		t.Fatalf("first response = %q, want %q", got, want)
	}
	if got, want := second.Content, `{"files":["pkg/second.go"]}`; got != want {
		t.Fatalf("second response = %q, want %q", got, want)
	}
}

func TestScriptedModelRecordsCalls(t *testing.T) {
	t.Parallel()

	model := NewScriptedModel(
		ScriptedResponse{
			Message: &schema.Message{Role: schema.Assistant, Content: "ok"},
		},
	)

	withTools, err := model.WithTools([]*schema.ToolInfo{{
		Name: "resolve_file",
		Desc: "Resolve a file candidate.",
	}})
	if err != nil {
		t.Fatalf("WithTools() error = %v", err)
	}

	_, err = withTools.Generate(context.Background(), []*schema.Message{
		{Role: schema.System, Content: "system prompt"},
		{Role: schema.User, Content: "user prompt"},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	calls := model.Calls()
	if len(calls) != 1 {
		t.Fatalf("len(Calls()) = %d, want 1", len(calls))
	}
	if got, want := calls[0].Method, "generate"; got != want {
		t.Fatalf("call method = %q, want %q", got, want)
	}
	if got, want := len(calls[0].Messages), 2; got != want {
		t.Fatalf("len(call messages) = %d, want %d", got, want)
	}
	if got, want := calls[0].Messages[1].Content, "user prompt"; got != want {
		t.Fatalf("call user prompt = %q, want %q", got, want)
	}
	if got, want := len(calls[0].Tools), 1; got != want {
		t.Fatalf("len(call tools) = %d, want %d", got, want)
	}
	if got, want := calls[0].Tools[0].Name, "resolve_file"; got != want {
		t.Fatalf("tool name = %q, want %q", got, want)
	}
}

func TestScriptedModelReturnsModelError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("provider unavailable")
	model := NewScriptedModel(ScriptedResponse{Err: wantErr})

	_, err := model.Generate(context.Background(), []*schema.Message{
		{Role: schema.User, Content: "trigger error"},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Generate() error = %v, want %v", err, wantErr)
	}
}

func TestScriptedModelSupportsMalformedAndEmptyResponses(t *testing.T) {
	t.Parallel()

	model := NewScriptedModel(
		ScriptedResponse{
			Message: &schema.Message{Role: schema.Assistant, Content: "{"},
		},
		ScriptedResponse{
			Message: &schema.Message{Role: schema.Assistant, Content: ""},
		},
	)

	malformed, err := model.Generate(context.Background(), []*schema.Message{
		{Role: schema.User, Content: "malformed"},
	})
	if err != nil {
		t.Fatalf("malformed Generate() error = %v", err)
	}
	empty, err := model.Generate(context.Background(), []*schema.Message{
		{Role: schema.User, Content: "empty"},
	})
	if err != nil {
		t.Fatalf("empty Generate() error = %v", err)
	}

	if got, want := malformed.Content, "{"; got != want {
		t.Fatalf("malformed response = %q, want %q", got, want)
	}
	if got, want := empty.Content, ""; got != want {
		t.Fatalf("empty response = %q, want empty string", got)
	}
}

func TestScriptedModelStreamReturnsScriptedChunks(t *testing.T) {
	t.Parallel()

	model := NewScriptedModel(
		ScriptedResponse{
			Stream: []*schema.Message{
				{Role: schema.Assistant, Content: "part-1"},
				{Role: schema.Assistant, Content: "part-2"},
			},
		},
	)

	stream, err := model.Stream(context.Background(), []*schema.Message{
		{Role: schema.User, Content: "stream"},
	})
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	defer stream.Close()

	first, err := stream.Recv()
	if err != nil {
		t.Fatalf("first Recv() error = %v", err)
	}
	second, err := stream.Recv()
	if err != nil {
		t.Fatalf("second Recv() error = %v", err)
	}
	_, err = stream.Recv()
	if !errors.Is(err, io.EOF) {
		t.Fatalf("third Recv() error = %v, want %v", err, io.EOF)
	}

	if got, want := first.Content, "part-1"; got != want {
		t.Fatalf("first chunk = %q, want %q", got, want)
	}
	if got, want := second.Content, "part-2"; got != want {
		t.Fatalf("second chunk = %q, want %q", got, want)
	}
}
