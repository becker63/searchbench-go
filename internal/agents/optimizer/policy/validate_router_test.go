package policy

import (
	"context"
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

func TestValidateICProposalRejectedWhenMarkdownFenceInCode(t *testing.T) {
	t.Parallel()

	_, fail := Validate(context.Background(), pureoptimizer.NextChallengerProposal{
		ArtifactID:   domain.ArtifactID("next-challenger-test"),
		ArtifactName: "proposal.py",
		InterfaceID:  IterativeContextSelectionPolicyInterfaceID,
		Code:         "```python\npass\n```",
	})
	if fail == nil {
		t.Fatal("expected validation failure")
	}
	if fail.PipelineFeedback == "" {
		t.Fatal("expected pipeline feedback for optimizer retries")
	}
}
