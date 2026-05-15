package bundlefs

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	pureround "github.com/becker63/searchbench-go/internal/pure/round"
)

const continuationPKLFileName = "continuation.pkl"

func renderContinuationPKL(bundleDir string, continuation pureround.Continuation, input ContinuationPKLInput) ([]byte, error) {
	if err := continuation.Validate(); err != nil {
		return nil, fmt.Errorf("render continuation.pkl: %w", err)
	}
	if strings.TrimSpace(input.SchemaPath) == "" {
		return nil, fmt.Errorf("render continuation.pkl: schema path is required")
	}
	if strings.TrimSpace(input.HelpersPath) == "" {
		return nil, fmt.Errorf("render continuation.pkl: helpers path is required")
	}

	schemaPath, err := relativeBundlePath(bundleDir, input.SchemaPath)
	if err != nil {
		return nil, fmt.Errorf("render continuation.pkl schema path: %w", err)
	}
	helpersPath, err := relativeBundlePath(bundleDir, input.HelpersPath)
	if err != nil {
		return nil, fmt.Errorf("render continuation.pkl helpers path: %w", err)
	}
	objectivePath := filepath.ToSlash(filepath.Clean(continuation.Scoring.ObjectivePath))

	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = continuation.Round.ID + "-continuation"
	}

	incumbentLines, err := renderContinuationIncumbent(bundleDir, continuation)
	if err != nil {
		return nil, err
	}
	challengerLines, err := renderContinuationChallenger(continuation)
	if err != nil {
		return nil, err
	}
	matchesLine, err := renderContinuationMatches(continuation)
	if err != nil {
		return nil, err
	}
	evaluatorLine, err := renderContinuationEvaluator(continuation)
	if err != nil {
		return nil, err
	}

	lines := []string{
		`amends ` + pklQuoted(schemaPath),
		`import ` + pklQuoted(helpersPath) + ` as game`,
		"",
		`name = ` + pklQuoted(name),
		"",
		`round = (game.continueFrom(".")) {`,
		`  id = ` + pklQuoted(continuation.Round.ID),
	}
	lines = append(lines, incumbentLines...)
	lines = append(lines, challengerLines...)
	lines = append(lines,
		`  matches = `+matchesLine,
		`  scoring = game.objective(`+pklQuoted(objectivePath)+`)`,
		`  evaluator = `+evaluatorLine,
		`}`,
	)
	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

func renderContinuationIncumbent(bundleDir string, continuation pureround.Continuation) ([]string, error) {
	system := continuation.SurvivingCandidate.System
	switch system.Backend {
	case domain.BackendJCodeMunch:
		if system.Policy != nil {
			return nil, fmt.Errorf("render continuation.pkl incumbent: jcodemunch with policy is unsupported")
		}
		return []string{`  incumbent = game.jcodemunch()`}, nil
	case domain.BackendIterativeContext:
		if system.Policy == nil {
			return nil, fmt.Errorf("render continuation.pkl incumbent: iterative-context policy is required")
		}
		policyPath := filepath.ToSlash(continuation.ResolveArtifactPath(domain.HostPath(bundleDir)))
		return []string{
			`  incumbent = (game.iterativeContextPolicy(` + pklQuoted(policyPath) + `)) {`,
			`    selectionPolicy {`,
			`      id = ` + pklQuoted(system.Policy.ID.String()),
			`    }`,
			`  }`,
		}, nil
	case domain.BackendFake:
		if system.Policy == nil {
			return []string{
				`  incumbent = game.fakeRoundPolicy(` + pklQuoted(system.ID.String()) + `, ` + pklQuoted(system.Name) + `)`,
			}, nil
		}
		policyPath := filepath.ToSlash(continuation.ResolveArtifactPath(domain.HostPath(bundleDir)))
		return []string{
			`  incumbent = (game.iterativeContextPolicy(` + pklQuoted(policyPath) + `)) {`,
			`    system {`,
			`      id = ` + pklQuoted(system.ID.String()),
			`      name = ` + pklQuoted(system.Name),
			`      backend = ` + pklQuoted("fake"),
			`    }`,
			`    selectionPolicy {`,
			`      id = ` + pklQuoted(system.Policy.ID.String()),
			`    }`,
			`  }`,
		}, nil
	default:
		return nil, fmt.Errorf("render continuation.pkl incumbent: backend %q is unsupported", system.Backend)
	}
}

func renderContinuationChallenger(continuation pureround.Continuation) ([]string, error) {
	system := continuation.SurvivingCandidate.System
	switch system.Backend {
	case domain.BackendJCodeMunch:
		if system.Policy != nil {
			return nil, fmt.Errorf("render continuation.pkl challenger: jcodemunch with policy is unsupported")
		}
		return []string{
			`  challenger {`,
			`    system {`,
			`      id = ` + pklQuoted(system.ID.String()),
			`      name = ` + pklQuoted(system.Name),
			`      backend = ` + pklQuoted("jcodemunch"),
			`    }`,
			`  }`,
		}, nil
	case domain.BackendIterativeContext:
		return []string{
			`  challenger {`,
			`    system {`,
			`      id = ` + pklQuoted(system.ID.String()),
			`      name = ` + pklQuoted(system.Name),
			`      backend = ` + pklQuoted("iterative_context"),
			`    }`,
			`  }`,
		}, nil
	case domain.BackendFake:
		return []string{
			`  challenger {`,
			`    system {`,
			`      id = ` + pklQuoted(system.ID.String()),
			`      name = ` + pklQuoted(system.Name),
			`      backend = ` + pklQuoted("fake"),
			`    }`,
			`  }`,
		}, nil
	default:
		return nil, fmt.Errorf("render continuation.pkl challenger: backend %q is unsupported", system.Backend)
	}
}

func renderContinuationMatches(continuation pureround.Continuation) (string, error) {
	dataset := continuation.Dataset
	if dataset.Kind == "" || dataset.Name == "" || dataset.Config == "" || dataset.Split == "" || dataset.MaxItems == nil {
		return "", fmt.Errorf("render continuation.pkl matches: dataset selection is incomplete")
	}
	if dataset.Kind != "lca" || dataset.Name != "JetBrains-Research/lca-bug-localization" {
		return "", fmt.Errorf("render continuation.pkl matches: dataset %q/%q is unsupported", dataset.Kind, dataset.Name)
	}
	return fmt.Sprintf(`game.lca(%s, %s, %d)`,
		pklQuoted(dataset.Config),
		pklQuoted(dataset.Split),
		*dataset.MaxItems,
	), nil
}

func renderContinuationEvaluator(continuation pureround.Continuation) (string, error) {
	evaluator := continuation.Evaluator
	if evaluator.Model.Provider != "fake" ||
		evaluator.Model.Name != "fake-evaluator" ||
		evaluator.Model.MaxOutputTokens != 2000 ||
		evaluator.Bounds.MaxModelTurns != 8 ||
		evaluator.Bounds.MaxToolCalls != 24 ||
		evaluator.Bounds.TimeoutSeconds != 300 ||
		evaluator.Retry.MaxAttempts != 2 ||
		!evaluator.Retry.RetryOnModelError ||
		evaluator.Retry.RetryOnToolFailure ||
		!evaluator.Retry.RetryOnFinalizationFailure ||
		!evaluator.Retry.RetryOnInvalidPrediction ||
		evaluator.SystemPrompt != "Use structural code evidence before guessing.\nPrefer the smallest plausible file set.\nDo not invent files that were not returned by tools." ||
		!sameStrings(evaluator.AllowedTools, []string{"resolve", "expand", "resolve_and_expand"}) ||
		!sameStrings(evaluator.DeniedTools, []string{"shell", "go_test", "write_file", "network"}) {
		return "", fmt.Errorf("render continuation.pkl evaluator: only the fake local evaluator shape is currently supported")
	}
	return "game.fakeEvaluator()", nil
}

func sameStrings(got []string, want []string) bool {
	gotCopy := append([]string(nil), got...)
	wantCopy := append([]string(nil), want...)
	slices.Sort(gotCopy)
	slices.Sort(wantCopy)
	return slices.Equal(gotCopy, wantCopy)
}

func relativeBundlePath(bundleDir string, target string) (string, error) {
	if strings.TrimSpace(target) == "" {
		return "", fmt.Errorf("target path is required")
	}
	targetPath := target
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Clean(filepath.Join(bundleDir, targetPath))
	}
	rel, err := filepath.Rel(bundleDir, targetPath)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

func pklQuoted(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\t", `\t`)
	return `"` + replacer.Replace(value) + `"`
}
