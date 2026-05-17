# SearchBench-Go

<p align="center">
  <strong>Test the interfaces your coding agents use.</strong>
</p>

<p align="center">
  <strong>Pkl round manifests</strong> → <strong>Buck targets</strong> → <strong>evidence bundles</strong> → <strong>release decisions</strong>
</p>

<p align="center">
  <a href="https://becker63.github.io/searchbench-go/">📚 Docs</a>
  ·
  <a href="docs/start-here.md">🚀 Start here</a>
  ·
  <a href="docs/concepts.md">🧭 Concepts</a>
  ·
  <a href="docs/development.md">🛠️ Development</a>
  ·
  <a href="docs/research/agent-interface-research.md">🔬 Research</a>
</p>

---

SearchBench is a **work-in-progress harness** for evaluating **agent-facing interfaces** over benchmark tasks.

It wraps task families as **games**, then asks a product-shaped research question:

> Which interface makes the same agent perform better?

The first stress-test game is **code localization**: bug-localization dataset slices test whether symbol/code-search tools with **lookahead** help an agent find the files that need to change. Long term, the same `Game` model is intended to wrap other software-engineering task families, such as SWE-bench-style issue resolution, proof-target selection, and docs/context selection.

```text
Game → Round → Match → Evidence → Objective → Decision → NextChallenger
```

## Current API

The user-facing API for repo-owned work is a **Pkl round manifest** plus **Buck targets** that publish canonical bundles. The Go harness binary exists for Buck to invoke; it is not the normal public run interface.

A round declares the game, dataset slice, incumbent, challenger, evaluator, and scoring objective:

```pkl
amends "../../schema/games/code-localization.pkl"
import "../../schema/games/code-localization-helpers.pkl" as game

name = "example-local-ic-vs-jcodemunch-round-001"

round = (game.defineFromScratch("round-001")) {
  incumbent = game.jcodemunch()

  challenger = (game.iterativeContext("policies/challenger_policy.py")) {
    selectionPolicy {
      id = "challenger-policy-round-001"
    }
  }

  matches = game.lca("py", "dev", 5)
  evaluator = game.fakeEvaluator()
  scoring = game.objective("scoring/localization-objective.pkl")
}
```

Repo-owned runs use Buck (example local round):

```bash
buck2 test //:check_full
# Live round: see configs/rounds/live-ic-vs-jcodemunch/README.md
```

The clearest current API is the **Pkl config surface**, **Buck entrypoints**, and the **bundle format** (`report.json` first).

## Evidence bundles

A completed round writes a static bundle that can be reviewed by humans, tools, or a future meta-harness:

```text
artifacts/games/code-localization/rounds/round-001/
  COMPLETE
  report.json
  report.txt
  resolved-round.json
  round-report.json
  round-report.txt
  evidence.pkl
  objective.json
  decision.json
  metadata.json
  continuation.json
  continuation.pkl
  policies/challenger_policy.py
```

The bundle is the durable source of truth. Inspect **`report.json`** first, then detailed reports and evidence. It records the resolved round, scoring-facing evidence, objective values, decision, metadata, and continuation surface for the next round.

A later round can amend the previous continuation file:

```pkl
amends "../local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001/continuation.pkl"
import "../../schema/games/code-localization-helpers.pkl" as game

name = "example-optimize-ic-round-002"

round {
  id = "round-002"

  challenger {
    generate {
      optimizer = game.fakeOptimizer()
      artifactName = "next_challenger_policy.round-002.py"
    }
  }
}
```

This is the release loop SearchBench is growing toward:

```text
previous bundle → continuation.pkl → next round → new challenger → new evidence → decision
```

## Git lifecycle

SearchBench treats the repository lifecycle as part of the evaluation system.

```text
Nix gives you the environment.
Buck gives you the work graph.
Git hooks call Buck gates.
Bundles record the evidence.
```

In practice:

```bash
nix develop
buck2 test //:check
buck2 test //:check_full
```

`git commit` runs the fast gate. `git push` runs the full gate. The full gate covers the Go harness, Iterative Context checks, docs build, and generated Pkl binding freshness.

## Why this matters

Most benchmarks ask:

> Which model is better?

SearchBench also asks:

> Which interface made the same model behave like a better engineer?

For the current code-localization game, the interface might be a code-search backend, MCP server, symbol graph, bounded lookahead policy, workspace provider, validation surface, or configuration shape.

The bet is simple: better interfaces can make agents more reliable, more inspectable, and easier to release.

## Repository map

| Path | Purpose |
| --- | --- |
| `configs/` | Pkl schemas, round manifests, objectives, and dataset slices |
| `src/searchbench-go/` | Go harness: games, rounds, scoring, bundles (Buck-invoked) |
| `src/iterative-context/` | Python MCP/code-search backend used as a challenger interface |
| `docs/` | Product docs, reference docs, and research notes |
| `BUCK` | Root Buck gates: `//:check`, `//:check_full` |

## Read next

Start with [the docs](https://becker63.github.io/searchbench-go/) or [docs/start-here.md](docs/start-here.md).

For the core model, read [docs/concepts.md](docs/concepts.md). For the monorepo map, read [docs/components.md](docs/components.md). For the Nix/Buck/Git workflow, read [docs/development.md](docs/development.md). For the broader thesis, read [docs/research/agent-interface-research.md](docs/research/agent-interface-research.md).

## Module

```text
github.com/becker63/searchbench-go
```
