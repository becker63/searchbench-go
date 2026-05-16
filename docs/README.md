# SearchBench docs

Work-in-progress harness for evaluating **agent-facing interfaces** over benchmark tasks. Start at [start-here.md](./start-here.md) or the [docs site landing](./index.md). Contributors: [AGENTS.md](../AGENTS.md).

**Research thesis:** [AGENT_INTERFACE_RESEARCH.md](https://github.com/becker63/searchbench-go/blob/main/AGENT_INTERFACE_RESEARCH.md)

## Start here

- [Start here](./start-here.md) — orientation and one local round
- [Concepts](./concepts.md) — Game, Interface, dataset slice, Round, bundle, decision
- [Architecture](./architecture.md) — system spine and package boundaries
- [Development](./development.md) — Nix, Buck2, Go, Repomix, validation commands
- [Workspace seed providers](./workspace-seeds.md) — `local_path` vs `buck_descriptor`, Pkl config, decision record

## Reference

- [Architecture (full)](./reference/architecture-full.md) — extended narrative, bundles, migrations
- [Package boundaries](./reference/package-boundaries.md)
- [Integration shape](./reference/integration-shape.md)
- [Build system](./reference/build-system.md)
- [Pkl round manifests](./reference/pkl-round-manifests.md)
- [Pkl scoring interface](./reference/pkl-scoring-interface.md)
- [Optimizer policy validation](./reference/optimizer-policy-validation.md)
- [LangSmith integration](./reference/langsmith-integration.md)
- [Visualization plan](./reference/visualization.md)
- [Agentic development flow](./reference/agentic-development-flow.md)
- [Pure center](./reference/pure-center.md)
- [Issue style guide](./reference/issue-style-guide.md)
- [Model testing](./reference/model-testing.md)

## Roadmap

- [Todo](./roadmap/todo.md)
- [Fake e2e runs](./roadmap/fake-e2e-runs.md)

## Guides

- [Replit](./guides/replit.md)

## Archive / research

- [Archive index](./archive/README.md)
- [Blog diff candidates](./archive/blog-diff-candidates/)
- [Issue wave manifest](./archive/issue-wave-manifest.md)

Legacy paths under `docs/architecture/` and `docs/engineering/ic-*.md` redirect to the spine above.
