# SearchBench-Go

**SearchBench-Go** is a review-first system for evaluating AI policy changes on structured matches. Each **Round** pits an **IncumbentPolicy** against a **ChallengerPolicy** on the same slice of **Matches**, gathers durable **Evidence**, records a **Decision**, and optionally proposes a **NextChallenger** so the workflow can iterate.

```text
Game → Round → Match → Evidence → Decision → NextChallenger
```

The Go implementation keeps deterministic model code and artifact semantics explicit, separates round orchestration from agent-specific prompts and runners, and uses Pkl manifests as the human-facing round surface. This repository pairs a small CLI (`searchbench`) with a layered Go tree under [`src/searchbench-go/internal/`](src/searchbench-go/internal/) described in [`docs/architecture/integration-shape.md`](docs/architecture/integration-shape.md).

---

## Highlights

| Area | What you get |
| --- | --- |
| **Orchestration** | A single lifecycle in [`src/searchbench-go/internal/app/round`](src/searchbench-go/internal/app/round): resolve manifest, run comparisons, score objectives, persist bundles |
| **Agents** | Evaluator and optimizer live in [`src/searchbench-go/internal/agents`](src/searchbench-go/internal/agents) (prompts, Eino execution, optimizer bundle + policy helpers) |
| **Model** | Stable vocabulary under [`src/searchbench-go/internal/pure`](src/searchbench-go/internal/pure); import boundaries enforced by [`src/searchbench-go/internal/architecture/imports_test.go`](src/searchbench-go/internal/architecture/imports_test.go) |
| **Config & scoring** | Pkl manifests and objective modules via [`src/searchbench-go/internal/adapters/config/pkl`](src/searchbench-go/internal/adapters/config/pkl) and [`src/searchbench-go/internal/adapters/scoring/pkl`](src/searchbench-go/internal/adapters/scoring/pkl) |
| **Local path** | Deterministic fake match sources and evaluator-side stubs for end-to-end runs without live backends |

---

## Requirements

| Tool | Notes |
| --- | --- |
| [Go 1.26.2](https://go.dev/dl/) | Matches [`go.mod`](src/searchbench-go/go.mod). |
| [Pkl](https://pkl-lang.org/) CLI | Needed on `PATH` to evaluate example `*.pkl` manifests (`pkl`). |
| *(optional)* | Real model providers/backends once you graduate past the bundled fake-local path. |
| [Nix](https://nixos.org/) with flakes *(optional)* | `nix develop` puts Go, Pkl, **buck2**, hooks, and `repomix` on `PATH`; `nix flake check` runs sandboxed formatting/Nix hygiene only. |

---

## Nix workflow (optional)

```bash
nix develop                              # shell with Go, Pkl, buck2, hooks, repomix
nix develop -c pre-commit run --all-files
nix flake check
nix develop -c buck2 test //:check        # Go tests + CLI build + IC smoke
nix develop -c buck2 test //:check_full   # above + Repomix snapshot gate
```

See [`AGENTS.md`](AGENTS.md) for the Git hook lifecycle (pre-commit vs pre-push) and Repomix.

---

## Quick start

Clone and build:

```bash
git clone https://github.com/becker63/searchbench-go.git
cd searchbench-go
cd src/searchbench-go
go build -o ../../searchbench ./cmd/searchbench
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

Useful flags (see [`src/searchbench-go/internal/surface/cli/run.go`](src/searchbench-go/internal/surface/cli/run.go)):

| Flag | Role |
| --- | --- |
| `--manifest` | Path to the round manifest (`*.pkl`) |
| `--bundle-root` | Override where round bundles collect |
| `--bundle-id` | Optional bundle / round identifier |
| `--no-human-report` | Skip optional human-readable report artifact |

Successful runs summarize `bundle`, `report_id`, `objective`, and final score metrics on stdout.

---

## Real MCP / providers (opt-in)

The bundled manifest [`configs/rounds/local-ic-vs-jcodemunch/round.pkl`](configs/rounds/local-ic-vs-jcodemunch/round.pkl) and [`src/searchbench-go/e2e_test.go`](src/searchbench-go/e2e_test.go) exercise an **offline fake-local** path (no network).

To run the same CLI surface against **live MCP tool servers** and **real chat models**:

| Variable | Role |
| --- | --- |
| `SEARCHBENCH_JCODEMUNCH_COMMAND` | Shell command starting the jCodeMunch MCP server (stdio JSON-RPC). Required when the manifest uses that backend. |
| `SEARCHBENCH_ITERATIVE_CONTEXT_COMMAND` | Shell command starting the Iterative Context MCP server. Required for IC-backed policies. |
| Provider secrets / URLs | Resolved by [`src/searchbench-go/internal/adapters/providers/evaluatormodel`](src/searchbench-go/internal/adapters/providers/evaluatormodel) (OpenAI-compatible providers). |
| `LANGSMITH_API_KEY` | Optional LangSmith export; traces are **not** authoritative for scoring or decisions. |

Round compare prefers **tree-sitter indexing** for hop-distance scoring when `MatchSpec.Repo.Path` points at a materialized checkout **and** CGO is enabled (`src/searchbench-go/internal/app/round/treesitter_graph_provider.go`, `localization_graph_scorer.go`). Without a usable index it falls back to the deterministic fake graph from `src/searchbench-go/internal/agents/evaluator/fake`; hop metrics then follow that fake graph unless predictions and gold files yield call-graph paths in whatever graph was loaded.

Token/cost/composite scalars still originate from the same fallback scorer family (`LocalizationGraphScorer` merges graph-derived hops into the baseline metric set).

**GitHub issue batches:** see [`docs/engineering/issue-wave-manifest.md`](docs/engineering/issue-wave-manifest.md) (manifest format + manual `gh`/`jq` flow).

---

## Repository layout

| Path | Responsibility |
| --- | --- |
| [`src/searchbench-go/cmd/searchbench`](src/searchbench-go/cmd/searchbench) | CLI entrypoint delegating into [`src/searchbench-go/internal/surface/cli`](src/searchbench-go/internal/surface/cli) |
| [`src/searchbench-go/internal/app/round`](src/searchbench-go/internal/app/round) | Round lifecycle: resolve, evaluate, evidence, objective, decide, bundle, proposed next challenger |
| [`src/searchbench-go/internal/agents`](src/searchbench-go/internal/agents) | Evaluator / optimizer vertical slices (prompt + Eino + local fakes) |
| [`src/searchbench-go/internal/pure`](src/searchbench-go/internal/pure) | Deterministic SearchBench types (reports, scoring, execution records, optimizer records, …) |
| [`src/searchbench-go/internal/adapters`](src/searchbench-go/internal/adapters) | Shared runtime edges: filesystem bundles, Pkl config, pipelines, scoring execution |
| [`src/searchbench-go/internal/games`](src/searchbench-go/internal/games) | Concrete games (today: code localization game wiring) |
| [`src/iterative-context`](src/iterative-context) | Iterative Context submodule (Python MCP + validators) |
| [`configs/schema`](configs/schema) | Authoritative [`SearchBenchRound.pkl`](configs/schema/SearchBenchRound.pkl) schema |
| [`configs/rounds`](configs/rounds) | Example manifests and support files |

[`src/searchbench-go/doc.go`](src/searchbench-go/doc.go) and [`src/searchbench-go/e2e_test.go`](src/searchbench-go/e2e_test.go) host canonical integration checks that compile the CLI and exercise a fake-local round via temp directories only.

---

## Documentation

Full documentation hub: **[`docs/README.md`](docs/README.md)**.

Suggested reading order for humans or coding agents:

1. **[`AGENTS.md`](AGENTS.md)** — project rules, vocabulary, and **Git hook lifecycle** (`nix develop`, `git commit`, `git push`)
2. **[`docs/architecture/architecture.md`](docs/architecture/architecture.md)** — architecture spine and product vocabulary
3. **[`docs/architecture/package-boundaries.md`](docs/architecture/package-boundaries.md)** — dependency rules mirrored in tests

Additional topics: visualization plan, roadmap, onboarding ([`docs/guides/replit.md`](docs/guides/replit.md)), LangSmith positioning, engineering workflow, Pkl seams.

---

## Development

With **`nix develop`**, **`git commit`** and **`git push`** run the routine checks (see **`AGENTS.md`**). To run the full suite directly:

```bash
cd src/searchbench-go
go test ./...
```

Regenerate Go bindings whenever you edit [`configs/schema/SearchBenchRound.pkl`](configs/schema/SearchBenchRound.pkl):

```bash
cd src/searchbench-go
pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl --output-path=. ../../configs/schema/SearchBenchRound.pkl
```

---

## Name and semantics

Prefer **Game**, **Round**, **Match**, **IncumbentPolicy**, **ChallengerPolicy**, **Evidence**, **Decision**, and **NextChallenger** in code and docs—the same vocabulary the CLI manifests and schemas encode. Older terms linger only where historical prompts or upstream datasets demand them (call those boundaries out explicitly).

---

## Module

Go module path: **`github.com/becker63/searchbench-go`**.

For questions about intent and layering, open [`docs/README.md`](docs/README.md) or start from [`AGENTS.md`](AGENTS.md).
