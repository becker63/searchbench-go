{
  description = "SearchBench-Go — Nix dev shell, pre-commit, Buck2 toolchains";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    git-hooks.url = "github:cachix/git-hooks.nix";
    git-hooks.inputs.nixpkgs.follows = "nixpkgs";
    buck2-nix = {
      url = "github:tweag/buck2.nix";
      flake = false;
    };
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      git-hooks,
      buck2-nix,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        # Matches AGENTS.md: on commit, refresh Repomix snapshot then run the fast Buck gate.
        repomixThenBuckCheck = pkgs.writeShellApplication {
          name = "repomix-then-buck-check";
          runtimeInputs = [
            pkgs.git
            pkgs.repomix
            pkgs.buck2
          ];
          text = ''
            set -euo pipefail
            _root="$(git rev-parse --show-toplevel)"
            cd "$_root"
            repomix \
              --output repomix-output.xml \
              --style xml \
              --compress \
              --no-git-sort-by-changes
            git add repomix-output.xml
            exec buck2 test //:check
          '';
        };

        buckCheckFull = pkgs.writeShellApplication {
          name = "buck-check-full";
          runtimeInputs = [
            pkgs.git
            pkgs.buck2
          ];
          text = ''
            set -euo pipefail
            cd "$(git rev-parse --show-toplevel)"
            exec buck2 test //:check_full
          '';
        };

        # Runs in `nix flake check` (sandbox). Keep this lightweight — no `buck2 test` here.
        flakeCheckHooks = {
          gofmt.enable = true;
          nixfmt-rfc-style.enable = true;
          deadnix.enable = true;
          statix = {
            enable = true;
            entry = "statix check -i 'designing-for-two/**'";
          };
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
        };

        devHooks = flakeCheckHooks // {
          repomix-then-buck-check = {
            enable = true;
            name = "Repomix + buck2 test //:check";
            entry = "${repomixThenBuckCheck}/bin/repomix-then-buck-check";
            pass_filenames = false;
          };

          buck2-check-full = {
            enable = true;
            name = "buck2 test //:check_full";
            entry = "${buckCheckFull}/bin/buck-check-full";
            pass_filenames = false;
            stages = [ "pre-push" ];
          };
        };

        hookExcludes = [
          "designing-for-two"
        ];

        preCommitCheck = git-hooks.lib.${system}.run {
          src = ./.;
          excludes = hookExcludes;
          hooks = flakeCheckHooks;
        };

        preCommitDev = git-hooks.lib.${system}.run {
          src = ./.;
          excludes = hookExcludes;
          hooks = devHooks;
        };
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
              buck2
              nixfmt
              templ
              repomix
              nodejs_22
            ])
            ++ preCommitDev.enabledPackages;

          shellHook = ''
            mkdir -p .buckconfig.d
            cat >.buckconfig.d/buck2-nix.config <<EOS
            [external_cell_nix]
              git_origin = https://github.com/tweag/buck2.nix.git
              commit_hash = ${buck2-nix.rev}
            EOS
          ''
          + preCommitDev.shellHook
          + ''
            echo "searchbench-go: dev shell (pre-commit → Buck2 + hygiene; see AGENTS.md)"
            go version
          '';
        };
      }
    );
}
