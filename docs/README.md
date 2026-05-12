# SearchBench-Go documentation

Project vocabulary and layering live under **`architecture/`**. Engineering practices under **`engineering/`**, and **`roadmap/`** tracks high-level implementation pressure.

| Area | Contents |
| --- | --- |
| [Architecture](architecture/architecture.md) | Naming model, flows, bundles, migrations |
| [Visualization](architecture/visualization.md) | Product visualization plan |
| [Integration shape](architecture/integration-shape.md) | Package layering and adapters vs agents |
| [Package boundaries](architecture/package-boundaries.md) | `internal/` import rules (mirrors [`internal/architecture/imports_test.go`](../internal/architecture/imports_test.go)) |
| [Pkl round manifests](architecture/pkl-round-manifests.md) | Manifest-facing notes |
| [Pkl scoring interface](architecture/pkl-scoring-interface.md) | Scoring interface notes |
| [LangSmith](integrations/langsmith-integration.md) | Trace/evaluator platform positioning |
| [Replit](guides/replit.md) | Quick environment and tech stack orientation |
| [Roadmap](roadmap/todo.md) | High-level implementation pressure |
| [Fake e2e runs](roadmap/fake-e2e-runs.md) | Offline fake-round commands, manifests, and artifact expectations |
| [Engineering](engineering/) | Agentic workflow, issue style, testing, pure center |
| [Optimizer policy validation](engineering/optimizer-policy-validation.md) | IC NextChallenger pipeline steps, allowlist, canonical `score_fn` |
| [Issue wave publisher](engineering/issue-wave-manifest.md) | `gh`-backed JSON manifest for batch issue creation (dev tooling) |

Read **`AGENTS.md`** at the repository root first; it lists the canonical “start here” paths for contributors and automation.
