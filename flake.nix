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

        buckCheck = pkgs.writeShellApplication {
          name = "buck-check";
          runtimeInputs = [
            pkgs.buck2
          ];
          text = ''
            set -euo pipefail
            cd "$(git rev-parse --show-toplevel)"
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
              "^configs/rounds/.*/artifacts/"
              "^attached_assets/"
            ];
          };
          end-of-file-fixer = {
            enable = true;
            excludes = [
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
          buck2-check = {
            enable = true;
            name = "buck2 test //:check";
            entry = "${buckCheck}/bin/buck-check";
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

        lcaExportPython = pkgs.python3.withPackages (
          ps: with ps; [
            datasets
            huggingface-hub
          ]
        );

      in
      {
        formatter = pkgs.nixfmt;

        checks.pre-commit-check = preCommitCheck;

        devShells.default = pkgs.mkShell {
          packages =
            (with pkgs; [
              clang
              lld
              stdenv.cc.cc
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
              nodejs_22
              lcaExportPython
            ])
            ++ preCommitDev.enabledPackages;

          shellHook = ''
            export LD_LIBRARY_PATH="${pkgs.stdenv.cc.cc.lib}/lib''${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
            echo "${pkgs.stdenv.cc.cc.lib}/lib" > tools/libstdcxx_libdir
            # Prelude go_test binaries use RUNPATH $PROJECT_ROOT/outputs/out/lib for libstdc++.
            mkdir -p outputs/out/lib
            for lib in libstdc++.so.6 libgcc_s.so.1; do
              ln -sfn "${pkgs.stdenv.cc.cc.lib}/lib/$lib" "outputs/out/lib/$lib"
            done
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
            echo ""
            echo "Go dependency projection (explicit; Buck does not run this):"
            echo "  cd src/searchbench-go && go mod vendor"
            echo "  buck2 run prelude//go/tools/gobuckify:gobuckify -- ."
            if [ ! -f src/searchbench-go/gobuckify.json ]; then
              echo "  [warn] missing src/searchbench-go/gobuckify.json"
            elif [ ! -d src/searchbench-go/vendor ]; then
              echo "  [warn] missing src/searchbench-go/vendor — run go mod vendor"
            fi
            echo ""
            echo "Python/IC: uv lock && uv sync in src/iterative-context; Buck uses wrapper targets only (no Elk)."
          '';
        };
      }
    );
}
