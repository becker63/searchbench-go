# Development

**Nix** supplies the toolchain and installs Git hooks. **Buck** names repo operations. Raw commands below are debugging fallbacks only.

## Environment

```bash
nix develop
```

- Installs Go, Pkl, buck2, node/npm, pre-commit.
- Writes `.buckconfig.d/buck2-nix.config` for the Buck Nix cell.
- Installs Git hooks via git-hooks.nix — **do not** hand-edit `.git/hooks/`*.

`nix flake check` — formatting, Nix, and shell hygiene only (no Buck test graph in the sandbox).

## Repo gates


| When         | Buck target                | Role                     |
| ------------ | -------------------------- | ------------------------ |
| `git commit` | `buck2 test //:check`      | Fast validation          |
| `git push`   | `buck2 test //:check_full` | Full deterministic gate  |
| Manual fast  | `buck2 test //:check`      | Same as commit gate      |
| Manual full  | `buck2 test //:check_full` | Same as pre-push         |


```bash
nix develop -c buck2 test //:check
nix develop -c buck2 test //:check_full
```

## Target catalog


| Area                  | Build (mutating)                    | Test (proof)                                                          |
| --------------------- | ----------------------------------- | --------------------------------------------------------------------- |
| **Whole repo (fast)** | —                                   | `//:check`                                                            |
| **Whole repo (full)** | —                                   | `//:check_full`                                                       |
| **Go harness**        | `//src/searchbench-go:pkl_go_types` | `//src/searchbench-go:check`, `//src/searchbench-go:pkl_go_types_check` |
| **Iterative Context** | —                                   | `//src/iterative-context:check`, `//src/iterative-context:check_full` |
| **Docs site**         | `//docs:site`                       | `//docs:check`                                                        |


### `//:check` includes

- `//src/searchbench-go:check` — `go test ./...` + CLI build
- `//src/iterative-context:check` — IC import smoke + pytest subset

### `//:check_full` includes

- Everything in `//:check` (via harness + IC full targets)
- `//src/searchbench-go:pkl_go_types_check` — generated Pkl bindings match `HEAD`
- `//src/iterative-context:check_full` — adds basedpyright
- `//docs:check` — VitePress production build

After editing `configs/schema/SearchBenchRound.pkl`:

```bash
buck2 build //src/searchbench-go:pkl_go_types
buck2 test //src/searchbench-go:pkl_go_types_check
```

Bump `src/iterative-context` submodule pointer only after:

```bash
buck2 test //src/iterative-context:check_full
buck2 test //:check_full
```

## Docs site (hosted)

GitHub Actions deploys `main` to [becker63.github.io/searchbench-go](https://becker63.github.io/searchbench-go/). Local proof: `buck2 test //docs:check`.

## Debugging fallbacks

Not the canonical proof interface. Use when Buck is unavailable or for interactive work.


| Intent       | Fallback command                                                                                                                                          |
| ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Go tests     | `cd src/searchbench-go && go test ./...`                                                                                                                  |
| Docs preview | `npm ci && npm run docs:dev`                                                                                                                              |
| Docs build   | `npm ci && npm run docs:build`                                                                                                                            |
| Pkl → Go gen | `cd src/searchbench-go && pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl --output-path=. ../../configs/schema/SearchBenchRound.pkl` |


## See also

[components.md](./components.md) · [AGENTS.md](../AGENTS.md) · [index.md](./index.md)
