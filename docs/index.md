# SearchBench docs

**Test the interfaces your coding agents use.**

SearchBench is a **work-in-progress** harness for evaluating agent-facing interfaces over benchmark tasks — not a stable public API yet. It wraps benchmark families as **games**, compares **incumbent** vs **challenger** interfaces on the same dataset slice, and records **bundles** for promote / review / reject decisions.

The **first game** is **code localization** (symbol/code-search with lookahead on bug-localization slices). Other task families are research directions, not shipped products.

```text
Game → Round → Match → Evidence → Decision → NextChallenger
```

## Read next

| Doc | Purpose |
| --- | --- |
| [Start here](./start-here.md) | Fast orientation and one local round |
| [Concepts](./concepts.md) | Vocabulary |
| [Architecture](./architecture.md) | Package layers and round lifecycle |
| [Development](./development.md) | Nix, Buck2, Go, hooks, docs build |
| [Workspace seeds](./workspace-seeds.md) | `local_path` vs `buck_descriptor` |

## Reference

Implementation detail: [reference/package-boundaries.md](./reference/package-boundaries.md), [reference/pkl-rounds.md](./reference/pkl-rounds.md), [reference/pkl-objectives.md](./reference/pkl-objectives.md), [reference/bundles.md](./reference/bundles.md), [reference/optimizer-policy-validation.md](./reference/optimizer-policy-validation.md).

## Research

Long-form thesis and experiments: [research/agent-interface-research.md](./research/agent-interface-research.md), [research/bxl-meta-harness.md](./research/bxl-meta-harness.md).

## Links

- [Repository](https://github.com/becker63/searchbench-go)
- [Hosted docs](https://becker63.github.io/searchbench-go/)
- [Contributors: AGENTS.md](../AGENTS.md)
