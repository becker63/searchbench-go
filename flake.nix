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

        # Modules live under nix/vendor/; root vendor/ is a symlink for Go.
        vendorExcludes = [
          "^nix/vendor/"
          "^vendor/"
        ];

        commonHooks = {
          gofmt = {
            enable = true;
            excludes = vendorExcludes;
          };
          govet = {
            enable = true;
            extraPackages = [ pkgs.go ];
            excludes = vendorExcludes;
          };
          golangci-lint = {
            enable = true;
            extraPackages = [ pkgs.go ];
            excludes = vendorExcludes;
          };
          nixfmt-rfc-style = {
            enable = true;
            excludes = vendorExcludes;
          };
          deadnix = {
            enable = true;
            excludes = vendorExcludes;
          };
          statix = {
            enable = true;
            settings.ignore = [
              "nix/vendor/**"
              "vendor/**"
            ];
          };
          shellcheck = {
            enable = true;
            excludes = vendorExcludes;
          };
          shfmt = {
            enable = true;
            excludes = vendorExcludes;
          };
          trim-trailing-whitespace = {
            enable = true;
            excludes = [
              "^repomix-output\\.xml$"
              "^configs/rounds/.*/artifacts/"
              "^attached_assets/"
            ]
            ++ vendorExcludes;
          };
          end-of-file-fixer = {
            enable = true;
            excludes = [
              "^repomix-output\\.xml$"
              "^testdata/"
              "^configs/rounds/.*/artifacts/"
              "^attached_assets/"
            ]
            ++ vendorExcludes;
          };
          check-merge-conflicts.enable = true;
          check-added-large-files = {
            enable = true;
            excludes = vendorExcludes;
          };
          check-symlinks.enable = true;
          check-json.enable = true;
          check-yaml.enable = true;
          check-toml.enable = true;

          searchbench-architecture = {
            enable = true;
            name = "searchbench architecture import boundaries";
            entry = "${tools.searchbench-architecture-check}/bin/searchbench-architecture-check";
            pass_filenames = false;
          };

          searchbench-vocabulary = {
            enable = true;
            name = "searchbench vocabulary (warnings only)";
            entry = "${tools.searchbench-vocabulary-check}/bin/searchbench-vocabulary-check";
            pass_filenames = false;
          };
        };

        prePushHooks = {
          go-test-all = {
            enable = true;
            name = "go test ./... (pre-push)";
            entry = "${tools.searchbench-go-test-all-push}/bin/searchbench-go-test-all-push";
            pass_filenames = false;
            stages = [ "pre-push" ];
          };

          nix-flake-check-push = {
            enable = true;
            name = "nix flake check (pre-push)";
            entry = "${tools.searchbench-nix-flake-check-push}/bin/searchbench-nix-flake-check-push";
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
          hooks = commonHooks;
        };

        preCommitDev = git-hooks.lib.${system}.run {
          src = ./.;
          hooks = devHooks;
        };

        projectToolPkgs = with tools; [
          searchbench-go-test-all
          searchbench-e2e
          searchbench-architecture-check
          searchbench-check-pkl-generated
          searchbench-check-templ-generated
          searchbench-check-generated
          searchbench-update-repomix
          searchbench-nix-flake-check
          searchbench-vocabulary-check
          searchbench-agent-start
          searchbench-agent-check
          searchbench-agent-pack
          searchbench-agent-merge-check
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
          agent-check = {
            type = "app";
            program = "${tools.searchbench-agent-check}/bin/searchbench-agent-check";
          };
          agent-merge-check = {
            type = "app";
            program = "${tools.searchbench-agent-merge-check}/bin/searchbench-agent-merge-check";
          };
        };
      }
    );
}
