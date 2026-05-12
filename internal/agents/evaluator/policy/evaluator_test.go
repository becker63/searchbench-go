package policy

import (
	"strings"
	"testing"

	config "github.com/becker63/searchbench-go/internal/adapters/config/pkl"
	evaluatorfake "github.com/becker63/searchbench-go/internal/agents/evaluator/fake"
)

func testRegistry() EvaluatorToolRegistry {
	return EvaluatorToolRegistry{
		Available:      evaluatorfake.LocalEvaluatorToolNames(),
		DefaultAllowed: evaluatorfake.LocalEvaluatorDefaultAllowedToolNames(),
	}
}

func TestPromptPolicyHashIgnoresToolListOrder(t *testing.T) {
	t.Parallel()

	h1 := PromptPolicyHash(
		[]string{"resolve", "expand"},
		[]string{"net", "shell"},
		"hello",
	)
	h2 := PromptPolicyHash(
		[]string{"expand", "resolve"},
		[]string{"shell", "net"},
		"hello",
	)
	if h1 != h2 {
		t.Fatalf("hash differs for same sets in different order: %s vs %s", h1, h2)
	}
}

func TestPromptPolicyHashWhitespaceSystemPromptNormalized(t *testing.T) {
	t.Parallel()

	h1 := PromptPolicyHash([]string{"expand"}, []string{"shell"}, "  trimmed  ")
	h2 := PromptPolicyHash([]string{"expand"}, []string{"shell"}, "trimmed")
	if h1 != h2 {
		t.Fatal("hash differs for whitespace-only system prompt variation")
	}
}

func TestPromptPolicyHashContentChange(t *testing.T) {
	t.Parallel()

	a := PromptPolicyHash([]string{"expand"}, []string{}, "a")
	b := PromptPolicyHash([]string{"expand"}, []string{}, "b")
	if a == b {
		t.Fatal("hash should differ when system prompt content changes")
	}
}

func TestPromptPolicyHashEffectiveToolsChange(t *testing.T) {
	t.Parallel()

	a := PromptPolicyHash([]string{"expand"}, []string{}, "")
	b := PromptPolicyHash([]string{"resolve"}, []string{}, "")
	if a == b {
		t.Fatal("hash should differ when effective tools change")
	}
}

func TestPromptPolicyHashDeniedChange(t *testing.T) {
	t.Parallel()

	a := PromptPolicyHash([]string{"expand", "resolve"}, []string{"shell"}, "")
	b := PromptPolicyHash([]string{"expand", "resolve"}, []string{"network"}, "")
	if a == b {
		t.Fatal("hash should differ when denied tools change")
	}
}

func TestResolveDenyOrderProducesSameHash(t *testing.T) {
	t.Parallel()

	sp := "x"
	policy1 := config.AgentToolPolicy{
		Deny: []string{"network", "shell"},
	}
	policy2 := config.AgentToolPolicy{
		Deny: []string{"shell", "network"},
	}
	_, d1, _, h1, err := ResolveEvaluatorRunPolicy(policy1, &sp, testRegistry())
	if err != nil {
		t.Fatal(err)
	}
	_, d2, _, h2, err := ResolveEvaluatorRunPolicy(policy2, &sp, testRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Fatalf("hash differs for deny order: %s vs %s", h1, h2)
	}
	if strings.Join(d1, ",") != strings.Join(d2, ",") {
		t.Fatalf("normalized deny differs: %v vs %v", d1, d2)
	}
}

func TestResolveUnknownAllowRejected(t *testing.T) {
	t.Parallel()

	policy := config.AgentToolPolicy{
		Allow: []string{"resolve", "not_a_tool"},
	}
	_, _, _, _, err := ResolveEvaluatorRunPolicy(policy, nil, testRegistry())
	if err == nil || !strings.Contains(err.Error(), `unknown evaluator tool "not_a_tool"`) {
		t.Fatalf("ResolveEvaluatorRunPolicy() error = %v", err)
	}
}

func TestResolveNoToolsAfterDeny(t *testing.T) {
	t.Parallel()

	policy := config.AgentToolPolicy{
		Deny: []string{"resolve", "expand", "resolve_and_expand"},
	}
	_, _, _, _, err := ResolveEvaluatorRunPolicy(policy, nil, testRegistry())
	if err == nil || !strings.Contains(err.Error(), "no tools") {
		t.Fatalf("ResolveEvaluatorRunPolicy() error = %v", err)
	}
}

func TestResolveEmptyAllowUsesDefaultsMinusDeny(t *testing.T) {
	t.Parallel()

	policy := config.AgentToolPolicy{
		Deny: []string{"resolve"},
	}
	effective, _, _, _, err := ResolveEvaluatorRunPolicy(policy, nil, testRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if len(effective) != 2 {
		t.Fatalf("effective = %v, want 2 tools after denying resolve", effective)
	}
	for _, name := range effective {
		if name == "resolve" {
			t.Fatalf("denied tool resolve leaked into effective: %v", effective)
		}
	}
}

func TestResolveExplicitAllowOnlyListsAllowed(t *testing.T) {
	t.Parallel()

	policy := config.AgentToolPolicy{
		Allow: []string{"expand", "resolve"},
	}
	effective, _, _, _, err := ResolveEvaluatorRunPolicy(policy, nil, testRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if len(effective) != 2 {
		t.Fatalf("effective = %v", effective)
	}
	for _, name := range effective {
		if name == "resolve_and_expand" {
			t.Fatal("explicit allow should exclude resolve_and_expand")
		}
	}
}

func TestResolveRegistrySliceOrderDoesNotChangeEffectiveOrHash(t *testing.T) {
	t.Parallel()

	policy := config.AgentToolPolicy{Deny: []string{"shell"}}
	sp := "policy"

	orderA := []string{"resolve", "expand", "resolve_and_expand"}
	orderB := []string{"resolve_and_expand", "resolve", "expand"}
	defaultAll := []string{"expand", "resolve", "resolve_and_expand"}

	regA := EvaluatorToolRegistry{Available: orderA, DefaultAllowed: append([]string(nil), defaultAll...)}
	regB := EvaluatorToolRegistry{Available: orderB, DefaultAllowed: append([]string(nil), defaultAll...)}

	e1, d1, s1, h1, err := ResolveEvaluatorRunPolicy(policy, strPtr(sp), regA)
	if err != nil {
		t.Fatal(err)
	}
	e2, d2, s2, h2, err := ResolveEvaluatorRunPolicy(policy, strPtr(sp), regB)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Fatalf("hash differs: %s vs %s", h1, h2)
	}
	if strings.Join(e1, ",") != strings.Join(e2, ",") || s1 != s2 || strings.Join(d1, ",") != strings.Join(d2, ",") {
		t.Fatalf("resolved policy differs for registry order:\ne1=%v e2=%v", e1, e2)
	}
}

func TestResolveDifferentSystemPromptsYieldsDifferentHash(t *testing.T) {
	t.Parallel()

	policy := config.AgentToolPolicy{}
	a := "one"
	b := "two"
	_, _, _, h1, err := ResolveEvaluatorRunPolicy(policy, &a, testRegistry())
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, h2, err := ResolveEvaluatorRunPolicy(policy, &b, testRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if h1 == h2 {
		t.Fatal("expected different hashes for different system prompts")
	}
}

func TestResolveEmptyAllowUsesDefaultAllowedNotFullAvailable(t *testing.T) {
	t.Parallel()

	reg := EvaluatorToolRegistry{
		Available:      []string{"expand", "resolve", "resolve_and_expand"},
		DefaultAllowed: []string{"expand", "resolve"},
	}
	effective, _, _, _, err := ResolveEvaluatorRunPolicy(config.AgentToolPolicy{}, nil, reg)
	if err != nil {
		t.Fatal(err)
	}
	if len(effective) != 2 {
		t.Fatalf("effective = %v, want expand + resolve only", effective)
	}
	for _, name := range effective {
		if name == "resolve_and_expand" {
			t.Fatal("default allow should not include resolve_and_expand")
		}
	}

	explicit := config.AgentToolPolicy{Allow: []string{"resolve_and_expand"}}
	e2, _, _, _, err := ResolveEvaluatorRunPolicy(explicit, nil, reg)
	if err != nil {
		t.Fatal(err)
	}
	if len(e2) != 1 || e2[0] != "resolve_and_expand" {
		t.Fatalf("explicit allow = %v, want [resolve_and_expand]", e2)
	}
}

func TestResolveDefaultAllowedNotInAvailableFails(t *testing.T) {
	t.Parallel()

	reg := EvaluatorToolRegistry{
		Available:      []string{"expand", "resolve"},
		DefaultAllowed: []string{"resolve_and_expand"},
	}
	_, _, _, _, err := ResolveEvaluatorRunPolicy(config.AgentToolPolicy{}, nil, reg)
	if err == nil || !strings.Contains(err.Error(), "not in Available") {
		t.Fatalf("ResolveEvaluatorRunPolicy() error = %v", err)
	}
}

func strPtr(s string) *string { return &s }
