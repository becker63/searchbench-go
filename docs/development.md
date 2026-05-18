# Development

**Nix** supplies the toolchain and Git hooks. **Buck** is the canonical proof and repo-owned run interface. Raw commands are debugging fallbacks only.

## Environment

```bash
nix develop
```

- Installs Go, Buck2, `uv`, Pkl, node/npm, pre-commit. Buck prelude is pinned in `.buckconfig` (git external cell; no local `prelude/` checkout). The shell hook also materializes `outputs/out/lib` symlinks so prelude `go_test` binaries can load `libstdc++.so.6`.
- Hooks: `git commit` → `buck2 test //:check`; `git push` → `buck2 test //:check_full`.
- Shell hook prints Go/IC dependency lifecycle reminders (no automatic regen).

## Canonical workflow

```bash
nix develop
buck2 test //:check          # fast — same as pre-commit
buck2 test //:check_full     # full — same as pre-push
```

| Target | Proves |
| --- | --- |
| `buck2 test //:check` | Native Go fast suite + deterministic CLI manifest validation + IC smoke |
| `buck2 test //:check_full` | Above + IC full + docs build + Pkl bindings |
| `buck2 test //docs:check` | VitePress production build |
| `buck2 test //src/searchbench-go:check` | Native Go fast suite + deterministic CLI manifest validation |
| `buck2 test //src/searchbench-go:go_native_full` | Same as `:check` without IC/docs/Pkl gates |
| `buck2 test //src/iterative-context:check_full` | `uv sync` + pytest + basedpyright |
| `buck2 build //src/searchbench-go/cmd/searchbench:searchbench` | CLI binary (`go_binary`) |
| `buck2 run //src/searchbench-go:pkl_go_types` | Regenerate Go from Pkl schema |
| `buck2 test //src/searchbench-go:pkl_go_types_check` | Generated code matches `HEAD` |

### `//:check` includes

- `//src/searchbench-go:check`
- `//src/iterative-context:check`

### `//:check_full` includes

- `//src/searchbench-go:check`
- `//src/searchbench-go:pkl_go_types_check`
- `//src/iterative-context:check_full`
- `//docs:check`

Repo-owned live MCP evaluation (not in `//:check`): see [reference/live-e2e.md](./reference/live-e2e.md) and [reference/run-entrypoints.md](./reference/run-entrypoints.md).

```bash
buck2 test //configs/rounds/live-ic-vs-jcodemunch:validate_bundle   # deterministic
buck2 test //configs/rounds/live-ic-vs-jcodemunch:live_smoke        # fresh live (secrets)
```

After editing `configs/schema/SearchBenchRound.pkl`:

```bash
buck2 run //src/searchbench-go:pkl_go_types
buck2 test //src/searchbench-go:pkl_go_types_check
```

## Dependency lifecycle

### SearchBench-Go (deep Buck modeling)

- **Third-party deps:** `go mod vendor` + **gobuckify** into `src/searchbench-go/vendor/` (see `gobuckify.json`). The tree is **generated locally** via `nix run .#project-go-deps`; it is gitignored and not packed by Repomix.
- **First-party checks:** prelude `go_library` / `go_test` / `go_binary`.
- **Native fast suite:** `buck2 test //src/searchbench-go:go_native_fast` — deterministic prelude `go_test` targets (see coverage table in [BUCK_MIGRATION.md](../src/searchbench-go/BUCK_MIGRATION.md)).
- **Native full suite:** `buck2 test //src/searchbench-go:go_native_full` — fast suite plus heavier deterministic targets and `//configs/rounds/live-ic-vs-jcodemunch:validate`.
- **Buck validates; it does not run `go mod vendor` or gobuckify during `buck2 test`.**

After `go.mod`, `go.sum`, `gobuckify.json`, `flake.lock`, or `toolchains/flake.lock` changes:

```bash
nix run .#project-go-deps
```

`nix develop` warns when `vendor/` or `vendor/.searchbench-vendor-projection.json` is missing or stale. See [../src/searchbench-go/BUCK_MIGRATION.md](../src/searchbench-go/BUCK_MIGRATION.md).

Future direction: a pure Nix derivation in `/nix/store` instead of a writable `vendor/` tree; this pass only pins tools via the Nix app.

Native examples: `buck2 test //src/searchbench-go:go_native_fast`, `buck2 test //src/searchbench-go/internal/pure/domain:domain_test`, `buck2 build //src/searchbench-go/cmd/searchbench:searchbench`.

### iterative-context (Python runtime wrappers)

- **Not modeled in Buck.** No Elk, no per-wheel `prebuilt_python_library` graph.
- Dependencies: `uv lock` / `uv sync` in `src/iterative-context` (Nix provides `uv`).
- Buck exposes stable targets: `import_smoke`, `pytest_all`, `basedpyright_check`, `check`, `check_full`.

After `pyproject.toml` changes:

```bash
cd src/iterative-context
uv lock
uv sync --locked
```

## Example round (local)

Repo-owned rounds use Buck targets under `configs/rounds/<name>/`. For the live round, see [reference/run-entrypoints.md](./reference/run-entrypoints.md) and [configs/rounds/live-ic-vs-jcodemunch/README.md](../configs/rounds/live-ic-vs-jcodemunch/README.md).

Do not use `./searchbench round run` or direct CLI invocation as the normal workflow; Buck invokes the native Go binary as private plumbing.

## Docs site

Hosted: [becker63.github.io/searchbench-go](https://becker63.github.io/searchbench-go/).
**Proves with:** `buck2 test //docs:check`

## Debugging fallbacks

| Intent | Command |
| --- | --- |
| Go tests | `cd src/searchbench-go && go test ./...` |
| IC tests | `cd src/iterative-context && uv run pytest` |
| Docs preview | `npm ci && npm run docs:dev` |
| Docs build | `npm ci && npm run docs:build` |
| Pkl → Go gen | see [reference/pkl-rounds.md](./reference/pkl-rounds.md) |

## See also

[components.md](./components.md) · [AGENTS.md](../AGENTS.md) · [index.md](./index.md)
