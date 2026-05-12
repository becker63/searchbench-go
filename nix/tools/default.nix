# Aggregates SearchBench flake `searchbench-*` commands (no `scripts/` directory).
{ pkgs }:
let
  mkInRepo = import ./mk-in-repo.nix { inherit pkgs; };
  core = import ./core.nix { inherit pkgs mkInRepo; };
  generated = import ./generated-checks.nix { inherit pkgs mkInRepo; };
  agent = import ./agent.nix {
    inherit pkgs mkInRepo;
    inherit (core) searchbench-e2e searchbench-update-repomix;
    inherit (generated) searchbench-check-generated;
  };
  push = import ./push-wrappers.nix {
    inherit pkgs;
    inherit (core) searchbench-go-test-all searchbench-e2e searchbench-nix-flake-check;
  };
in
core // generated // agent // push
