package iterativecontext

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// ValidateAcceptedLaunch checks launch cwd matches the accepted workspace root.
func ValidateAcceptedLaunch(accepted optimizer.AcceptedICCandidate) error {
	if err := accepted.Validate(); err != nil {
		return err
	}
	wsRoot, err := filepath.Abs(strings.TrimSpace(accepted.Workspace.Root))
	if err != nil {
		return err
	}
	launchCWD, err := filepath.Abs(strings.TrimSpace(accepted.Launch.CWD))
	if err != nil {
		return err
	}
	if wsRoot != launchCWD {
		return fmt.Errorf("launch cwd %q does not match accepted workspace root %q", launchCWD, wsRoot)
	}
	return nil
}

// BuildLaunchCommand constructs the MCP subprocess command for an accepted candidate.
func BuildLaunchCommand(accepted optimizer.AcceptedICCandidate) (*exec.Cmd, error) {
	if err := ValidateAcceptedLaunch(accepted); err != nil {
		return nil, &Error{Kind: KindSession, Op: "validate accepted launch", Err: err}
	}
	argv := append([]string(nil), accepted.Launch.Argv...)
	if len(argv) == 0 {
		return nil, &Error{Kind: KindSession, Op: "launch argv", Err: fmt.Errorf("empty argv")}
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = accepted.Launch.CWD
	cmd.Env = os.Environ()
	for k, v := range accepted.Launch.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	return cmd, nil
}

// RuntimeIdentityFromAccepted records seed and policy identity for runtime evidence.
func RuntimeIdentityFromAccepted(accepted optimizer.AcceptedICCandidate, verified bool) optimizer.ICRuntimeIdentity {
	return optimizer.ICRuntimeIdentity{
		WorkspaceRoot: accepted.Workspace.Root,
		SeedIdentity:  accepted.Workspace.Seed,
		PolicyID:      accepted.Policy.PolicyID,
		PolicySha256:  accepted.Policy.Sha256,
		ActiveScoreID: accepted.Policy.PolicyID,
		Verified:      verified,
	}
}

// InstallRequestFromAccepted maps accepted policy metadata to harness install input.
func InstallRequestFromAccepted(accepted optimizer.AcceptedICCandidate) optimizer.ICPolicyInstallRequest {
	return optimizer.ICPolicyInstallRequest{
		WorkspaceRoot: accepted.Workspace.Root,
		PolicyPath:    accepted.Policy.Path,
		PolicyID:      accepted.Policy.PolicyID,
		Symbol:        accepted.Policy.Symbol,
		Sha256:        accepted.Policy.Sha256,
		Interface:     accepted.Policy.Interface,
	}
}

// OpenAcceptedCandidate is the seam for launching MCP from AcceptedICCandidate.
// Live MCP wiring is deferred when no command is available; callers use BuildLaunchCommand + OpenCommand.
func OpenAcceptedCandidate(ctx context.Context, accepted optimizer.AcceptedICCandidate) (*Runtime, optimizer.ICRuntimeIdentity, error) {
	if err := ValidateAcceptedLaunch(accepted); err != nil {
		return nil, optimizer.ICRuntimeIdentity{}, err
	}
	cmd, err := BuildLaunchCommand(accepted)
	if err != nil {
		return nil, optimizer.ICRuntimeIdentity{}, err
	}
	install := InstallRequestFromAccepted(accepted)
	rt, err := OpenCommand(ctx, CommandConfig{
		Command:      cmd,
		RepoPath:     accepted.Workspace.Root,
		ScoreInstall: scoreInstallFromRequest(install),
	})
	if err != nil {
		return nil, optimizer.ICRuntimeIdentity{}, err
	}
	identity := RuntimeIdentityFromAccepted(accepted, true)
	return rt, identity, nil
}

func scoreInstallFromRequest(req optimizer.ICPolicyInstallRequest) *ScoreInstallParams {
	return &ScoreInstallParams{
		PolicyPath: req.PolicyPath,
		PolicyID:   req.PolicyID,
		Symbol:     req.Symbol,
	}
}
