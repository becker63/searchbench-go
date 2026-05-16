# Architecture

Monorepo component map: [components.md](./components.md). This page is **internal Go layering** inside `src/searchbench-go/`.

Generic harness shape: **Game** → dataset slice → incumbent vs challenger → **bundle**. First game implemented: **code localization**. Code layout:

```text
pure      model (Game, Round, Match, evidence, scores)
ports     shared contracts (e.g. pipeline steps)
app       round orchestration (internal/app/round)
agents    evaluator + optimizer (must not import app)
adapters  Pkl config, FS bundles, subprocess pipelines
surface   CLI
games     game wiring (e.g. code localization)
```

```text
surface / app / adapters / agents  →  ports  →  pure
```

## Round lifecycle

`src/searchbench-go/internal/app/round`:

```text
resolve Pkl manifest → run matches → evidence + objective → decision → bundle
  → optional NextChallenger (optimizer)
```

IC optimizer path: [workspace-seeds.md](./workspace-seeds.md) → `ValidateProposalInWorkspace` → MCP launch from the validated workspace.

## Buck2

Contributor operation graph: `//:check`, `//:check_full`, `//docs:check`, IC `//src/iterative-context:check*`. Not required for public `local_path` round runs.

## Reference

| Topic | Doc |
| --- | --- |
| Import rules | [reference/package-boundaries.md](./reference/package-boundaries.md) |
| Pkl rounds | [reference/pkl-rounds.md](./reference/pkl-rounds.md) |
| Pkl objectives | [reference/pkl-objectives.md](./reference/pkl-objectives.md) |
| Bundles | [reference/bundles.md](./reference/bundles.md) |
| IC optimizer pipeline | [reference/optimizer-policy-validation.md](./reference/optimizer-policy-validation.md) |

Enforced by `src/searchbench-go/internal/architecture/imports_test.go`.
