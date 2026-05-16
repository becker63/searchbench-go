#!/usr/bin/env python3
"""Regenerate agents-need-fewer-commands.md with real unified-diff bodies."""
from __future__ import annotations

from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
PATCH_DIR = Path(__file__).resolve().parent / "full-diffs"
OUT = Path(__file__).resolve().parent / "agents-need-fewer-commands.md"


def p(name: str) -> str:
    return (PATCH_DIR / name).read_text().rstrip("\n")


def fence(body: str) -> str:
    return "```diff\n" + body + "\n```"


def diff_block(path: str, commit: str, body: str) -> str:
    return f"**Diff:** `{path}` · `{commit}`\n\n{fence(body)}"


intro = """# Blog diff candidates: *Agents Need Fewer Commands*

Curated from **verified** `git diff` / `git show` output on this repository (May 2026). Large unrelated churn (lockfiles, `repomix-output.xml`, bulk `src/searchbench-go/` renames) is still omitted from *this* document’s scope notes, but each fenced excerpt below is a **real unified diff** (`diff --git`, index lines, `---`/`+++`, and `@@` hunks) produced by:

`git diff <parent>..<commit> -- <path>`

**Diff line convention:** On the line above each fence, **Diff:** `relative/path` · `commit` names the slice. The fence language tag stays `diff` for syntax highlighting.

**Buck2 landing commit:** `291daa5` — *buck2 start* (2026-05-15). This commit adds root/toolchain Buck config, moves the Go module under `src/searchbench-go/`, wires `buck2.nix`, collapses git-hooks toward **`buck2 test //:check`** and **`buck2 test //:check_full`**, and deletes the old `nix/tools/` command layer plus flake `apps`.

**Submodule caveat:** `git show 291daa5 -- src/iterative-context/BUCK` is empty in the **parent** repository because `src/iterative-context` is a git submodule (gitlink updated in `291daa5`). The root `BUCK` still references `//src/iterative-context:check` — cite root `BUCK` + `docs/architecture/build-system.md`, or `git show` inside the submodule checkout, for the Python-side `BUCK` file body.

---

## Candidate 1 — Root `BUCK`: two semantic gates + Repomix target

**Commit:** `291daa5` — buck2 start
**Files:** `BUCK`
**Theme:** commands → graph; Git hooks → named graph targets (`//:check`, `//:check_full`)
**Suggested insertion point:** `## After: a graph` or `## Buck2 as an agent interface`

**Why it works:**
The entire “what do I run?” answer collapses to two `test_suite` names and one explicit `sh_test`. The comments on each suite document intent (fast vs full) without a README table.

**Diff excerpt:**

"""

c1 = diff_block("BUCK", "291daa5", p("patch-BUCK.diff"))

c2_section = f"""
---

## Candidate 2 — Hooks delegate to Buck2 (flake shrinks the checklist)

**Commit:** `291daa5` — buck2 start
**Files:** `flake.nix`
**Theme:** Git owns lifecycle; Nix owns environment; graph runs substantive checks
**Suggested insertion point:** `## This matters more for agents than for humans` or `## Nix still owns the environment`

**Why it works:**
Shows `repomixThenBuckCheck` / `buckCheckFull` shell apps that **only** wrap `repomix` + `buck2 test`, then replaces a long list of `searchbench-*` pre-push hooks with two entries. Deletes `projectToolPkgs` and flake `apps` — the operational surface moves off “many Nix apps.”

**Diff excerpt:** Full unified diff for `flake.nix` (`git diff 291daa5^..291daa5 -- flake.nix`).

{diff_block("flake.nix", "291daa5", p("patch-flake.nix.diff"))}

**Possible surrounding prose:**
Hooks stop enumerating policy line-by-line; they become **thin triggers** for graph-shaped work (`//:check`, `//:check_full`). Nix still supplies `buck2` and writes `.buckconfig.d/buck2-nix.config` (see Candidate 7).

---

## Candidate 3 — `AGENTS.md`: from `searchbench-*` catalog to two Buck commands

**Commit:** `291daa5` — buck2 start
**Files:** `AGENTS.md`
**Theme:** docs → contracts; fewer operational choices
**Suggested insertion point:** `## Removing knobs is a feature` or `## Before` / `## After` boundary

**Why it works:**
The hook table shrinks to “Repomix + `buck2 test //:check`” vs “`buck2 test //:check_full`”. The giant “Debugging commands” `searchbench-*` table is deleted; Nix + Buck2 section states the flake **does not** ship `nix/tools/` anymore.

**Diff excerpt:** Full unified diff (`git diff 291daa5^..291daa5 -- AGENTS.md`).

{diff_block("AGENTS.md", "291daa5", p("patch-AGENTS.md.diff"))}

**Possible surrounding prose:**
This is the human/agent-visible proof: the “menu” in `AGENTS.md` stops being a catalog of wrappers and becomes **two graph-shaped invocations** plus hygiene.

---

## Candidate 4 — `README.md`: Nix workflow block replaces `searchbench-*` / `nix run .#…`

**Commit:** `291daa5` — buck2 start
**Files:** `README.md`
**Theme:** sanctioned operations; monorepo paths
**Suggested insertion point:** Opening hook or `## After: a graph`

**Why it works:**
Side-by-side deletion of `nix develop -c searchbench-*` and `nix run .#e2e` / `update-repomix` / `publish-issue-wave`, replaced with **`buck2 test //:check`** / **`buck2 test //:check_full`**. Highlights table paths move under `src/searchbench-go/`.

**Diff excerpt:** Full unified diff (`git diff 291daa5^..291daa5 -- README.md`).

{diff_block("README.md", "291daa5", p("patch-README.md.diff"))}

**Possible surrounding prose:**
Use as the “README stopped advertising ten spells” screenshot — same section header, fewer forks.

---

## Candidate 5 — New contract doc: `docs/architecture/build-system.md`

**Commit:** `291daa5` — buck2 start
**Files:** `docs/architecture/build-system.md`
**Theme:** docs explain the graph; Nix vs Buck responsibilities
**Suggested insertion point:** `## Docs changed from instructions to contracts`

**Why it works:**
Explicitly states monorepo layout, that **`nix flake check`** does not run Buck tests, and documents **`buck2 test //:check`** vs **`buck2 test //:check_full`** semantics (including Repomix and `pkl_go_types` race note).

**Diff excerpt:** Full unified diff (`git diff 291daa5^..291daa5 -- docs/architecture/build-system.md`).

{diff_block("docs/architecture/build-system.md", "291daa5", p("patch-build-system.md.diff"))}

**Possible surrounding prose:**
Position as “the checklist moved into architecture prose that names **targets**, not bash one-liners.”

---

## Candidate 6 — Go harness: `src/searchbench-go/BUCK` composes local shims

**Commit:** `291daa5` — buck2 start
**Files:** `src/searchbench-go/BUCK`
**Theme:** Go target vs repo target; separation of `pkl_go_types`
**Suggested insertion point:** `## Buck2 as an agent interface`

**Why it works:**
Shows `go_tests`, `go_cli_build`, and opt-in `pkl_go_types` as `sh_test` nodes composed into `:check` — the “Go local loop vs system loop” story without reading shell bodies.

**Diff excerpt:**

{diff_block("src/searchbench-go/BUCK", "291daa5", p("patch-src-searchbench-go-BUCK.diff"))}

**Possible surrounding prose:**
Contrast with root `//:check`: Go owns **module-scoped** steps; root **composes** Go + Python (+ Repomix only on full).

---

## Candidate 7 — Toolchain cell: `toolchains/BUCK` + `.buckconfig` (Nix packages on the graph)

**Commit:** `291daa5` — buck2 start
**Files:** `toolchains/BUCK`, `.buckconfig`
**Theme:** Nix owns toolchain closure; Buck owns action graph
**Suggested insertion point:** `## Nix still owns the environment` + `## Buck2 as an agent interface`

**Why it works:**
`.buckconfig` declares the `nix` external cell and ignore patterns; `toolchains/BUCK` wires `flake.package` for `python`, `go`, `pkl`, `ruff`, `uv` — concrete “graph execution pulls binaries from the flake-backed toolchain package.”

**Diff excerpt:**

{diff_block(".buckconfig", "291daa5", p("patch-.buckconfig.diff"))}

{diff_block("toolchains/BUCK", "291daa5", p("patch-toolchains-BUCK.diff"))}

**Possible surrounding prose:**
Tie to thesis: Nix is not “where we type `nix run` for every check” — it is **where the Buck `nix` cell pins interpreters and CLIs** the graph uses.

---

## Candidate 8 — Precursor: sandbox flake-check vs module graph (`flakeCheckHooks`)

**Commit:** `cceb6a7` — Remove checked-in Go vendor; split flake-check hooks from Go graph hooks
**Files:** `flake.nix`
**Theme:** layered validation before Buck existed
**Suggested insertion point:** `## After: a graph` (historical contrast: “we already thought in layers”)

**Why it works:**
Still the cleanest pre-Buck artifact for “not every gate runs in every sandbox.” Pairs well with post-`291daa5` docs that say `nix flake check` does not run Buck tests.

**Diff excerpt:** Full unified diff (`git diff cceb6a7^..cceb6a7 -- flake.nix`).

{diff_block("flake.nix", "cceb6a7", p("patch-cceb6a7-flake.nix.diff"))}

**Possible surrounding prose:**
Bridge sentence: layering started in Nix; `291daa5` moves substantive “full repo” meaning to `buck2 test`, while `nix flake check` stays intentionally thin.

---

## Candidate 9 — Bridge: Repomix flags tightened for determinism (pre-Buck hook era)

**Commit:** `811c97c` — nix: add Repomix pre-push freshness check (deterministic snapshot)
**Files:** `nix/tools/core.nix` (file **deleted** in `291daa5`, but the diff is still valid history)
**Theme:** Repomix gate; policy encoded in scripts
**Suggested insertion point:** “Repomix as a gate” sidebar, or evolution footnote

**Why it works:**
Shows `--include-diffs` / `--include-logs` removed in favor of `--no-git-sort-by-changes` — the moral “make the artifact checkable” survived into `repomix_fresh_check.sh` / `//:repomix_fresh_check` in `291daa5`.

**Diff excerpt:** Full unified diff for `nix/tools/core.nix` at that commit (`git diff 811c97c^..811c97c -- nix/tools/core.nix`).

{diff_block("nix/tools/core.nix", "811c97c", p("patch-811c97c-core.nix.diff"))}

**Possible surrounding prose:**
Use as a one-step “we stopped trusting ‘rich’ packs for CI” before the gate became a Buck test target.

---

# Top 3 diffs to include in the blog

1. **`291daa5` — root `BUCK` (`//:check` / `//:check_full`)**
  - **Why it is visually strong:** Entirely greenfield file; two `test_suite` names read like an API.
  - **Where it should go:** `## After: a graph` / `## Buck2 as an agent interface`.
  - **What claim it proves:** The repo exposes **semantic entrypoints**, not a procedure list.
  - **Surrounding prose:** Minimal — let labels + comments carry meaning.
2. **`291daa5` — `flake.nix` hook collapse + `buck2 test` wrappers**
  - **Why it is visually strong:** Large `-` blocks (many `searchbench-*` hooks, `projectToolPkgs`, `apps`) replaced by two hook entries that exec Buck.
  - **Where it should go:** `## This matters more for agents than for humans` or hook lifecycle subsection.
  - **What claim it proves:** Git hooks **delegate**; they no longer encode the whole checklist.
  - **Surrounding prose:** One sentence on thin wrappers vs graph (Candidate 2 body).
3. **`291daa5` — `AGENTS.md` (hook table + command table replacement)**
  - **Why it is visually strong:** Readers see the `searchbench-*` wall disappear and `buck2 test //:check*` appear.
  - **Where it should go:** `## Removing knobs is a feature` or right after the “before” screenshots.
  - **What claim it proves:** Agent-facing docs align with **sanctioned operations** (`buck2 test …`), not rediscovered shell trivia.
  - **Surrounding prose:** Short contrast with earlier `9f8bb7c`-era tables if you show a “before” excerpt elsewhere.

---

## Themes now covered vs still awkward


| Theme                                      | Status after `291daa5`                                                                                                                                                                |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Root `BUCK` / `//:check` / `//:check_full` | **Strong** committed diff (`291daa5`).                                                                                                                                                |
| Buck2 + buck2.nix cell / `.buckconfig.d`   | **Strong** (`291daa5` — `.buckconfig`, `flake.nix` shellHook, `toolchains/`).                                                                                                         |
| Hooks → Buck delegation                    | **Strong** (`291daa5` `flake.nix`).                                                                                                                                                   |
| Monorepo `src/searchbench-go`              | **Strong** as rename stats in `291daa5` (omit giant rename diff in blog; cite `README`/`AGENTS` path edits or `build-system.md`).                                                     |
| `src/iterative-context/BUCK` body          | **Awkward in parent repo** — inspect submodule commit or rely on root `BUCK` + `build-system.md`.                                                                                     |
| “Fewer `nix run` apps”                     | **Strong** — `apps` removed in `291daa5`.                                                                                                                                             |
| Pre-`291daa5` command sprawl               | Still cite `9f8bb7c` / `e5aaedb` separately if you want a explicit “before” screenshot (run `git show 9f8bb7c -- AGENTS.md`). |


---

*Report updated after `291daa5` landed; suitable for curation into `content/posts/ai-cognitive-overload.mdx` (post body not modified here). Patches re-embedded via `docs/blog-diff-candidates/_embed_patches.py`.*
"""

OUT.write_text(intro + c1 + c2_section, encoding="utf-8")
print(f"Wrote {OUT} ({OUT.stat().st_size} bytes)")
