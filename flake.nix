{
  description = "SearchBench-Go — Nix dev shell, pre-commit, and CI checks";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    git-hooks.url = "github:cachix/git-hooks.nix";
    git-hooks.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      git-hooks,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        tools = import ./nix/tools { inherit pkgs; };

        # Runs in `nix flake check` (sandbox, no network). Go linters, `go test`, etc. need the
        # module proxy or a local module cache, so those hooks live only in `preCommitDev` / pre-push.
        flakeCheckHooks = {
          gofmt.enable = true;
          nixfmt-rfc-style.enable = true;
          deadnix.enable = true;
          statix.enable = true;
          shellcheck.enable = true;
          shfmt.enable = true;
          trim-trailing-whitespace = {
            enable = true;
            excludes = [
              "^repomix-output\\.xml$"
              "^configs/rounds/.*/artifacts/"
              "^attached_assets/"
            ];
          };
          end-of-file-fixer = {
            enable = true;
            excludes = [
              "^repomix-output\\.xml$"
              "^testdata/"
              "^configs/rounds/.*/artifacts/"
              "^attached_assets/"
            ];
          };
          check-merge-conflicts.enable = true;
          check-added-large-files.enable = true;
          check-symlinks.enable = true;
          check-json.enable = true;
          check-yaml.enable = true;
          check-toml.enable = true;

          searchbench-vocabulary = {
            enable = true;
            name = "searchbench vocabulary (warnings only)";
            entry = "${tools.searchbench-vocabulary-check}/bin/searchbench-vocabulary-check";
            pass_filenames = false;
          };

          searchbench-no-scripts = {
            enable = true;
            name = "searchbench no legacy scripts/ directory";
            entry = "${tools.searchbench-no-scripts-check}/bin/searchbench-no-scripts-check";
            pass_filenames = false;
          };
        };

        goModuleGraphHooks =
          # Standalone git-hooks `staticcheck` is intentionally omitted: `.golangci.yml`
          # enables staticcheck inside golangci-lint on pre-commit. Pre-push runs explicit
          # `searchbench-staticcheck` for a full-module proof (see AGENTS.md).
          {
            govet = {
              enable = true;
              extraPackages = [ pkgs.go ];
            };
            golangci-lint = {
              enable = true;
              extraPackages = [ pkgs.go ];
            };
            searchbench-architecture = {
              enable = true;
              name = "searchbench architecture import boundaries";
              entry = "${tools.searchbench-architecture-check}/bin/searchbench-architecture-check";
              pass_filenames = false;
            };
            searchbench-prompt-contract = {
              enable = true;
              name = "searchbench prompt contract tests";
              entry = "${tools.searchbench-prompt-contract-check}/bin/searchbench-prompt-contract-check";
              pass_filenames = false;
            };
          };

        commonHooks = flakeCheckHooks // goModuleGraphHooks;

        prePushHooks = {
          go-test-all = {
            enable = true;
            name = "go test ./... (pre-push)";
            entry = "${tools.searchbench-go-test-all-push}/bin/searchbench-go-test-all-push";
            pass_filenames = false;
            stages = [ "pre-push" ];
          };

          searchbench-e2e-push = {
            enable = true;
            name = "searchbench root e2e (pre-push)";
            entry = "${tools.searchbench-e2e-push}/bin/searchbench-e2e-push";
            pass_filenames = false;
            stages = [ "pre-push" ];
          };

          searchbench-check-generated-push = {
            enable = true;
            name = "searchbench check generated (pre-push)";
            entry = "${tools.searchbench-check-generated-push}/bin/searchbench-check-generated-push";
            pass_filenames = false;
            stages = [ "pre-push" ];
          };

          searchbench-go-mod-tidy-check-push = {
            enable = true;
            name = "searchbench go mod tidy check (pre-push)";
            entry = "${tools.searchbench-go-mod-tidy-check-push}/bin/searchbench-go-mod-tidy-check-push";
            pass_filenames = false;
            stages = [ "pre-push" ];
          };

          searchbench-staticcheck-push = {
            enable = true;
            name = "searchbench staticcheck (pre-push)";
            entry = "${tools.searchbench-staticcheck-push}/bin/searchbench-staticcheck-push";
            pass_filenames = false;
            stages = [ "pre-push" ];
          };

          searchbench-golangci-push = {
            enable = true;
            name = "searchbench golangci-lint (pre-push)";
            entry = "${tools.searchbench-golangci-push}/bin/searchbench-golangci-push";
            pass_filenames = false;
            stages = [ "pre-push" ];
          };

          searchbench-nix-flake-check-push = {
            enable = true;
            name = "nix flake check (pre-push)";
            entry = "${tools.searchbench-nix-flake-check-push}/bin/searchbench-nix-flake-check-push";
            pass_filenames = false;
            stages = [ "pre-push" ];
          };

          searchbench-repomix-fresh-check-push = {
            enable = true;
            name = "Repomix snapshot fresh (pre-push)";
            entry = "${tools.searchbench-repomix-fresh-check}/bin/searchbench-repomix-fresh-check";
            pass_filenames = false;
            stages = [ "pre-push" ];
          };
        };

        devHooks =
          commonHooks
          // prePushHooks
          // {
            repomix-snapshot = {
              enable = true;
              name = "Repomix snapshot";
              entry = "${tools.searchbench-update-repomix}/bin/searchbench-update-repomix";
              pass_filenames = false;
            };
          };

        preCommitCheck = git-hooks.lib.${system}.run {
          src = ./.;
          hooks = flakeCheckHooks;
        };

        preCommitDev = git-hooks.lib.${system}.run {
          src = ./.;
          hooks = devHooks;
        };

        projectToolPkgs = with tools; [
          searchbench-go-test-all
          searchbench-go-test-race
          searchbench-e2e
          searchbench-architecture-check
          searchbench-check-pkl-generated
          searchbench-check-templ-generated
          searchbench-check-generated
          searchbench-update-repomix
          searchbench-repomix-fresh-check
          searchbench-nix-flake-check
          searchbench-vocabulary-check
          searchbench-staticcheck
          searchbench-golangci
          searchbench-go-mod-tidy-check
          searchbench-prompt-contract-check
          searchbench-refresh-pkl-example-fixtures
          searchbench-go-build-root
          searchbench-publish-issue-wave
          searchbench-no-scripts-check
        ];
      in
      {
        formatter = pkgs.nixfmt;

        checks.pre-commit-check = preCommitCheck;

        devShells.default = pkgs.mkShell {
          packages =
            (with pkgs; [
              go
              gopls
              gotools
              go-tools
              pkl
              pkl-lsp
              jdk25
              pre-commit
              golangci-lint
              nixfmt
              templ
            ])
            ++ projectToolPkgs
            ++ preCommitDev.enabledPackages;

          shellHook = preCommitDev.shellHook + ''
            echo "searchbench-go: dev shell (pre-commit installed via git-hooks.nix)"
            go version
          '';
        };

        apps = {
          update-repomix = {
            type = "app";
            program = "${tools.searchbench-update-repomix}/bin/searchbench-update-repomix";
          };
          e2e = {
            type = "app";
            program = "${tools.searchbench-e2e}/bin/searchbench-e2e";
          };
          publish-issue-wave = {
            type = "app";
            program = "${tools.searchbench-publish-issue-wave}/bin/searchbench-publish-issue-wave";
          };
        };
      }
    );
}
