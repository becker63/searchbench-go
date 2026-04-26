#!/usr/bin/env nix-shell
#! nix-shell -i bash -p bash coreutils gnugrep gnused gawk iproute2 lsof dnsutils procps strace tcpdump

set -euo pipefail

DOMAINS_DEFAULT="api.openai.com chatgpt.com chat.openai.com gateway.openai.com"
DOMAINS="${OPENAI_DOMAINS:-$DOMAINS_DEFAULT}"
INTERVAL="${INTERVAL:-2}"

usage() {
  cat <<'EOF'
Usage:
  ./openai-netwatch list
  ./openai-netwatch watch
  ./openai-netwatch sniff
  ./openai-netwatch trace <pid>

What it does:
  list   - show active HTTPS connections to OpenAI/ChatGPT-related hosts
  watch  - refresh the connection list repeatedly
  sniff  - show packet activity to those hosts with tcpdump
  trace  - attach strace to a Codex/OpenAI process and show network syscalls

Notes:
  - This cannot read request/response bodies because HTTPS is encrypted.
  - One TCP connection may carry many OpenAI requests via HTTP/2, so this counts active connections, not exact API calls.
  - Run with sudo if process names/PIDs or tcpdump output are incomplete.

Environment:
  OPENAI_DOMAINS="api.openai.com chatgpt.com ..." ./openai-netwatch watch
  INTERVAL=1 ./openai-netwatch watch
EOF
}

resolve_ips() {
  for domain in $DOMAINS; do
    dig +short A "$domain" || true
    dig +short AAAA "$domain" || true
  done \
    | grep -E '^[0-9a-fA-F:.]+$' \
    | sort -u
}

print_targets() {
  echo "Domains:"
  for domain in $DOMAINS; do
    echo "  - $domain"
  done

  echo
  echo "Resolved IPs:"
  resolve_ips | sed 's/^/  - /' || true
}

list_connections() {
  mapfile -t IPS < <(resolve_ips)

  if [[ "${#IPS[@]}" -eq 0 ]]; then
    echo "No IPs resolved for: $DOMAINS"
    exit 1
  fi

  echo "OpenAI/ChatGPT-ish active HTTPS connections"
  echo "Time: $(date)"
  echo

  printf "%-8s %-8s %-24s %-24s %-10s %s\n" \
    "PID" "STATE" "LOCAL" "REMOTE" "QUEUE" "PROCESS"
  printf "%-8s %-8s %-24s %-24s %-10s %s\n" \
    "--------" "--------" "------------------------" "------------------------" "----------" "----------------"

  ss -H -tnp state established '( dport = :443 or sport = :443 )' 2>/dev/null \
    | while read -r line; do
        matched=0
        for ip in "${IPS[@]}"; do
          if grep -q "$ip" <<<"$line"; then
            matched=1
            break
          fi
        done

        [[ "$matched" -eq 1 ]] || continue

        state="$(awk '{print $1}' <<<"$line")"
        recvq="$(awk '{print $2}' <<<"$line")"
        sendq="$(awk '{print $3}' <<<"$line")"
        local_addr="$(awk '{print $4}' <<<"$line")"
        remote_addr="$(awk '{print $5}' <<<"$line")"
        queue="${recvq}/${sendq}"

        pid="$(grep -oE 'pid=[0-9]+' <<<"$line" | head -n1 | cut -d= -f2 || true)"
        proc="unknown"

        if [[ -n "${pid:-}" && -r "/proc/$pid/cmdline" ]]; then
          proc="$(tr '\0' ' ' < "/proc/$pid/cmdline" | sed 's/[[:space:]]*$//')"
          [[ -n "$proc" ]] || proc="$(cat "/proc/$pid/comm" 2>/dev/null || echo unknown)"
        fi

        printf "%-8s %-8s %-24s %-24s %-10s %s\n" \
          "${pid:-?}" "$state" "$local_addr" "$remote_addr" "$queue" "$proc"
      done

  echo
  echo "Detailed TCP info for matching connections:"
  echo

  for ip in "${IPS[@]}"; do
    ss -tinp dst "$ip":443 2>/dev/null || true
    ss -tinp src "$ip":443 2>/dev/null || true
  done | sed '/^$/d' | sed 's/^/  /'
}

watch_connections() {
  while true; do
    clear
    print_targets
    echo
    list_connections
    echo
    echo "Refreshing every ${INTERVAL}s. Ctrl-C to stop."
    sleep "$INTERVAL"
  done
}

sniff_packets() {
  mapfile -t IPS < <(resolve_ips)

  if [[ "${#IPS[@]}" -eq 0 ]]; then
    echo "No IPs resolved for: $DOMAINS"
    exit 1
  fi

  filter="tcp port 443 and ("
  first=1
  for ip in "${IPS[@]}"; do
    if [[ "$first" -eq 0 ]]; then
      filter+=" or "
    fi
    filter+="host $ip"
    first=0
  done
  filter+=")"

  echo "Running tcpdump with filter:"
  echo "  $filter"
  echo
  echo "You may need sudo. Packet flow means the session is not completely idle."
  echo

  exec tcpdump -i any -nn -tttt "$filter"
}

trace_pid() {
  local pid="${1:-}"

  if [[ -z "$pid" ]]; then
    echo "Missing PID."
    echo
    usage
    exit 1
  fi

  if [[ ! -d "/proc/$pid" ]]; then
    echo "No such PID: $pid"
    exit 1
  fi

  echo "Tracing network syscalls for PID $pid."
  echo "You may need sudo."
  echo
  exec strace -f -p "$pid" \
    -e trace=network,read,write,poll,ppoll,select,pselect6,epoll_wait,epoll_pwait,connect,recvfrom,sendto \
    -s 128
}

cmd="${1:-list}"

case "$cmd" in
  list)
    print_targets
    echo
    list_connections
    ;;
  watch)
    watch_connections
    ;;
  sniff)
    sniff_packets
    ;;
  trace)
    trace_pid "${2:-}"
    ;;
  -h|--help|help)
    usage
    ;;
  *)
    echo "Unknown command: $cmd"
    echo
    usage
    exit 1
    ;;
esac
