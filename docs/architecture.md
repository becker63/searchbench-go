# Architecture

SearchBench keeps deterministic model code pure, orchestrates rounds in `app`, colocates agent behavior under `agents`, and pushes filesystem/MCP/Pkl edges into `adapters` and `surface`.

```text
pure      stable SearchBench model
ports     project-owned contracts
app       round lifecycle orchestration
agents    evaluator + optimizer (prompt, Eino, policy)
adapters  config, bundles, pipelines, scoring execution
surface   CLI and human presentation
games     concrete game wiring (e.g. code localization)
```

Dependency direction:

```text
surface / app / adapters / agents
    → ports (where shared)
    → pure
```

`agents` does **not** import `app`; `app/round` composes agents.

## Round lifecycle

Primary orchestration: `src/searchbench-go/internal/app/round`

```text
resolve manifest (Pkl)
  → run matches (evaluator)
  → evidence + objective (pure + scoring adapters)
  → decision
  → persist bundle (adapters/bundle)
  → optional NextChallenger (optimizer)
```

## IC optimizer path

For Iterative Context selection policies:

```text
WorkspaceSeedProvider → WorkspaceSeed → ICCandidateWorkspace
  → ValidateProposalInWorkspace → AcceptedICCandidate → MCP launch
```

See [workspace-seeds.md](./workspace-seeds.md).

## Buck in the repo

Buck2 is the **repo operation graph** for contributors: `//:check`, `//:check_full`, IC `//src/iterative-context:check*`, and the optimizable backend descriptor target. It is not required for public round runs that use `local_path` seeds only.

## Boundaries (enforced)

Import rules are tested in `src/searchbench-go/internal/architecture/imports_test.go`. Summary:

| Package | Owns |
| --- | --- |
| `internal/pure` | Game/Round/Match vocabulary, reports, scores, optimizer records |
| `internal/ports` | Shared contracts (e.g. pipeline steps) |
| `internal/app` | Round workflow composition |
| `internal/agents` | Evaluator/optimizer prompts, Eino runners, policy validation |
| `internal/adapters` | Pkl config, FS bundles, subprocess pipelines |
| `internal/surface` | CLI |

Do not add live MCP, LangSmith, provider execution, or visualization UI unless the task explicitly requires it.

## Reference docs

| Topic | Doc |
| --- | --- |
| Full architecture narrative | [reference/architecture-full.md](./reference/architecture-full.md) |
| Package import rules | [reference/package-boundaries.md](./reference/package-boundaries.md) |
| Integration / adapter map | [reference/integration-shape.md](./reference/integration-shape.md) |
| Pkl round manifests | [reference/pkl-round-manifests.md](./reference/pkl-round-manifests.md) |
| Pkl scoring | [reference/pkl-scoring-interface.md](./reference/pkl-scoring-interface.md) |
| Build system detail | [reference/build-system.md](./reference/build-system.md) |
| Visualization plan | [reference/visualization.md](./reference/visualization.md) |
| LangSmith | [reference/langsmith-integration.md](./reference/langsmith-integration.md) |
