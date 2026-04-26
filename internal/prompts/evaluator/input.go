package evaluator

import (
	"sort"

	"github.com/becker63/searchbench-go/internal/domain"
)

// OutputSchemaJSON is the strict JSON contract the evaluator must satisfy.
const OutputSchemaJSON = `{"type":"object","additionalProperties":false,"properties":{"predicted_files":{"type":"array","items":{"type":"string"}},"reasoning":{"type":"string"}},"required":["predicted_files"]}`

// DefaultConstraints are the stable evaluator prompt constraints.
var DefaultConstraints = []string{
	"Use only the available tools when they help localize the bug.",
	"Do not include gold labels, oracle data, or scorer-only fields.",
	"Return strict JSON only. Do not wrap the final answer in markdown.",
}

// Input is the typed prompt contract for the minimal evaluator prompt.
type Input struct {
	TaskID           string
	RepoName         string
	RepoSHA          string
	IssueTitle       string
	IssueBody        string
	AllowedTools     []string
	Constraints      []string
	OutputSchemaJSON string
}

// InputFromTask projects the prompt-safe task data used by the evaluator
// prompt.
func InputFromTask(task domain.TaskSpec, allowedTools []string) Input {
	tools := append([]string(nil), allowedTools...)
	sort.Strings(tools)

	constraints := append([]string(nil), DefaultConstraints...)

	return Input{
		TaskID:           task.ID.String(),
		RepoName:         string(task.Repo.Name),
		RepoSHA:          string(task.Repo.SHA),
		IssueTitle:       task.Input.Title,
		IssueBody:        task.Input.Body,
		AllowedTools:     tools,
		Constraints:      constraints,
		OutputSchemaJSON: OutputSchemaJSON,
	}
}
