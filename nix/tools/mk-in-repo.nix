# Shared helper: run body from the git repo root that contains go.mod.
{ pkgs }:
{
  name,
  text,
  runtimeInputs ? [ ],
}:
pkgs.writeShellApplication {
  inherit name;
  runtimeInputs = [ pkgs.git ] ++ runtimeInputs;
  text = ''
    set -euo pipefail
    _root="$(git rev-parse --show-toplevel 2>/dev/null || true)"
    if [[ -z "$_root" || ! -f "$_root/go.mod" ]]; then
      echo "${name}: not at the root of the searchbench-go git repository" >&2
      exit 1
    fi
    cd "$_root"
  ''
  + text;
}
