# SearchBench-Go

<p align="center">
  <strong>Test the interfaces your coding agents use.</strong>
</p>

<p align="center">
  <strong>Pkl manifests</strong> → <strong>SearchBench CLI</strong> → <strong>evidence bundles</strong> → <strong>release decisions</strong>
</p>

<p align="center">
  <a href="https://becker63.github.io/searchbench-go/">📚 Docs</a>
  ·
  <a href="docs/start-here.md">🚀 Start here</a>
  ·
  <a href="docs/concepts.md">🧭 Concepts</a>
  ·
  <a href="docs/components.md">🧩 Components</a>
  ·
  <a href="docs/development.md">🛠️ Development</a>
  ·
  <a href="docs/research/agent-interface-research.md">🔬 Research</a>
</p>

---

SearchBench is a **work-in-progress harness** for evaluating **agent-facing interfaces** over benchmark tasks. It is not a stable public API yet, but its current user-facing surface is already taking shape: write a **Pkl round manifest**, run it with the SearchBench CLI, and inspect the resulting **evidence bundle**.

SearchBench wraps task families as **games**, then asks:

> Which interface makes the same agent perform better?

The first stress-test game is **code localization**. SearchBench uses bug-localization dataset slices to test whether symbol/code-search tools with **lookahead** help an agent find the files that need to change. The longer-term goal is for the same `Game` model to wrap other task families, such as SWE-bench-style issue resolution, proof-target selection, and docs/context selection.

```text
Game → Round → Match → Evidence → Objective → Decision → NextChallenger
```

## The current API

Today, the cleanest API is not a Go package. It is the round manifest.

```text
configs/rounds/*/round.pkl
  declares what is being tested

searchbench run
  executes the round

artifacts/games/<game-id>/rounds/<round-id>/
  records what happened
```

Pkl owns the typed, readable manifest surface: game defaults, round shape, local constraints, and composition through `amends`. Go owns execution: match identity, backend semantics, validation, scoring models, reports, decisions, and bundle writing.

A from-scratch round currently looks like this:

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

That is the core shape of SearchBench: a game, a dataset slice, an incumbent interface, a challenger interface, an evaluator, and an objective.

The objective is also visible Pkl. The round manifest points to it; it does not bury scoring math inside the runner.

```pkl
amends "../../../schema/SearchBenchObjective.pkl"
import "../../../schema/SearchBenchObjectiveHelpers.pkl" as helpers

local maxHop = 12.0
local tokenBudget = 20000.0

objectiveId = "localization-v1"

local challengerQuality =
  1.0 - (if (current.localizationDistance.goldHop.challenger < maxHop) current.localizationDistance.goldHop.challenger else maxHop) / maxHop

local tokenEfficiency =
  1.0 - (if (current.challengerUsage.totalTokens < tokenBudget) current.challengerUsage.totalTokens else tokenBudget) / tokenBudget

local base = (challengerQuality * 0.8) + (tokenEfficiency * 0.2)

local regressionPenalty =
  if (current.regressions.severeCount > 0) 0.0 else 1.0

local invalidPredictionPenalty =
  if (current.invalidPredictions.known && current.invalidPredictions.count > 0) 0.0 else 1.0

local finalScore = base * regressionPenalty * invalidPredictionPenalty

values = new {
  helpers.intermediate("challengerQuality", challengerQuality)
  helpers.intermediate("tokenEfficiency", tokenEfficiency)
  helpers.intermediate("base", base)
  helpers.penalty("regressionPenalty", regressionPenalty)
  helpers.penalty("invalidPredictionPenalty", invalidPredictionPenalty)
  helpers.finalValue("final", finalScore)
}

final = "final"
```

This is the release-engineering shape of the project. The Pkl says what is being evaluated. The CLI runs it. The bundle records what happened. The Git lifecycle proves the repo state around it.

## Run a local round

```bash
git clone https://github.com/becker63/searchbench-go.git
cd searchbench-go

cd src/searchbench-go
go build -o ../../searchbench ./cmd/searchbench
cd ../..

./searchbench run \
  --manifest=configs/rounds/local-ic-vs-jcodemunch/round.pkl \
  --bundle-root="$(pwd)/.tmp-artifacts"
```

The checked-in local examples use the fake/offline path. They are meant to prove manifest resolution, round execution, bundle writing, evidence generation, objective evaluation, and decisions without requiring live MCP servers or model API keys.

## What a bundle contains

A completed round writes a durable bundle under the bundle root.

```text
artifacts/games/code-localization/rounds/round-001/
  COMPLETE
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

The bundle is the source of truth for review. `resolved-round.json` captures the executable round after Pkl resolution. `evidence.pkl` is the scoring-facing evidence document. `objective.json` records the named objective values and final score. `decision.json` records whether the challenger should be promoted, reviewed, or rejected. `metadata.json` inventories files and content hashes.

A real bundle evidence file looks like this in shape:

```pkl
schemaVersion = "searchbench.round_evidence.v1"
gameId = "code-localization"
roundId = "round-001"
reportId = "report-round-001"

policies {
  incumbent {
    id = "jcodemunch"
    name = "jCodeMunch"
    backend = "jcodemunch"
  }

  challenger {
    id = "iterative-context"
    name = "Iterative Context"
    backend = "iterative-context"
    policy {
      id = "challenger-policy-round-001"
      language = "python"
      entrypoint = "score_fn"
    }
  }
}

matchCounts {
  total = 5
}

localizationDistance {
  goldHop {
    metric = "gold_hop"
    direction = "lower_is_better"
    incumbent = 0
    challenger = 0
    delta = 0
    improved = false
    regressed = false
  }
}

challengerUsage {
  available = false
  measuredRuns = 0
  totalTokens = 0
}

decision {
  decision = "REVIEW"
  reason = "challenger did not improve the composite score in local fake comparison"
}
```

That evidence is intentionally inspectable. SearchBench should not be a black box that only emits a leaderboard number.

## Continuation and releases

Every completed round can write continuation artifacts. `continuation.json` is the machine-readable authority. `continuation.pkl` is the human-editable lineage surface.

A later round can amend the previous bundle’s `continuation.pkl` and only patch what changed:

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

That is the core release loop SearchBench is growing toward:

```text
previous bundle
  → continuation.pkl
  → next round manifest
  → generated or supplied challenger
  → new evidence bundle
  → decision
```

The meta-harness can eventually propose changes, but SearchBench owns the round, evidence, and judgment.

## Git lifecycle

SearchBench uses the repository lifecycle as part of the product surface.

Nix supplies the toolchain and installs Git hooks. Buck names repo operations. Git hooks call Buck gates.

```text
nix develop
  → installs the dev shell and Git hooks

git commit
  → buck2 build //tooling:repomix
  → buck2 test //:check

git push
  → buck2 test //:check_full
```

The fast gate proves the basic repo state. The full gate includes the Go harness, Iterative Context checks, generated Pkl binding freshness, docs build, and Repomix freshness.

```bash
nix develop -c buck2 test //:check
nix develop -c buck2 test //:check_full
```

The important split is:

```text
Pkl declares the round.
Go executes the lifecycle.
Buck proves repo operations.
Git records the change.
Bundles record the evidence.
```

## Why SearchBench exists

Most benchmarks ask:

> Which model is better?

SearchBench also asks:

> Which interface made the same model behave like a better engineer?

For the current code-localization game, an interface might be a code-search backend, MCP server, symbol graph, bounded lookahead policy, validation surface, workspace provider, or configuration shape. The novel experiment is to move more code search into deterministic computation so the agent can consume a better interface instead of manually performing every search step through tool calls.

SearchBench is early, but the direction is concrete: wrap benchmark tasks as games, vary the interface, run controlled rounds, and keep the evidence.

## Repository map

| Path | Purpose |
| --- | --- |
| `configs/` | Pkl schemas, round manifests, objectives, and dataset slices |
| `src/searchbench-go/` | Go harness: games, rounds, scoring, bundles, CLI |
| `src/iterative-context/` | Python MCP/code-search backend used as a challenger interface |
| `docs/` | Public docs, reference docs, and research notes |
| `tooling/` | Repo lifecycle tools such as Repomix |
| `BUCK` | Root Buck gates: `//:check`, `//:check_full` |

More: [docs/components.md](docs/components.md)

## Read next

Start with [docs/index.md](docs/index.md) or the hosted docs at [becker63.github.io/searchbench-go](https://becker63.github.io/searchbench-go/).

The concepts doc explains `Game`, `Round`, `Match`, `Interface`, `Evidence`, and `Bundle`. The development doc explains the Nix/Buck/Git workflow. The research note explains the broader claim: the model is not the whole system; the interface changes what the agent is capable of.

## Module

```text
github.com/becker63/searchbench-go
```
