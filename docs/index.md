# SearchBench docs

**Test the interfaces your coding agents use.**

SearchBench is a **work-in-progress** harness for evaluating agent-facing interfaces over benchmark tasks — not a stable public API yet. It wraps benchmark families as **games**, compares **incumbent** vs **challenger** interfaces on the same dataset slice, and records **bundles** for promote / review / reject decisions.

The **first game** is **code localization** (symbol/code-search with lookahead on bug-localization slices).

```text
Game → Round → Match → Evidence → Decision → NextChallenger
```

## Current interface

SearchBench’s user-facing surface today:

| Surface | Example |
| --- | --- |
| Pkl round manifest | `configs/rounds/local-ic-vs-jcodemunch/round.pkl` |
| CLI | `./searchbench run --manifest=... --bundle-root=...` |
| Evidence bundle | `configs/rounds/local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001/` |
| Continuation | `continuation.pkl` in that bundle |
| Proof gates | `buck2 test //:check`, `buck2 test //:check_full` |

Excerpt from `configs/rounds/local-ic-vs-jcodemunch/round.pkl`:

```pkl
round = (game.defineFromScratch("round-001")) {
  incumbent = game.jcodemunch()
  challenger = game.iterativeContext("policies/challenger_policy.py") { ... }
  matches = game.lca("py", "dev", 5)
  evaluator = game.fakeEvaluator()
  scoring = game.objective("scoring/localization-objective.pkl")
}
```

**Meaning:** Pkl declares intent; Go runs matches, scores evidence, writes the bundle. See [start-here.md](./start-here.md).

## Read next

| Doc | Purpose |
| --- | --- |
| [Start here](./start-here.md) | Manifest → CLI → bundle walkthrough |
| [Concepts](./concepts.md) | Vocabulary with file anchors |
| [Components](./components.md) | Monorepo trees, example paths, proof targets |
| [Architecture](./architecture.md) | Go lifecycle and implementation paths |
| [Development](./development.md) | Nix, Buck gates, docs build |

## Reference

| Doc | Anchors on |
| --- | --- |
| [pkl-rounds.md](./reference/pkl-rounds.md) | `configs/rounds/*/round.pkl`, `configs/schema/SearchBenchRound.pkl` |
| [pkl-objectives.md](./reference/pkl-objectives.md) | `configs/rounds/local-ic-vs-jcodemunch/scoring/localization-objective.pkl` |
| [bundles.md](./reference/bundles.md) | Checked-in `round-001` artifact tree |
| [candidate-workspaces.md](./candidate-workspaces.md) | Sandboxing, `local_path` / `buck_descriptor` |
| [package-boundaries.md](./reference/package-boundaries.md) | Go import rules |
| [optimizer-policy-validation.md](./reference/optimizer-policy-validation.md) | IC candidate pipeline |

## Research

[research/agent-interface-research.md](./research/agent-interface-research.md), [research/bxl-meta-harness.md](./research/bxl-meta-harness.md).

## Links

- [Repository](https://github.com/becker63/searchbench-go)
- [Hosted docs](https://becker63.github.io/searchbench-go/)
- [Contributors: AGENTS.md](../AGENTS.md)
