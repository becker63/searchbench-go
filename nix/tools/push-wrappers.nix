# Thin wrappers for pre-push git-hook stages.
{
  pkgs,
  searchbench-go-test-all,
  searchbench-e2e,
  searchbench-check-generated,
  searchbench-go-mod-tidy-check,
  searchbench-staticcheck,
  searchbench-golangci,
}:
{
  searchbench-go-test-all-push = pkgs.writeShellApplication {
    name = "searchbench-go-test-all-push";
    text = ''
      set -euo pipefail
      exec ${searchbench-go-test-all}/bin/searchbench-go-test-all
    '';
    runtimeInputs = [ searchbench-go-test-all ];
  };

  searchbench-e2e-push = pkgs.writeShellApplication {
    name = "searchbench-e2e-push";
    text = ''
      set -euo pipefail
      exec ${searchbench-e2e}/bin/searchbench-e2e
    '';
    runtimeInputs = [ searchbench-e2e ];
  };

  searchbench-check-generated-push = pkgs.writeShellApplication {
    name = "searchbench-check-generated-push";
    text = ''
      set -euo pipefail
      exec ${searchbench-check-generated}/bin/searchbench-check-generated
    '';
    runtimeInputs = [ searchbench-check-generated ];
  };

  searchbench-go-mod-tidy-check-push = pkgs.writeShellApplication {
    name = "searchbench-go-mod-tidy-check-push";
    text = ''
      set -euo pipefail
      exec ${searchbench-go-mod-tidy-check}/bin/searchbench-go-mod-tidy-check
    '';
    runtimeInputs = [ searchbench-go-mod-tidy-check ];
  };

  searchbench-staticcheck-push = pkgs.writeShellApplication {
    name = "searchbench-staticcheck-push";
    text = ''
      set -euo pipefail
      exec ${searchbench-staticcheck}/bin/searchbench-staticcheck
    '';
    runtimeInputs = [ searchbench-staticcheck ];
  };

  searchbench-golangci-push = pkgs.writeShellApplication {
    name = "searchbench-golangci-push";
    text = ''
      set -euo pipefail
      exec ${searchbench-golangci}/bin/searchbench-golangci
    '';
    runtimeInputs = [ searchbench-golangci ];
  };
}
