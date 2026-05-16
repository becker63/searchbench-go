# Start here

**Work in progress** — research harness, not a polished platform. **First game:** code localization (bug-localization slices, symbol/code-search with lookahead).

## 1. Round manifest

**File:** `configs/rounds/local-ic-vs-jcodemunch/round.pkl`

**Excerpt:**

```pkl
amends "../../schema/games/code-localization.pkl"
import "../../schema/games/code-localization-helpers.pkl" as game

name = "example-local-ic-vs-jcodemunch-round-001"

round = (game.defineFromScratch("round-001")) {
  incumbent = game.jcodemunch()
  challenger = (game.iterativeContext("policies/challenger_policy.py")) {
    selectionPolicy { id = "challenger-policy-round-001" }
  }
  matches = game.lca("py", "dev", 5)
  evaluator = game.fakeEvaluator()
  scoring = game.objective("scoring/localization-objective.pkl")
}
```

| Stanza | Meaning | Example path |
| --- | --- | --- |
| `incumbent` | Baseline interface | jCodeMunch backend |
| `challenger` | Interface under test | IC + `policies/challenger_policy.py` |
| `matches` | Dataset slice (5 LCA rows) | `game.lca("py", "dev", 5)` |
| `evaluator` | Match runner (fake = offline) | No live models |
| `scoring` | Pkl objective module | `scoring/localization-objective.pkl` |

**Schema:** `configs/schema/games/code-localization.pkl` · **Helpers:** `configs/schema/games/code-localization-helpers.pkl`

## 2. Run it

From repo root (after [building the CLI](../README.md#run-one-local-round)):

```bash
./searchbench run \
  --manifest=configs/rounds/local-ic-vs-jcodemunch/round.pkl \
  --bundle-root="$(pwd)/.tmp-artifacts"
```

**Proves with:** same flow as `buck2 test //src/searchbench-go:check` (Go tests exercise manifests).

Offline only today — `game.fakeEvaluator()`, no live MCP or provider calls.

## 3. Bundle output

**Path:** `{bundle-root}/games/code-localization/rounds/round-001/`

Checked-in reference copy:

`configs/rounds/local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001/`

```text
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

| File | Inspect for |
| --- | --- |
| `decision.json` | `PROMOTE_CHALLENGER` / `REVIEW` / `REJECT` |
| `objective.json` | Scored values from the Pkl objective |
| `evidence.pkl` | Round evidence fed into scoring |
| `continuation.pkl` | Input for the **next** round manifest |

Details: [reference/bundles.md](./reference/bundles.md).

## 4. Next round (continuation)

**File:** `configs/rounds/optimize-ic/round.pkl`

**Excerpt:**

```pkl
amends "../local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001/continuation.pkl"
import "../../schema/games/code-localization-helpers.pkl" as game

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

**Meaning:** amends prior bundle continuation; optimizer proposes a new challenger policy for round 002. See [reference/pkl-rounds.md](./reference/pkl-rounds.md).

## 5. Read next

1. [concepts.md](./concepts.md) — vocabulary → files
2. [reference/pkl-rounds.md](./reference/pkl-rounds.md) — manifest API
3. [reference/bundles.md](./reference/bundles.md) — bundle fields
4. [components.md](./components.md) — who owns what
5. [development.md](./development.md) — `buck2 test //:check_full`

Docs home: [index.md](./index.md). Contributors: [AGENTS.md](../AGENTS.md).
