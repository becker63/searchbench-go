# Aggregates SearchBench flake `searchbench-*` commands (no `scripts/` directory).
{ pkgs }:
let
  mkInRepo = import ./mk-in-repo.nix { inherit pkgs; };
  core = import ./core.nix { inherit pkgs mkInRepo; };
  generated = import ./generated-checks.nix { inherit pkgs mkInRepo; };
  push = import ./push-wrappers.nix {
    inherit pkgs;
    inherit (core)
      searchbench-go-test-all
      searchbench-e2e
      searchbench-go-mod-tidy-check
      searchbench-staticcheck
      searchbench-golangci
      searchbench-nix-flake-check
      ;
    inherit (generated) searchbench-check-generated;
  };
in
core // generated // push
