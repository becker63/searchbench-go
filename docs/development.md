# Development

## Environment

```bash
nix develop   # Go, Pkl, buck2, node/npm, repomix, git-hooks (only hook installer)
```

`nix flake check` — formatting/Nix/shell only; no Buck tests in the sandbox.

## Validation

| When | What |
| --- | --- |
| `git commit` | Repomix stage + `buck2 test //:check` |
| `git push` | `buck2 test //:check_full` (Go, IC, **docs build**, Repomix freshness) |

```bash
cd src/searchbench-go && go test ./...
nix develop -c buck2 test //:check
nix develop -c buck2 test //:check_full
nix develop -c buck2 test //docs:check
```

## Repomix

`repomix-output.xml` is committed for AI review. Pre-commit regenerates and stages it; pre-push fails if it is not at `HEAD`.

## Pkl Go bindings

After `configs/schema/SearchBenchRound.pkl` changes:

```bash
cd src/searchbench-go
pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl \
  --output-path=. ../../configs/schema/SearchBenchRound.pkl
```

Do not run `//src/searchbench-go:pkl_go_types` in parallel with `go_tests`.

## Docs site

```bash
npm ci
npm run docs:dev      # preview
npm run docs:build    # → docs/.vitepress/dist
```

Published: [becker63.github.io/searchbench-go](https://becker63.github.io/searchbench-go/) (GitHub Actions on `main`).

## Submodule

`src/iterative-context` — bump pointer with `go test` and `buck2 test //:check_full`.

## See also

[AGENTS.md](../AGENTS.md) · [index.md](./index.md)
