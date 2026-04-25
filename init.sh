#!/usr/bin/env bash
set -euo pipefail

MODULE="github.com/becker63/searchbench-go"

mkdir -p \
  cmd/searchbench \
  internal/domain \
  internal/backend \
  internal/codegraph \
  internal/run \
  internal/score \
  internal/report \
  testdata/fixtures/{tasks,repos,reports} \
  scripts

write_go_file() {
  local path="$1"
  local package="$2"

  if [[ -f "$path" ]]; then
    echo "skip existing $path"
    return
  fi

  cat > "$path" <<EOF
package $package

EOF

  echo "created $path"
}

# CLI
write_go_file "cmd/searchbench/main.go" "main"

# Domain: stable nouns
write_go_file "internal/domain/ids.go" "domain"
write_go_file "internal/domain/pair.go" "domain"
write_go_file "internal/domain/repo.go" "domain"
write_go_file "internal/domain/task.go" "domain"
write_go_file "internal/domain/system.go" "domain"
write_go_file "internal/domain/prediction.go" "domain"
write_go_file "internal/domain/usage.go" "domain"
write_go_file "internal/domain/artifact.go" "domain"

# Backend/session boundary
write_go_file "internal/backend/backend.go" "backend"
write_go_file "internal/backend/session.go" "backend"

# Code graph domain interface
write_go_file "internal/codegraph/node.go" "codegraph"
write_go_file "internal/codegraph/edge.go" "codegraph"
write_go_file "internal/codegraph/path.go" "codegraph"
write_go_file "internal/codegraph/graph.go" "codegraph"

# Run lifecycle
write_go_file "internal/run/phases.go" "run"
write_go_file "internal/run/spec.go" "run"
write_go_file "internal/run/record.go" "run"

# Scoring
write_go_file "internal/score/metric.go" "score"
write_go_file "internal/score/scoreset.go" "score"
write_go_file "internal/score/input.go" "score"
write_go_file "internal/score/scorer.go" "score"
write_go_file "internal/score/scored_run.go" "score"

# Report/release boundary
write_go_file "internal/report/comparison.go" "report"
write_go_file "internal/report/regression.go" "report"
write_go_file "internal/report/promotion.go" "report"
write_go_file "internal/report/report.go" "report"

# Minimal compiling main.
cat > cmd/searchbench/main.go <<'EOF'
package main

import "fmt"

func main() {
	fmt.Println("searchbench-go")
}
EOF

gofmt -w cmd internal || true

echo
echo "Type spine initialized."
echo "Next:"
echo "  go test ./..."
echo "  go run ./cmd/searchbench"
