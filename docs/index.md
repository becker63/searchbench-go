# SearchBench documentation

**Test the interfaces your coding agents use.**

SearchBench is a work-in-progress harness for evaluating agent-facing interfaces over benchmark tasks. It is **not** a mature public API yet.

## What SearchBench is for

Most benchmarks ask which **model** is better.

SearchBench asks which **interface** makes the **same model** behave better — tools, code-search backends, MCP servers, graph lookahead, configs, and validation surfaces.

The **first game** is **code localization**: bug-localization dataset slices stress **symbol/code-search interfaces with lookahead**. The same `Game` model is intended to wrap other benchmark families later (for example SWE-bench-style issue resolution); those are directions, not shipped products today.

```text
Game → Round → Match → Evidence → Decision → NextChallenger
```

```text
Bug-localization dataset slice
  → Code-localization game
  → Symbol/search interface candidate
  → Agent run
  → Predicted files
  → Evidence bundle
  → Decision
```

Research thesis (long form): [AGENT_INTERFACE_RESEARCH.md](https://github.com/becker63/searchbench-go/blob/main/AGENT_INTERFACE_RESEARCH.md) on GitHub.

## Start here

- [Start here](./start-here.md) — orientation and one local round
- [Concepts](./concepts.md) — Game, Interface, Round, Evidence, …
- [Architecture](./architecture.md) — layers and lifecycle
- [Development](./development.md) — Nix, Buck2, Go, validation
- [Workspace seeds](./workspace-seeds.md) — `local_path` vs `buck_descriptor`

Full index: [Docs README](./README.md).

Contributors: [AGENTS.md](https://github.com/becker63/searchbench-go/blob/main/AGENTS.md).
