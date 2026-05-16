# Architecture

Monorepo map: [components.md](./components.md). This page ties the **round lifecycle** to Go paths.

```text
pure      model (Game, Round, Match, evidence, scores)
ports     shared contracts
app       round orchestration
agents    evaluator + optimizer (must not import app)
adapters  Pkl, bundles, pipelines, workspaces
surface   CLI
games     code-localization wiring
```

```text
surface / app / adapters / agents  →  ports  →  pure
```

## Round lifecycle → code

| Stage | Path |
| --- | --- |
| Load Pkl manifest | `internal/adapters/config/pkl/` |
| Run round | `internal/app/round/` |
| Match execution / reports | `internal/app/round/` (evaluator wiring) |
| Build evidence | `internal/pure/report`, `internal/pure/score` |
| Run objective | `internal/adapters/scoring/pkl/` |
| Write bundle | `internal/adapters/bundle/fs/` |
| Candidate workspace | `internal/adapters/workspace/` |
| Optimizer policy validation | `internal/agents/optimizer/policy/` |

```text
round.pkl → config/pkl → app/round → matches → evidence.pkl
  → scoring/pkl → objective.json → decision → bundle/fs
  → optional optimizer → candidate workspace → MCP
```

**Example manifest:** `configs/rounds/local-ic-vs-jcodemunch/round.pkl`
**Example bundle:** `configs/rounds/local-ic-vs-jcodemunch/artifacts/.../round-001/`

## IC optimizer path

[candidate-workspaces.md](./candidate-workspaces.md) — `ValidateProposalInWorkspace` then MCP from the validated workspace only.

## Buck2 (contributors)

| Gate | Target |
| --- | --- |
| Fast (commit) | `buck2 test //:check` |
| Full (push) | `buck2 test //:check_full` |
| Docs | `buck2 test //docs:check` |

Not required for public `local_path` round runs.

## Reference

| Topic | Doc |
| --- | --- |
| Import rules | [reference/package-boundaries.md](./reference/package-boundaries.md) |
| Pkl rounds | [reference/pkl-rounds.md](./reference/pkl-rounds.md) |
| Objectives | [reference/pkl-objectives.md](./reference/pkl-objectives.md) |
| Bundles | [reference/bundles.md](./reference/bundles.md) |

Enforced by `internal/architecture/imports_test.go`.
