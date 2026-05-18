package config

import (
	"testing"
)

func TestValidateWorkspaceSeedConfig(t *testing.T) {
	t.Parallel()
	local := "src/iterative-context"
	buck := "//src/iterative-context:optimizable_backend"
	if err := ValidateWorkspaceSeedConfig("local_path", &local, nil); err != nil {
		t.Fatal(err)
	}
	if err := ValidateWorkspaceSeedConfig("buck_descriptor", nil, &buck); err != nil {
		t.Fatal(err)
	}
	if err := ValidateWorkspaceSeedConfig("local_path", nil, nil); err == nil {
		t.Fatal("expected local path required")
	}
	if err := ValidateWorkspaceSeedConfig("buck_descriptor", nil, nil); err == nil {
		t.Fatal("expected buck target required")
	}
	if err := ValidateWorkspaceSeedConfig("git", &local, nil); err == nil {
		t.Fatal("expected git unsupported")
	}
}
