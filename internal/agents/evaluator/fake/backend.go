package fake

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// Decider is the deterministic fake compare.Decider.
type Decider struct{}

// NewDecider constructs a Decider.
func NewDecider() Decider { return Decider{} }

// Decide satisfies the structural compare.Decider interface.
func (Decider) Decide(comparisons []report.ScoreComparison, regressions []report.Regression) report.Decision {
	if len(regressions) > 0 {
		return report.Decision{
			Decision: report.DecisionReview,
			Reason:   "challenger has regressions in local fake comparison",
		}
	}
	for _, comparison := range comparisons {
		if comparison.Metric == score.MetricComposite && comparison.Challenger > comparison.Incumbent {
			return report.Decision{
				Decision: report.DecisionPromoteChallenger,
				Reason:   "challenger improves the composite score in local fake comparison",
			}
		}
	}
	return report.Decision{
		Decision: report.DecisionReview,
		Reason:   "challenger did not improve the composite score in local fake comparison",
	}
}

// ModelFactory returns a deterministic fake ToolCallingChatModel for a spec.
func ModelFactory(spec run.Spec) (model.ToolCallingChatModel, error) {
	return &evaluatorModel{spec: spec}, nil
}

// ToolFactory returns the deterministic fake repository tool set for a spec.
func ToolFactory(spec run.Spec) ([]tool.BaseTool, error) {
	return []tool.BaseTool{
		repoTool{
			spec: spec,
			name: "resolve",
			desc: "Return the most plausible fake repository paths for the issue.",
			run: func(spec run.Spec, _ map[string]any) any {
				return map[string]any{
					"paths": []string{"src/search_target.go", "src/incumbent_guess.go"},
					"issue": spec.Match.Input.Title,
				}
			},
		},
		repoTool{
			spec: spec,
			name: "expand",
			desc: "Return a deterministic fake file snippet for the requested path.",
			run: func(spec run.Spec, args map[string]any) any {
				path, _ := args["path"].(string)
				if strings.TrimSpace(path) == "" {
					path = "src/search_target.go"
				}
				return map[string]any{
					"path":    path,
					"snippet": snippetForPath(path),
				}
			},
		},
		repoTool{
			spec: spec,
			name: "resolve_and_expand",
			desc: "Resolve the likely file set and return one fake structural snippet.",
			run: func(spec run.Spec, _ map[string]any) any {
				return map[string]any{
					"paths": []string{"src/search_target.go", "src/incumbent_guess.go"},
					"files": []map[string]string{
						{
							"path":    "src/search_target.go",
							"snippet": snippetForPath("src/search_target.go"),
						},
						{
							"path":    "src/incumbent_guess.go",
							"snippet": snippetForPath("src/incumbent_guess.go"),
						},
					},
					"repo": string(spec.Match.Repo.Name),
				}
			},
		},
	}, nil
}

type evaluatorModel struct {
	spec  run.Spec
	tools []*schema.ToolInfo
}

func (m *evaluatorModel) Generate(ctx context.Context, input []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if hasToolResponse(input) {
		return schema.AssistantMessage(finalPredictionJSON(m.spec), nil), nil
	}

	toolName := "resolve_and_expand"
	if len(m.tools) > 0 {
		toolName = m.tools[0].Name
		for _, info := range m.tools {
			if info != nil && info.Name == "resolve_and_expand" {
				toolName = info.Name
				break
			}
		}
	}

	args := map[string]string{
		"query": m.spec.Match.Input.Title,
	}
	rawArgs, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	return schema.AssistantMessage("", []schema.ToolCall{{
		ID: "call-resolve-and-expand",
		Function: schema.FunctionCall{
			Name:      toolName,
			Arguments: string(rawArgs),
		},
	}}), nil
}

func (m *evaluatorModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	message, err := m.Generate(ctx, input, opts...)
	if err != nil {
		return nil, err
	}
	return schema.StreamReaderFromArray([]*schema.Message{message}), nil
}

func (m *evaluatorModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return &evaluatorModel{
		spec:  m.spec,
		tools: cloneToolInfos(tools),
	}, nil
}

type repoTool struct {
	spec run.Spec
	name string
	desc string
	run  func(run.Spec, map[string]any) any
}

func (t repoTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Desc: "Issue or symbol hint to search for.",
				Type: schema.String,
			},
			"path": {
				Desc: "Repository-relative path to expand.",
				Type: schema.String,
			},
		}),
	}, nil
}

func (t repoTool) InvokableRun(_ context.Context, input string, _ ...tool.Option) (string, error) {
	var args map[string]any
	if strings.TrimSpace(input) != "" {
		if err := json.Unmarshal([]byte(input), &args); err != nil {
			return "", fmt.Errorf("%s arguments: %w", t.name, err)
		}
	}
	raw, err := json.Marshal(t.run(t.spec, args))
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

var _ tool.InvokableTool = repoTool{}

func hasToolResponse(messages []*schema.Message) bool {
	for _, message := range messages {
		if message != nil && message.Role == schema.Tool {
			return true
		}
	}
	return false
}

func finalPredictionJSON(spec run.Spec) string {
	path := "src/incumbent_guess.go"
	reasoning := "incumbent stayed conservative after the fake structural lookup"
	if spec.System.Backend == domain.BackendIterativeContext {
		path = string(spec.Match.Oracle.GoldFiles[0])
		reasoning = "challenger used the fake structural lookup to narrow onto the search target"
	}

	payload := struct {
		PredictedFiles []string `json:"predicted_files"`
		Reasoning      string   `json:"reasoning"`
	}{
		PredictedFiles: []string{path},
		Reasoning:      reasoning,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Sprintf(`{"predicted_files":["%s"]}`, path)
	}
	return string(data)
}

func snippetForPath(repoPath string) string {
	switch strings.TrimSpace(repoPath) {
	case "src/search_target.go":
		return "func locateRetryTarget() { /* deterministic fake target */ }"
	case "src/incumbent_guess.go":
		return "func incumbentGuess() { /* deterministic fake fallback */ }"
	default:
		return "func fakeSnippet() {}"
	}
}

func cloneToolInfos(tools []*schema.ToolInfo) []*schema.ToolInfo {
	if len(tools) == 0 {
		return nil
	}

	cloned := make([]*schema.ToolInfo, len(tools))
	for i, info := range tools {
		if info == nil {
			continue
		}
		copyInfo := *info
		cloned[i] = &copyInfo
	}
	return cloned
}
