# Thin wrappers for pre-push git-hook stages.
{
  pkgs,
  searchbench-go-test-all,
  searchbench-e2e,
  searchbench-nix-flake-check,
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

  searchbench-nix-flake-check-push = pkgs.writeShellApplication {
    name = "searchbench-nix-flake-check-push";
    text = ''
      set -euo pipefail
      exec ${searchbench-nix-flake-check}/bin/searchbench-nix-flake-check
    '';
    runtimeInputs = [ searchbench-nix-flake-check ];
  };

  searchbench-e2e-push = pkgs.writeShellApplication {
    name = "searchbench-e2e-push";
    text = ''
      set -euo pipefail
      exec ${searchbench-e2e}/bin/searchbench-e2e
    '';
    runtimeInputs = [ searchbench-e2e ];
  };
}
