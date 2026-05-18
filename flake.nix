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

        projectGoDeps = pkgs.writeShellApplication {
          name = "project-go-deps";
          runtimeInputs = with pkgs; [
            go
            buck2
            git
            coreutils
            jq
          ];
          text = ''
            set -euo pipefail
            root="$(git rev-parse --show-toplevel)"
            cd "$root"

            go_mod="src/searchbench-go/go.mod"
            go_sum="src/searchbench-go/go.sum"
            gobuckify_json="src/searchbench-go/gobuckify.json"
            vendor_dir="src/searchbench-go/vendor"
            manifest="$vendor_dir/.searchbench-vendor-projection.json"

            for f in "$go_mod" "$gobuckify_json"; do
              if [[ ! -f "$f" ]]; then
                echo "error: missing $f" >&2
                exit 1
              fi
            done

            echo "project-go-deps: go mod vendor"
            (cd src/searchbench-go && go mod vendor)

            echo "project-go-deps: gobuckify"
            (cd src/searchbench-go && buck2 run prelude//go/tools/gobuckify:gobuckify -- .)

            hash_file() {
              if [[ -f "$1" ]]; then
                sha256sum "$1" | awk '{print $1}'
              else
                echo "missing"
              fi
            }

            mkdir -p "$vendor_dir"
            jq -n \
              --arg generated_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
              --arg go_mod "$(hash_file "$go_mod")" \
              --arg go_sum "$(hash_file "$go_sum")" \
              --arg gobuckify "$(hash_file "$gobuckify_json")" \
              --arg flake_lock "$(hash_file flake.lock)" \
              --arg toolchains_lock "$(hash_file toolchains/flake.lock)" \
              --arg go_version "$(go version | head -n1)" \
              --arg buck2_version "$(buck2 --version | head -n1)" \
              '{
                generated_at: $generated_at,
                inputs: {
                  "src/searchbench-go/go.mod": $go_mod,
                  "src/searchbench-go/go.sum": $go_sum,
                  "src/searchbench-go/gobuckify.json": $gobuckify,
                  "flake.lock": $flake_lock,
                  "toolchains/flake.lock": $toolchains_lock
                },
                tools: {
                  go: $go_version,
                  buck2: $buck2_version
                }
              }' > "$manifest"

            echo "project-go-deps: wrote $manifest"
          '';
        };

        vendorProjectionWarn = pkgs.writeShellApplication {
          name = "vendor-projection-warn";
          runtimeInputs = with pkgs; [
            coreutils
            jq
          ];
          text = ''
            set -euo pipefail
            vendor_dir="src/searchbench-go/vendor"
            manifest="$vendor_dir/.searchbench-vendor-projection.json"
            gobuckify_json="src/searchbench-go/gobuckify.json"

            warn() {
              echo ""
              echo "Go vendor projection is missing or stale."
              echo ""
              echo "Run:"
              echo "  nix run .#project-go-deps"
              echo ""
              echo "Buck will not regenerate Go deps during tests."
              echo ""
            }

            hash_file() {
              if [[ -f "$1" ]]; then
                sha256sum "$1" | awk '{print $1}'
              else
                echo "missing"
              fi
            }

            if [[ ! -f "$gobuckify_json" ]]; then
              echo "[warn] missing $gobuckify_json"
              exit 0
            fi

            if [[ ! -d "$vendor_dir" ]]; then
              warn
              echo "[warn] missing $vendor_dir/"
              exit 0
            fi

            if [[ ! -f "$manifest" ]]; then
              warn
              echo "[warn] missing $manifest"
              exit 0
            fi

            current_go_mod=$(hash_file src/searchbench-go/go.mod)
            current_go_sum=$(hash_file src/searchbench-go/go.sum)
            current_gobuckify=$(hash_file "$gobuckify_json")
            current_flake=$(hash_file flake.lock)
            current_toolchains=$(hash_file toolchains/flake.lock)

            manifest_go_mod=$(jq -r '.inputs["src/searchbench-go/go.mod"] // empty' "$manifest")
            manifest_go_sum=$(jq -r '.inputs["src/searchbench-go/go.sum"] // empty' "$manifest")
            manifest_gobuckify=$(jq -r '.inputs["src/searchbench-go/gobuckify.json"] // empty' "$manifest")
            manifest_flake=$(jq -r '.inputs["flake.lock"] // empty' "$manifest")
            manifest_toolchains=$(jq -r '.inputs["toolchains/flake.lock"] // empty' "$manifest")

            stale=false
            check_stale() {
              local name="$1" cur="$2" man="$3"
              if [[ -z "$man" || "$cur" != "$man" ]]; then
                stale=true
                echo "[warn] vendor projection stale: $name"
              fi
            }
            check_stale go.mod "$current_go_mod" "$manifest_go_mod"
            check_stale go.sum "$current_go_sum" "$manifest_go_sum"
            check_stale gobuckify.json "$current_gobuckify" "$manifest_gobuckify"
            check_stale flake.lock "$current_flake" "$manifest_flake"
            check_stale toolchains/flake.lock "$current_toolchains" "$manifest_toolchains"

            if [[ "$stale" == true ]]; then
              warn
            fi
          '';
        };

      in
      {
        formatter = pkgs.nixfmt;

        checks.pre-commit-check = preCommitCheck;

        apps.project-go-deps = {
          type = "app";
          program = "${projectGoDeps}/bin/project-go-deps";
        };

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
            ${vendorProjectionWarn}/bin/vendor-projection-warn || true
            echo "searchbench-go: dev shell (pre-commit → Buck2 + hygiene; see AGENTS.md)"
            go version
            echo ""
            echo "Go vendor projection (generated locally; not in git):"
            echo "  nix run .#project-go-deps"
            echo ""
            echo "Python/IC: uv lock && uv sync in src/iterative-context; Buck uses wrapper targets only (no Elk)."
          '';
        };
      }
    );
}
