# Components

SearchBench is a small monorepo. Top-level source trees are **separate components** because they play different roles in the evaluation loop.

The **harness** owns games, rounds, evidence, scoring, and bundles. **Backends** expose agent-facing interfaces. **Visualization** (planned) will project bundles into an inspection UI — not own scoring or bundle semantics.

## Overview

| Component | Path | Role | Language | Main proof |
| --- | --- | --- | --- | --- |
| SearchBench-Go | `src/searchbench-go/` | Harness, round execution, bundles, config adapters | Go | `buck2 test //src/searchbench-go:check` |
| Iterative Context | `src/iterative-context/` | MCP / code-search backend and IC policy surface | Python | `buck2 test //src/iterative-context:check_full` |
| Visualization | *not in repo yet* | Bundle/trace replay and evidence inspection | TBD | TBD |
| Configs | `configs/` | Pkl schema and round/objective manifests | Pkl | `buck2 test //:check_full` |
| Docs | `docs/` | Product, contributor, reference, research docs | Markdown | `buck2 test //docs:check` |

## Components vs package boundaries

| Level | Question | Doc |
| --- | --- | --- |
| **Component** | What are the major projects in this repo? | This page |
| **Package** | How is the Go harness layered internally? | [reference/package-boundaries.md](./reference/package-boundaries.md) |

---

## SearchBench-Go

**Path:** `src/searchbench-go/`

**Owns:** Game/Round/Match/Evidence/Decision model; round execution; evaluator and optimizer orchestration; workspace seed provider wiring; bundle writing; Pkl config loading; CLI (`searchbench`).

**Must not own:** External meta-harness worktree orchestration; Iterative Context Python internals; visualization UI; live provider calls in deterministic CI gates.

**Key docs:** [architecture.md](./architecture.md), [reference/package-boundaries.md](./reference/package-boundaries.md), [candidate-workspaces.md](./candidate-workspaces.md), [development.md](./development.md).

**Validation:** `buck2 test //src/searchbench-go:check` · Pkl regen: `buck2 build //src/searchbench-go:pkl_go_types`

---

## Iterative Context

**Path:** `src/iterative-context/` (git submodule)

**Owns:** MCP server; code-search and graph-lookahead behavior; policy validation (`validate_policy`); `install_score` / admin contract; `optimizable_backend.json` and Buck descriptor target.

**Role in SearchBench:** Primary **challenger interface** for the code-localization game. The harness materializes a copy into an isolated **candidate workspace**, validates proposals there, then launches MCP from that same tree.

**Must not own:** SearchBench scoring or decisions; bundle layout; public round manifest schema.

**Key docs:** [candidate-workspaces.md](./candidate-workspaces.md), [development.md](./development.md).

**Validation:** `buck2 test //src/iterative-context:check` · `buck2 test //src/iterative-context:check_full`

---

## Visualization

**Status:** planned — **not checked in** to this repository yet.

**Expected role:** Static bundle replay; trace/event views; incumbent vs challenger comparison; evidence inspection; optional release-decision UI.

**Expected path:** TBD (e.g. `src/searchbench-visualization/`).

**Must not own:** Source-of-truth scoring; bundle semantics; runtime mutation; provider validation pipelines.

**Integration:** Consumes **static bundles** and projected JSON from the harness. A projection layer, not the system of record.

---

## Configs

**Path:** `configs/`

**Owns:** `configs/schema/SearchBenchRound.pkl`; example round manifests under `configs/rounds/`; objective modules referenced by manifests.

**Role:** Declares game/round **intent**. Execution stays in SearchBench-Go.

**Validation:** `buck2 test //:check_full`; regenerate Go bindings after schema edits ([development.md](./development.md)).

---

## Docs

**Path:** `docs/`

**Owns:** Public product spine, contributor workflow, narrow reference docs, research notes.

**Role:** Onboarding for humans and agents. Hosted at [becker63.github.io/searchbench-go](https://becker63.github.io/searchbench-go/).

**Validation:** `buck2 test //docs:check` · artifact: `buck2 build //docs:site`

---

## Root tooling (Nix / Buck)

**Paths:** `flake.nix`, `BUCK`, `toolchains/`, `package.json` (VitePress)

**Owns:** `nix develop` dev shell; **git-hooks.nix** hook installation; Buck gates (`//:check`, `//:check_full`, `//docs:check`).

**Split:**

- **Nix** — toolchain and Git hook lifecycle (`nix develop` only; no separate hook installer).
- **Buck** — repo operation graph contributors run via hooks or manually.
- **SearchBench-Go** — evaluation lifecycle and product semantics.

See [development.md](./development.md).
