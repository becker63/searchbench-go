# Components

SearchBench is a small monorepo. Each top-level tree is a **component** with concrete files and Buck proof targets.

## Overview

| Component | Path | Example files | Proof |
| --- | --- | --- | --- |
| SearchBench-Go | `src/searchbench-go/` | `internal/app/round`, `internal/adapters/config/pkl`, `internal/adapters/bundle/fs` | `buck2 test //src/searchbench-go:check` |
| Iterative Context | `src/iterative-context/` | `src/iterative_context/server.py`, `optimizable_backend.json` | `buck2 test //src/iterative-context:check_full` |
| Configs | `configs/` | `rounds/local-ic-vs-jcodemunch/round.pkl`, `schema/SearchBenchRound.pkl` | `buck2 test //:check_full` |
| Docs | `docs/` | `index.md`, `start-here.md`, `BUCK` | `buck2 test //docs:check` |

**Component vs package:** this page = repo projects; [reference/package-boundaries.md](./reference/package-boundaries.md) = Go import layers inside the harness.

---

## SearchBench-Go

**Path:** `src/searchbench-go/`

**Owns:** Game/Round/Match model; round execution; bundle writing; Pkl loading; CLI (`searchbench`).

**Does not own:** IC Python internals; visualization UI; live provider calls in CI gates.

| Area | Example path |
| --- | --- |
| Round orchestration | `internal/app/round/` |
| Pkl config | `internal/adapters/config/pkl/` |
| Bundles | `internal/adapters/bundle/fs/` |
| Objectives | `internal/adapters/scoring/pkl/` |
| Workspaces | `internal/adapters/workspace/` |
| Optimizer validation | `internal/agents/optimizer/policy/` |
| CLI | `cmd/searchbench/` |

**Proves with:** `buck2 test //src/searchbench-go:check` · Pkl regen: `buck2 build //src/searchbench-go:pkl_go_types`

---

## Iterative Context

**Path:** `src/iterative-context/` (submodule)

**Owns:** MCP server; code-search / lookahead; `validate_policy`; `optimizable_backend.json`.

**Does not own:** SearchBench scoring, bundle layout, public round schema.

| Area | Example path |
| --- | --- |
| MCP server | `src/iterative_context/server.py` |
| Policy validation | `src/iterative_context/validate_policy.py` |
| Backend descriptor | `optimizable_backend.json` |
| Example policy | `configs/rounds/local-ic-vs-jcodemunch/policies/challenger_policy.py` |

**Role:** Primary **challenger** for code-localization; materialized into a [candidate workspace](./candidate-workspaces.md) before validate + launch.

**Proves with:** `buck2 test //src/iterative-context:check` · `buck2 test //src/iterative-context:check_full`

---

## Visualization

**Status:** planned — not in this repo.

**Expected role:** Static bundle replay; incumbent vs challenger inspection. Consumes bundles only; does not own scoring.

---

## Configs

**Path:** `configs/`

**Owns:** Schema and round manifests; objectives; checked-in example bundles.

| Area | Example path |
| --- | --- |
| Round schema | `schema/SearchBenchRound.pkl` |
| Game | `schema/games/code-localization.pkl` |
| From-scratch round | `rounds/local-ic-vs-jcodemunch/round.pkl` |
| Continuation round | `rounds/optimize-ic/round.pkl` |
| Objective | `rounds/local-ic-vs-jcodemunch/scoring/localization-objective.pkl` |
| Example bundle | `rounds/local-ic-vs-jcodemunch/artifacts/.../round-001/` |

**Proves with:** `buck2 test //:check_full` (includes harness + schema binding checks).

---

## Docs

**Path:** `docs/`

**Owns:** Product spine, reference, research. Hosted: [becker63.github.io/searchbench-go](https://becker63.github.io/searchbench-go/).

| Area | Example path |
| --- | --- |
| Entry | `index.md`, `start-here.md` |
| Reference | `reference/pkl-rounds.md`, `reference/bundles.md` |
| Buck target | `docs/BUCK` → `//docs:check` |

**Proves with:** `buck2 test //docs:check` · `buck2 build //docs:site`

---

## Root gates (Nix / Buck)

**Paths:** `flake.nix`, `BUCK`, `toolchains/`

**Owns:** `nix develop`; git-hooks → `buck2 test //:check` (commit), `buck2 test //:check_full` (push).

See [development.md](./development.md).
