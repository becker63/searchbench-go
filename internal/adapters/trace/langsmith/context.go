package langsmith

import (
	"context"
	"os"
	"strings"
	"sync"

	extlangsmith "github.com/cloudwego/eino-ext/callbacks/langsmith"
)

// ContextLabels carries non-authoritative dimensions for LangSmith trace UX only.
type ContextLabels struct {
	SessionName string
	MatchID     string
	RunID       string
	Role        string
}

// AugmentContext attaches LangSmith trace options when LANGSMITH_API_KEY is set.
// It does not perform scoring or decisions and does not replace SearchBench usage
// accounting.
func AugmentContext(ctx context.Context, labels ContextLabels) context.Context {
	if strings.TrimSpace(os.Getenv(EnvAPIKey)) == "" {
		return ctx
	}

	session := strings.TrimSpace(labels.SessionName)
	if session == "" {
		session = strings.TrimSpace(os.Getenv(EnvSession))
	}
	if session == "" {
		session = "searchbench"
	}

	opts := []extlangsmith.TraceOption{
		extlangsmith.WithSessionName(session),
		extlangsmith.AddTag("searchbench"),
	}
	if id := strings.TrimSpace(labels.MatchID); id != "" {
		opts = append(opts, extlangsmith.AddTag("match_id:"+id))
	}
	if id := strings.TrimSpace(labels.RunID); id != "" {
		opts = append(opts, extlangsmith.AddTag("run_id:"+id))
	}
	if role := strings.TrimSpace(labels.Role); role != "" {
		opts = append(opts, extlangsmith.AddTag("role:"+role))
	}

	meta := &sync.Map{}
	if v := strings.TrimSpace(labels.MatchID); v != "" {
		meta.Store("searchbench.match_id", v)
	}
	if v := strings.TrimSpace(labels.RunID); v != "" {
		meta.Store("searchbench.run_id", v)
	}
	if v := strings.TrimSpace(labels.Role); v != "" {
		meta.Store("searchbench.role", v)
	}
	opts = append(opts, extlangsmith.SetMetadata(meta))

	return extlangsmith.SetTrace(ctx, opts...)
}
