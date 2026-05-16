package config_test

import (
	"testing"

	config "github.com/becker63/searchbench-go/internal/adapters/config/pkl"
)

func TestValidateWorkspaceSeedConfigBuckDescriptor(t *testing.T) {
	t.Parallel()
	target := "//src/iterative-context:optimizable_backend"
	if err := config.ValidateWorkspaceSeedConfig("buck_descriptor", &target); err != nil {
		t.Fatal(err)
	}
	if err := config.ValidateWorkspaceSeedConfig("buck_descriptor", nil); err == nil {
		t.Fatal("expected buck target required")
	}
	if err := config.ValidateWorkspaceSeedConfig("local_path", &target); err == nil {
		t.Fatal("expected unknown provider on buck branch")
	}
}
