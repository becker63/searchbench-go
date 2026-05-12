# Aggregates SearchBench flake `searchbench-*` commands (no `scripts/` directory).
{ pkgs }:
let
  mkInRepo = import ./mk-in-repo.nix { inherit pkgs; };
  core = import ./core.nix { inherit pkgs mkInRepo; };
  generated = import ./generated-checks.nix { inherit pkgs mkInRepo; };
  netwatch = import ./netwatch.nix { inherit pkgs; };
  agent = import ./agent.nix {
    inherit pkgs mkInRepo;
    inherit (core)
      searchbench-e2e
      searchbench-update-repomix
      searchbench-staticcheck
      searchbench-golangci
      searchbench-nix-flake-check
      searchbench-go-mod-tidy-check
      searchbench-go-test-race
      ;
    inherit (generated) searchbench-check-generated;
  };
  push = import ./push-wrappers.nix {
    inherit pkgs;
    inherit (core)
      searchbench-go-test-all
      searchbench-e2e
      searchbench-go-mod-tidy-check
      searchbench-staticcheck
      searchbench-golangci
      ;
    inherit (generated) searchbench-check-generated;
  };
in
core // generated // agent // push // netwatch
