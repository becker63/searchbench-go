# Blog diff candidates: *Agents Need Fewer Commands*

Curated from **verified** `git show` output on this repository (May 2026). Trims omit lockfiles, `repomix-output.xml`, and vendor blobs.

**History gaps (at audit time):** `git log -- BUCK` is empty — root `BUCK` / `//:check` / `//:check_full` exist only in the working tree, not as committed history. Likewise `git log -S'buck2'`, `toolchains/BUCK`, `.buckconfig`*, `docs/architecture/build-system.md`, and `src/searchbench-go/**` / `src/iterative-context/BUCK` have **no** useful commit history for the Buck2/monorepo-interface story; use the live tree or future commits for those excerpts.

---

## Candidate 1 — Flake becomes the automation spine (origin story)

**Commit:** `415c153` — Add Nix dev shell, git-hooks pre-commit, vendor, and Repomix workflow
**Files:** `flake.nix`
**Theme:** Nix owns environment + lifecycle hooks; commands → many named `searchbench-`* tools (the “before” pile begins)
**Suggested insertion point:** after `## Before: a list of procedures` in `content/posts/ai-cognitive-overload.mdx`

**Why it works:**
One diff shows the flake morphing from a one-line “Go development shell” into git-hooks wiring, `commonHooks`, and imports of `./nix/dev-tools.nix`. It is the concrete birth of “policy lives in Nix + hooks,” which the post later argues should collapse further into a Buck graph.

**Diff excerpt:**

```diff
--- a/flake.nix
+++ b/flake.nix
@@ -1,38 +1,80 @@
 {
-  description = "Go development shell";
+  description = "SearchBench-Go — Nix dev shell, pre-commit, and CI checks";

   inputs = {
     nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
     flake-utils.url = "github:numtide/flake-utils";
+    git-hooks.url = "github:cachix/git-hooks.nix";
+    git-hooks.inputs.nixpkgs.follows = "nixpkgs";
   };

   outputs =
-    { nixpkgs, flake-utils, ... }:
+    {
+      nixpkgs,
+      flake-utils,
+      git-hooks,
+      ...
+    }:
     flake-utils.lib.eachDefaultSystem (
       system:
       let
-        pkgs = import nixpkgs {
-          inherit system;
+        pkgs = nixpkgs.legacyPackages.${system};
+        tools = import ./nix/dev-tools.nix { inherit pkgs; };
+
+        commonHooks = {
+          gofmt = {
+            enable = true;
+            excludes = [ "^vendor/" ];
+          };
+          govet = {
+            enable = true;
+            extraPackages = [ pkgs.go ];
+            excludes = [ "^vendor/" ];
+          };
+          golangci-lint = {
+            enable = true;
+            extraPackages = [ pkgs.go ];
+            excludes = [ "^vendor/" ];
+          };
+          …
```

**Possible surrounding prose:**
Use this as the “once upon a time we centralized chaos” frame: reproducible, yes, but the interface is still an explosion of hook entries and flake-defined tools—the opposite of a single semantic `//:check`.

---

## Candidate 2 — Tooling leaves `flake.nix` bodies (module boundary)

**Commit:** `b6e66af` — nix: vendor under nix/, split dev-tools into nix/tools
**Files:** `flake.nix`
**Theme:** Nix owns environment (structure), not yet “fewer commands” but cleaner closure layout
**Suggested insertion point:** `## Nix still owns the environment`

**Why it works:**
Shows `import ./nix/dev-tools.nix` → `import ./nix/tools { inherit pkgs; }` and tightens vendor exclude lists. Supports “Nix stays the environment and packaging cell; the flake file stops being the entire program.”

**Diff excerpt:**

```diff
--- a/flake.nix
+++ b/flake.nix
@@ -14,76 +14,85 @@
       flake-utils,
       git-hooks,
       …
     }:
     flake-utils.lib.eachDefaultSystem (
       system:
       let
         pkgs = nixpkgs.legacyPackages.${system};
-        tools = import ./nix/dev-tools.nix { inherit pkgs; };
+        tools = import ./nix/tools { inherit pkgs; };
+
+        # Modules live under nix/vendor/; root vendor/ is a symlink for Go.
+        vendorExcludes = [
+          "^nix/vendor/"
+          "^vendor/"
+        ];

         commonHooks = {
           gofmt = {
             enable = true;
-            excludes = [ "^vendor/" ];
+            excludes = vendorExcludes;
           };
           govet = {
             enable = true;
             extraPackages = [ pkgs.go ];
-            excludes = [ "^vendor/" ];
+            excludes = vendorExcludes;
           };
           …
```

**Possible surrounding prose:**
A stepping-stone: the repo admits tooling is a *system* with structure (`nix/tools/`), foreshadowing that the *action graph* might live elsewhere (Buck) while Nix keeps the toolchain closure.

---

## Candidate 3 — `nix flake check` vs “full module graph” (explicit split)

**Commit:** `cceb6a7` — Remove checked-in Go vendor; split flake-check hooks from Go graph hooks
**Files:** `flake.nix`
**Theme:** Git/lifecycle + Nix sandbox semantics; commands → named hook groups (precursor to “graph layers”)
**Suggested insertion point:** `## After: a graph` (as analog) or `## Nix still owns the environment`

**Why it works:**
Visually splits `flakeCheckHooks` (sandbox-safe) from `goModuleGraphHooks` (needs network/module cache), and wires `checks.pre-commit-check` to `**flakeCheckHooks` only**. That is a crisp “two speeds” story even before Buck2: not every command belongs in every context.

**Diff excerpt:**

```diff
--- a/flake.nix
+++ b/flake.nix
@@ -32,55 +32,50 @@
-        commonHooks = {
-          gofmt = { … };
-          govet = { … extraPackages = [ pkgs.go ]; … };
-          golangci-lint = { … };
+        # Runs in `nix flake check` (sandbox, no network). Go linters, `go test`, etc. need the
+        # module proxy or a local module cache, so those hooks live only in `preCommitDev` / pre-push.
+        flakeCheckHooks = {
+          gofmt.enable = true;
+          nixfmt-rfc-style.enable = true;
+          deadnix.enable = true;
+          statix.enable = true;
+          shellcheck.enable = true;
+          shfmt.enable = true;
           …
-          searchbench-architecture = { … };
           searchbench-vocabulary = { … };
         };

+        goModuleGraphHooks = {
+          govet = { enable = true; extraPackages = [ pkgs.go ]; };
+          golangci-lint = { enable = true; extraPackages = [ pkgs.go ]; };
+          searchbench-architecture = { … };
+        };
+
+        commonHooks = flakeCheckHooks // goModuleGraphHooks;
+
         …
         preCommitCheck = git-hooks.lib.${system}.run {
           src = ./.;
-          hooks = commonHooks;
+          hooks = flakeCheckHooks;
         };
```

**Possible surrounding prose:**
This is the firewall-style “policy encoded in structure” moment: the sandboxed check is not a weaker opinion—it is a *different* slice of the same repository, explicitly named.

---

## Candidate 4 — Pre-push becomes a chorus of `searchbench-`* wrappers

**Commit:** `e5aaedb` — Add staticcheck/golangci gates, Nix tooling, importcheck; migrate scripts
**Files:** `flake.nix`, `AGENTS.md`
**Theme:** Commands → many operational choices; Git hooks own lifecycle timing
**Suggested insertion point:** `## Before: a list of procedures` or `## Removing knobs is a feature` (as the thing you later delete)

**Why it works:**
The flake gains discrete pre-push hooks (`searchbench-staticcheck-push`, `searchbench-golangci-push`, …) and AGENTS gains “Quality gate tiers” plus a “Handy commands” table. It is the peak “agent must read a catalog” shape—ideal contrast for a post arguing for `//:check`-style semantics.

**Diff excerpt (AGENTS.md):**

```diff
--- a/AGENTS.md
+++ b/AGENTS.md
@@ -54,13 +54,36 @@
 …
-**Staticcheck:** … run `staticcheck ./...` locally …
+**`staticcheck` binary:** … Run `nix develop -c staticcheck ./...` or `nix develop -c searchbench-staticcheck`.
+
+**Go / lint policy:** `.golangci.yml` enables high-signal checks …
+
+**Quality gate tiers:**
+
+| Tier | What runs |
+| --- | --- |
+| `nix flake check` | Sandboxed: `gofmt`, Nix … **no** full Go module graph |
+| `nix develop` + pre-commit | Full dev hook set: … |
+| `git push` (pre-push) | `go test ./...`, root e2e, `searchbench-check-generated`, … `searchbench-staticcheck`, `searchbench-golangci` |
+| `nix develop -c searchbench-agent-merge-check` | Strictest local gate: … |
+
+**Handy commands:**
+
+| Command | Purpose |
+| --- | --- |
+| `nix develop -c searchbench-staticcheck` | `staticcheck ./...` |
+| `nix develop -c searchbench-golangci` | `golangci-lint run ./...` |
+| `nix develop -c searchbench-go-mod-tidy-check` | Fail if `go mod tidy` would change … |
+| `nix develop -c searchbench-prompt-contract-check` | Tests for `.templ` XML prompt contracts |
+| `nix develop -c searchbench-refresh-pkl-example-fixtures` | … |
+| `nix develop -c searchbench-openai-netwatch` | … |
+| `nix develop -c searchbench-go-build-root` | `go build -o searchbench ./cmd/searchbench` |
```

**Diff excerpt (flake.nix pre-push hooks):**

```diff
         prePushHooks = {
           go-test-all = { … };
-          nix-flake-check-push = { … };
+          searchbench-e2e-push = { … };

-          searchbench-e2e-push = { … };
+          searchbench-check-generated-push = { … };
+
+          searchbench-go-mod-tidy-check-push = { … };
+
+          searchbench-staticcheck-push = { … };
+
+          searchbench-golangci-push = { … };
         };
```

**Possible surrounding prose:**
Every row is a *knob* with a name. The blog’s punchline writes itself: agents do not need more spelling lessons—they need fewer spelled-out entrypoints.

---

## Candidate 5 — AGENTS reframes validation as Git lifecycle (not `go test ./...`)

**Commit:** `9f8bb7c` — precommit
**Files:** `AGENTS.md`
**Theme:** Git hooks own lifecycle timing; docs → contracts
**Suggested insertion point:** `## This matters more for agents than for humans` or opening hook (contrast with later Buck one-liners)

**Why it works:**
Replaces “Run `go test ./...` before handing off” with an explicit **pre-commit / pre-push** contract and tables listing what each stage runs—including many `searchbench-`* debug commands. Strong “documentation as operational spec” artifact; also shows why you want *fewer* spelled procedures next.

**Diff excerpt:**

```diff
 ## Validation

-Run `go test ./...` before handing off code changes. For schema changes, regenerate Pkl bindings with:
+The routine gate is Git-driven: use `nix develop`, then rely on **`git commit`** and **`git push`** to run hooks.
+
+For schema changes, regenerate Pkl bindings with:
 …
 ## Nix development (preferred)

 …
-Use the flake for … the `searchbench-`* helper commands …
+Use the flake for … `searchbench-*` helpers …
+
+*`*nix develop`** installs Git hooks …
+
+| Stage | What runs |
+| --- | --- |
+| **`git commit` (pre-commit)** | … **golangci-lint** … **govet**, architecture + prompt contract tests, … **Repomix** snapshot … |
+| **`git push` (pre-push)** | **`go test ./...`**, root **e2e**, **searchbench-check-generated**, … **`nix flake check`** |
+
+Hook staging avoids duplicate **staticcheck** …
 …
-**Quality gate tiers:** … (tier table)
+**Orchestration outside this repo:** … external meta harness …
+**Debugging commands** (use only when a hook failed …):
+
+| Command | Purpose |
+| --- | --- |
+| `nix develop -c searchbench-staticcheck` | … |
+| `nix develop -c searchbench-golangci` | … |
+| … | … |
```

**Possible surrounding prose:**
This is the “hooks stopped being a mystery bash script and became a documented state machine” moment—good bridge to “next step: stop encoding the state machine in prose at all; expose `//:check`.”

---

## Candidate 6 — Repomix: deterministic regen + pre-push freshness gate

**Commit:** `811c97c` — nix: add Repomix pre-push freshness check (deterministic snapshot)
**Files:** `nix/tools/core.nix`, `flake.nix`, `AGENTS.md`, `docs/engineering/agentic-development-flow.md`
**Theme:** Repomix gate; hooks delegate policy; docs clarify contract
**Suggested insertion point:** `## Removing knobs is a feature` or a new subsection on “artifacts as checks”

**Why it works:**
`searchbench-update-repomix` drops `--include-diffs` / `--include-logs` for reproducibility; new `searchbench-repomix-fresh-check` encodes “regen and fail if not committed”; flake adds `searchbench-repomix-fresh-check-push` on **pre-push**. AGENTS and `agentic-development-flow.md` spell the contract. This is “memory → enforced target” without Buck, but same moral as graph targets.

**Diff excerpt (core.nix + flake hook):**

```diff
   searchbench-update-repomix = mkInRepo {
     name = "searchbench-update-repomix";
     text = ''
       repomix \
         --output repomix-output.xml \
         --style xml \
         --compress \
-        --include-diffs \
-        --include-logs \
-        --include-logs-count 10
+        --no-git-sort-by-changes
       git add repomix-output.xml
     '';
     runtimeInputs = [ pkgs.repomix ];
   };

+  searchbench-repomix-fresh-check = mkInRepo {
+    name = "searchbench-repomix-fresh-check";
+    text = ''
+      ${searchbench-update-repomix}/bin/searchbench-update-repomix
+      …
+      if ! git diff --quiet -- repomix-output.xml || ! git diff --quiet --cached -- repomix-output.xml; then
+        echo "searchbench-repomix-fresh-check: repomix-output.xml is not committed at HEAD after regeneration." >&2
+        …
+        exit 1
+      fi
+    '';
+    runtimeInputs = [ pkgs.gnugrep ];
+  };
```

```diff
           searchbench-nix-flake-check-push = { … };
+
+          searchbench-repomix-fresh-check-push = {
+            enable = true;
+            name = "Repomix snapshot fresh (pre-push)";
+            entry = "${tools.searchbench-repomix-fresh-check}/bin/searchbench-repomix-fresh-check";
+            pass_filenames = false;
+            stages = [ "pre-push" ];
+          };
```

**Possible surrounding prose:**
Use to argue agents should not “remember Repomix”—they should encounter a **named failure mode** tied to lifecycle (push), like encountering a red `//:check` node.

---

## Candidate 7 — Python IC validation as an explicit staged pipeline (not “just Go”)

**Commit:** `b623612` — IC optimizer validation pipeline and iterative-context validate_policy
**Files:** `docs/engineering/optimizer-policy-validation.md` (new)
**Theme:** Monorepo honesty; Python target semantics (doc-level; no `src/iterative-context/BUCK` in history)
**Suggested insertion point:** `## The monorepo shape` or `## Buck2 as an agent interface` (as “what a dedicated target *means*”)

**Why it works:**
Lists ordered steps (`py_compile`, `ic_validate_policy`, `basedpyright`, `ruff_check`, `pytest`) with explicit “no `sh -c` / no raw binaries / argv allowlist” constraints. Supports “second runtime gets its own contract,” which parallels “Python gets its own Buck target” once that lands in commits.

**Diff excerpt:**

```diff
+## Default pipeline steps (IC)
+…
+1. **`stage_policy`** — …
+2. **`policy_static_precheck`** — `python -m py_compile` …
+3. **`ic_validate_policy`** — `uv run python -m iterative_context.validate_policy …` from the **`iterative-context`** submodule root.
+4. **`basedpyright`** — `uv run basedpyright`
+5. **`ruff_check`** — `uv run ruff check`
+6. **`pytest`** — `uv run pytest`
+
+Commands are executed without `sh -c`, without raw `pytest`/`pyright` binaries on PATH, and without `PYTHONPATH` hacks. Dynamic argv is constrained by `execpipeline.ICOptimizerAllowlist`.
```

**Possible surrounding prose:**
Even before Buck exposes `//iterative-context:check`, this diff shows the *intent*: validation is a composable pipeline with named stages, not “hope the agent runs the right uv incantation.”

---

## Candidate 8 — AGENTS “Start Here” becomes a docs hub (instructions → map)

**Commit:** `f08a09b` — Refactor agents vertical slices; relocate docs; add root README
**Files:** `AGENTS.md`
**Theme:** Docs changed from loose filenames to `docs/` hub; boundaries for agents
**Suggested insertion point:** Opening hook or `## Docs changed from instructions to contracts`

**Why it works:**
Small, readable diff: pointers switch from root-level `architecture.md` to `docs/README.md` and architecture paths; boundaries split `internal/agents` from adapters. Shows early “make the repo navigable as a system” work—complements the later “fewer commands” thesis (structure over tribal knowledge).

**Diff excerpt:**

```diff
 ## Start Here

 Read these files first when working in the repo:

-- `architecture.md`
-- `visualization.md`
+- `docs/README.md` (documentation index)
+- `docs/architecture/architecture.md`
+- `docs/architecture/visualization.md`
+- `docs/architecture/integration-shape.md`
 - `docs/architecture/package-boundaries.md`
 - `docs/engineering/agentic-development-flow.md`
 …
+- `internal/agents/evaluator`
+- `internal/agents/optimizer`
 …
-Keep deterministic SearchBench model code in `internal/pure`. Keep orchestration in `internal/app`. …
+Keep deterministic SearchBench model code in `internal/pure`. Keep round lifecycle orchestration in `internal/app`. Colocate evaluator- and optimizer-specific behavior under `internal/agents` …
```

**Possible surrounding prose:**
Pair with the claim that agents need *maps* (where truth lives) in addition to *fewer commands* (how to verify).

---

# Top 3 diffs to include in the blog

1. `**cceb6a7` (`flake.nix`): flake-check vs module-graph hooks + `preCommitCheck = flakeCheckHooks`**
  - **Why it is visually strong:** Clear `-` / `+` split of hook sets; one-line comment explains *why* the split exists; `preCommitCheck` wiring is the punchline.
  - **Where it should go:** `## After: a graph` as **analog** (“layers of validation before Buck”), or `## Nix still owns the environment` to show sandbox boundaries.
  - **What claim it proves:** Not every sanctioned operation can run everywhere; name the contexts instead of listing ad hoc commands.
  - **Surrounding prose:** Short setup on `nix flake check` sandboxes—then let the diff carry the structure.
2. `**811c97c` (`nix/tools/core.nix` + `flake.nix` pre-push hook + `AGENTS.md` bullets)**
  - **Why it is visually strong:** Shows policy tightening (deterministic Repomix flags), a new **named** gate, and lifecycle binding (`pre-push`).
  - **Where it should go:** `## Removing knobs is a feature` or a “committed AI context artifact” aside.
  - **What claim it proves:** “Remember to refresh Repomix” → “the push lifecycle fails closed.”
  - **Surrounding prose:** One sentence on why diffs/logs were removed (`--no-git-sort-by-changes`) so the gate is byte-stable—then the excerpt.
3. `**9f8bb7c` (`AGENTS.md`): validation moves from `go test ./...` to hook tables**
  - **Why it is visually strong:** Agents recognize themselves in the “routinely run these tables” anti-pattern; dense `+` lines = cognitive load you are arguing against.
  - **Where it should go:** `## Before: a list of procedures` or `## This matters more for agents than for humans`.
  - **What claim it proves:** Good-faith documentation can still be an operational labyrinth; the fix is fewer entrypoints, not prettier tables.
  - **Surrounding prose:** Acknowledge this was a *correct* step (explicit lifecycle), then pivot: this is the ceiling of prose + Nix tables—next compression is graph targets.

---

## Themes with weak or missing committed diffs


| Theme                                                                          | Finding                                                                                                                        |
| ------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------ |
| `**buck2 test //:check` / root `test_suite`**                                  | No commits touch `BUCK`; `git log -S'buck2'` is empty on audited history.                                                      |
| **Toolchains / `.buckconfig.d`**                                               | No history in `git log` for these paths.                                                                                       |
| `**src/searchbench-go` / `src/iterative-context` layout + `BUCK` per subtree** | Not present as renames in committed history at audit time (working tree changes only).                                         |
| `**docs/architecture/build-system.md`**                                        | No commits (file not yet in history).                                                                                          |
| **Hooks delegating to a single Buck target**                                   | Not in history; current story is many discrete `searchbench-`* hook entries.                                                   |
| **Removing flake `apps` / shrinking `searchbench-`* surface**                  | History trend is mostly **growth** (`e5aaedb`, `6df67e5` adds `publish-issue-wave` app); shrinking diffs await future commits. |


---

*Report generated for curation into `content/posts/ai-cognitive-overload.mdx` (post body not modified here).*
