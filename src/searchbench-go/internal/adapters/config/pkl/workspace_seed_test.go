package config_test

import (
	"testing"

	config "github.com/becker63/searchbench-go/internal/adapters/config/pkl"
)

func TestValidateWorkspaceSeedConfigLocalPath(t *testing.T) {
	t.Parallel()
	local := "src/iterative-context"
	if err := config.ValidateWorkspaceSeedConfig("local_path", &local); err != nil {
		t.Fatal(err)
	}
	if err := config.ValidateWorkspaceSeedConfig("local_path", nil); err == nil {
		t.Fatal("expected local path required")
	}
	if err := config.ValidateWorkspaceSeedConfig("buck_descriptor", &local); err == nil {
		t.Fatal("expected unknown provider on localpath branch")
	}
}
