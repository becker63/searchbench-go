# Development

**Nix** supplies the toolchain and Git hooks. **Buck** is the canonical proof and repo-owned run interface. Raw commands are debugging fallbacks only.

## Environment

```bash
nix develop
```

- Installs Go, Pkl, buck2, node/npm, pre-commit.
- Hooks: `git commit` → `buck2 test //:check`; `git push` → `buck2 test //:check_full`.

## Canonical workflow

```bash
nix develop
buck2 test //:check          # fast — same as pre-commit
buck2 test //:check_full     # full — same as pre-push
```

| Target | Proves |
| --- | --- |
| `buck2 test //:check` | Go harness + IC smoke |
| `buck2 test //:check_full` | Above + IC full + docs build + Pkl bindings |
| `buck2 test //docs:check` | VitePress production build |
| `buck2 test //src/searchbench-go:check` | `go test ./...` + CLI build |
| `buck2 test //src/iterative-context:check_full` | pytest + basedpyright |
| `buck2 build //src/searchbench-go:pkl_go_types` | Regenerate Go from Pkl schema |
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
buck2 build //src/searchbench-go:pkl_go_types
buck2 test //src/searchbench-go:pkl_go_types_check
```

## Example round (local)

Repo-owned rounds use Buck targets under `configs/rounds/<name>/`. For the live round, see [reference/run-entrypoints.md](./reference/run-entrypoints.md) and [configs/rounds/live-ic-vs-jcodemunch/README.md](../configs/rounds/live-ic-vs-jcodemunch/README.md).

Do not use `./searchbench round run` or direct CLI invocation as the normal workflow; Buck invokes the Go binary as private plumbing.

## Docs site

Hosted: [becker63.github.io/searchbench-go](https://becker63.github.io/searchbench-go/).
**Proves with:** `buck2 test //docs:check`

## Debugging fallbacks

| Intent | Command |
| --- | --- |
| Go tests | `cd src/searchbench-go && go test ./...` |
| Docs preview | `npm ci && npm run docs:dev` |
| Docs build | `npm ci && npm run docs:build` |
| Pkl → Go gen | see [reference/pkl-rounds.md](./reference/pkl-rounds.md) |

## See also

[components.md](./components.md) · [AGENTS.md](../AGENTS.md) · [index.md](./index.md)
