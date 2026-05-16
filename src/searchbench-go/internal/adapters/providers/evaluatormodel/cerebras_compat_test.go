package evaluatormodel

import (
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestStripReasoningMessages(t *testing.T) {
	t.Parallel()
	in := []*schema.Message{
		{Role: schema.Assistant, Content: "ok", ReasoningContent: "think"},
	}
	out := stripReasoningMessages(in)
	if out[0].ReasoningContent != "" {
		t.Fatalf("ReasoningContent = %q, want empty", out[0].ReasoningContent)
	}
	if in[0].ReasoningContent != "think" {
		t.Fatal("expected input slice to remain unchanged")
	}
}
