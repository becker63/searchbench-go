# SearchBench-Go

**SearchBench-Go** is a review-first system for evaluating AI policy changes on structured matches. Each **Round** pits an **IncumbentPolicy** against a **ChallengerPolicy** on the same slice of **Matches**, gathers durable **Evidence**, records a **Decision**, and optionally proposes a **NextChallenger** so the workflow can iterate.

```text
Game → Round → Match → Evidence → Decision → NextChallenger
```

The Go implementation keeps deterministic model code and artifact semantics explicit, separates round orchestration from agent-specific prompts and runners, and uses Pkl manifests as the human-facing round surface. This repository pairs a small CLI (`searchbench`) with a layered `internal/` tree described in [`docs/architecture/integration-shape.md`](docs/architecture/integration-shape.md).

---

## Highlights

| Area | What you get |
| --- | --- |
| **Orchestration** | A single lifecycle in [`internal/app/round`](internal/app/round): resolve manifest, run comparisons, score objectives, persist bundles |
| **Agents** | Evaluator and optimizer live in [`internal/agents`](internal/agents) (prompts, Eino execution, optimizer bundle + policy helpers) |
| **Model** | Stable vocabulary under [`internal/pure`](internal/pure); import boundaries enforced by [`internal/architecture/imports_test.go`](internal/architecture/imports_test.go) |
| **Config & scoring** | Pkl manifests and objective modules via [`internal/adapters/config/pkl`](internal/adapters/config/pkl) and [`internal/adapters/scoring/pkl`](internal/adapters/scoring/pkl) |
| **Local path** | Deterministic fake match sources and evaluator-side stubs for end-to-end runs without live backends |

---

## Requirements

| Tool | Notes |
| --- | --- |
| [Go 1.26.2](https://go.dev/dl/) | Matches [`go.mod`](go.mod). |
| [Pkl](https://pkl-lang.org/) CLI | Needed on `PATH` to evaluate example `*.pkl` manifests (`pkl`). |
| *(optional)* | Real model providers/backends once you graduate past the bundled fake-local path. |
| [Nix](https://nixos.org/) with flakes *(optional)* | `nix develop` for the full toolchain, pre-commit, and `searchbench-*` commands; `nix flake check` for sandboxed CI-like checks. |

---

## Nix workflow (optional)

```bash
nix develop              # shell with Go, Pkl, hooks, searchbench-* tools
nix develop -c pre-commit run --all-files
nix develop -c searchbench-staticcheck
nix develop -c searchbench-golangci
nix flake check
nix run .#e2e             # root package integration tests
nix run .#update-repomix  # refresh committed repomix-output.xml
```

See [`AGENTS.md`](AGENTS.md) for Repomix rationale, hook tiers (flake vs pre-commit vs pre-push vs agent merge-check), and command reference.

---

## Quick start

Clone and build:

```bash
git clone https://github.com/becker63/searchbench-go.git
cd searchbench-go
go build -o searchbench ./cmd/searchbench
```

Run a local evaluation manifest (writes bundle artifacts under a temp bundles root):

```bash
./searchbench round run \
  --manifest=configs/rounds/local-ic-vs-jcodemunch/round.pkl \
  --bundle-root="$(pwd)/.tmp-artifacts"
```

Shorthand (`run` at the root of the CLI tree mirrors the hidden `searchbench run …` shortcut used in tests):

```bash
./searchbench run \
  --manifest=configs/rounds/local-ic-vs-jcodemunch/round.pkl \
  --bundle-root="$(pwd)/.tmp-artifacts"
```

Useful flags (see [`internal/surface/cli/run.go`](internal/surface/cli/run.go)):

| Flag | Role |
| --- | --- |
| `--manifest` | Path to the round manifest (`*.pkl`) |
| `--bundle-root` | Override where round bundles collect |
| `--bundle-id` | Optional bundle / round identifier |
| `--no-human-report` | Skip optional human-readable report artifact |

Successful runs summarize `bundle`, `report_id`, `objective`, and final score metrics on stdout.

---

## Repository layout

| Path | Responsibility |
| --- | --- |
| [`cmd/searchbench`](cmd/searchbench) | CLI entrypoint delegating into [`internal/surface/cli`](internal/surface/cli) |
| [`internal/app/round`](internal/app/round) | Round lifecycle: resolve, evaluate, evidence, objective, decide, bundle, proposed next challenger |
| [`internal/agents`](internal/agents) | Evaluator / optimizer vertical slices (prompt + Eino + local fakes) |
| [`internal/pure`](internal/pure) | Deterministic SearchBench types (reports, scoring, execution records, optimizer records, …) |
| [`internal/adapters`](internal/adapters) | Shared runtime edges: filesystem bundles, Pkl config, pipelines, scoring execution |
| [`internal/games`](internal/games) | Concrete games (today: code localization game wiring) |
| [`configs/schema`](configs/schema) | Authoritative [`SearchBenchRound.pkl`](configs/schema/SearchBenchRound.pkl) schema |
| [`configs/rounds`](configs/rounds) | Example manifests and support files |

Root [`doc.go`](doc.go) and [`e2e_test.go`](e2e_test.go) host canonical integration checks that compile the CLI and exercise a fake-local round via temp directories only.

---

## Documentation

Full documentation hub: **[`docs/README.md`](docs/README.md)**.

Suggested reading order for humans or coding agents:

1. **[`AGENTS.md`](AGENTS.md)** — project rules, vocabulary, and validation commands at the repo root
2. **[`docs/architecture/architecture.md`](docs/architecture/architecture.md)** — architecture spine and product vocabulary
3. **[`docs/architecture/package-boundaries.md`](docs/architecture/package-boundaries.md)** — dependency rules mirrored in tests

Additional topics: visualization plan, roadmap, onboarding ([`docs/guides/replit.md`](docs/guides/replit.md)), LangSmith positioning, engineering workflow, Pkl seams.

---

## Development

Run the entire test suite (including CLI and architecture import checks):

```bash
go test ./...
```

Regenerate Go bindings whenever you edit [`configs/schema/SearchBenchRound.pkl`](configs/schema/SearchBenchRound.pkl):

```bash
pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl --output-path=. configs/schema/SearchBenchRound.pkl
```

---

## Name and semantics

Prefer **Game**, **Round**, **Match**, **IncumbentPolicy**, **ChallengerPolicy**, **Evidence**, **Decision**, and **NextChallenger** in code and docs—the same vocabulary the CLI manifests and schemas encode. Older terms linger only where historical prompts or upstream datasets demand them (call those boundaries out explicitly).

---

## Module

Go module path: **`github.com/becker63/searchbench-go`**.

For questions about intent and layering, open [`docs/README.md`](docs/README.md) or start from [`AGENTS.md`](AGENTS.md).
