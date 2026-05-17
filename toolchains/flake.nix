{
  description = "SearchBench Buck2 Nix cell (exposes host tools to Starlark)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    buck2-nix = {
      url = "github:tweag/buck2.nix";
      flake = false;
    };
  };

  outputs =
    { nixpkgs, buck2-nix, ... }:
    let
      inherit (nixpkgs) lib;
      forAllSystems = lib.genAttrs lib.systems.flakeExposed;
    in
    {
      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            packages = [
              pkgs.go
              pkgs.git
              pkgs.pkl
              pkgs.ruff
              pkgs.uv
              pkgs.buck2
              pkgs.python3
              pkgs.nodejs
            ];
            shellHook = ''
              _root=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
              mkdir -p "$_root/.buckconfig.d"
              cat >"$_root/.buckconfig.d/buck2-nix.config" <<EOS
              [external_cell_nix]
                git_origin = https://github.com/tweag/buck2.nix.git
                commit_hash = ${buck2-nix.rev}
              EOS
            '';
          };
        }
      );

      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          inherit (pkgs)
            go
            git
            pkl
            ruff
            uv
            python3
            nodejs
            ;
        }
      );
    };
}
