# Optional diagnostics: network connection viewer for OpenAI-ish HTTPS traffic.
{ pkgs }:
let
  netwatchScript = builtins.readFile ./openai-netwatch.bash;
in
{
  searchbench-openai-netwatch = pkgs.writeShellApplication {
    name = "searchbench-openai-netwatch";
    runtimeInputs = with pkgs; [
      bash
      coreutils
      gnugrep
      gnused
      gawk
      iproute2
      lsof
      dnsutils
      procps
      strace
      tcpdump
    ];
    text = netwatchScript;
  };
}
