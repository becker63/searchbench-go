// Package policy projects manifest evaluator tool allow/deny and systemPrompt into
// the effective runtime policy. Tool name validation is against a caller-supplied
// registry (the tools the evaluator factory can actually register), not the Pkl adapter.
package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	config "github.com/becker63/searchbench-go/internal/adapters/config/pkl"
)

// EvaluatorToolRegistry describes which evaluator tools exist at runtime versus which
// are allowed when the manifest omits an explicit allow list.
//
// Available lists every tool the evaluator factory can register; manifest allow entries
// must name only tools from this set.
//
// DefaultAllowed is the base allow set when agents.evaluator.tools.allow is empty;
// it must be non-empty and every entry must appear in Available. Deny subtracts
// from DefaultAllowed in that case, or from the explicit allow list when set.
type EvaluatorToolRegistry struct {
	Available      []string
	DefaultAllowed []string
}

// ResolveEvaluatorRunPolicy computes the effective allowed tool list, normalized
// denied names, trimmed system prompt text, and a stable SHA256 hex digest of the
// final resolved policy (effective allowed tools, denied, trimmed system prompt).
func ResolveEvaluatorRunPolicy(
	policy config.AgentToolPolicy,
	systemPrompt *string,
	reg EvaluatorToolRegistry,
) (
	effectiveAllowed []string,
	deniedNormalized []string,
	systemTrimmed string,
	policySHA256 string,
	err error,
) {
	avail, err := buildAvailableRegistry(reg.Available)
	if err != nil {
		return nil, nil, "", "", err
	}
	defaultSorted, err := normalizeDefaultAllowed(reg.DefaultAllowed, avail)
	if err != nil {
		return nil, nil, "", "", err
	}

	deniedNormalized = NormalizeDenied(policy.Deny)
	systemTrimmed = TrimSystemPrompt(systemPrompt)

	effectiveAllowed, err = EffectiveAllowed(policy, avail, defaultSorted)
	if err != nil {
		return nil, nil, "", "", err
	}

	policySHA256 = PromptPolicyHash(effectiveAllowed, deniedNormalized, systemTrimmed)
	return effectiveAllowed, deniedNormalized, systemTrimmed, policySHA256, nil
}

func buildAvailableRegistry(names []string) (map[string]struct{}, error) {
	seen := make(map[string]struct{})
	for _, raw := range names {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		seen[name] = struct{}{}
	}
	if len(seen) == 0 {
		return nil, fmt.Errorf("evaluator tool registry Available is empty")
	}
	return seen, nil
}

func normalizeDefaultAllowed(defaultNames []string, available map[string]struct{}) ([]string, error) {
	seen := make(map[string]struct{})
	var out []string
	for _, raw := range defaultNames {
		name := strings.TrimSpace(raw)
		if name == "" {
			return nil, fmt.Errorf("evaluator tool registry DefaultAllowed contains an empty tool name")
		}
		if _, ok := available[name]; !ok {
			return nil, fmt.Errorf("evaluator tool registry DefaultAllowed contains %q which is not in Available", name)
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("evaluator tool registry DefaultAllowed is empty")
	}
	sort.Strings(out)
	return out, nil
}

// NormalizeDenied returns manifest deny entries trimmed and sorted for stable hashing and output.
func NormalizeDenied(deny []string) []string {
	out := make([]string, 0, len(deny))
	for _, d := range deny {
		out = append(out, strings.TrimSpace(d))
	}
	sort.Strings(out)
	return out
}

// TrimSystemPrompt returns the manifest systemPrompt with surrounding whitespace removed.
func TrimSystemPrompt(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}

// EffectiveAllowed applies allow/deny rules. When the manifest allow list is empty,
// candidates are DefaultAllowed (caller must pass the normalized sorted slice).
// When allow is non-empty, each entry must appear in available; candidates are sorted.
// Deny always subtracts from the chosen candidate set.
func EffectiveAllowed(
	policy config.AgentToolPolicy,
	available map[string]struct{},
	defaultAllowedSorted []string,
) ([]string, error) {
	deny := make(map[string]struct{}, len(policy.Deny))
	for _, d := range policy.Deny {
		deny[strings.TrimSpace(d)] = struct{}{}
	}

	var candidates []string
	if len(policy.Allow) == 0 {
		candidates = append([]string(nil), defaultAllowedSorted...)
	} else {
		for _, a := range policy.Allow {
			name := strings.TrimSpace(a)
			if _, ok := available[name]; !ok {
				return nil, fmt.Errorf("agents.evaluator.tools.allow contains unknown evaluator tool %q", name)
			}
			candidates = append(candidates, name)
		}
		sort.Strings(candidates)
	}

	var out []string
	for _, name := range candidates {
		if _, banned := deny[name]; !banned {
			out = append(out, name)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("evaluator tool policy leaves no tools after applying deny list")
	}
	return out, nil
}

// PromptPolicyHash hashes the normalized resolved policy: sorted effective allowed tools,
// sorted normalized denied tools, and trimmed system prompt text.
func PromptPolicyHash(effectiveAllowed, deniedNormalized []string, systemPrompt string) string {
	type payload struct {
		AllowedTools []string `json:"allowed_tools"`
		DeniedTools  []string `json:"denied_tools"`
		SystemPrompt string   `json:"system_prompt,omitempty"`
	}
	a := append([]string(nil), effectiveAllowed...)
	sort.Strings(a)
	d := append([]string(nil), deniedNormalized...)
	sort.Strings(d)
	sp := strings.TrimSpace(systemPrompt)

	data, err := json.Marshal(payload{
		AllowedTools: a,
		DeniedTools:  d,
		SystemPrompt: sp,
	})
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
