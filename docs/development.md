# Development workflow

Routine validation is **Git-driven** after `nix develop`: hooks run Buck aggregates and Repomix on commit/push.

## Environment

```bash
nix develop    # Go, Pkl, buck2, repomix, pre-commit; writes .buckconfig.d/buck2-nix.config
```

`nix flake check` — sandboxed formatting/Nix/shell hygiene only; **no** Buck tests in the default sandbox.

## Validation commands

| When | Command |
| --- | --- |
| Pre-commit (via `git commit`) | Repomix regenerate + stage, then `buck2 test //:check` |
| Pre-push (via `git push`) | `buck2 test //:check_full` |
| Manual Go | `cd src/searchbench-go && go test ./...` |
| Manual Buck (fast) | `nix develop -c buck2 test //:check` |
| Manual Buck (full) | `nix develop -c buck2 test //:check_full` |
| All hooks locally | `nix develop -c pre-commit run --all-files` |

### What `//:check` and `//:check_full` include

- **`//:check`** — `//src/searchbench-go:check` (Go tests + CLI build) and `//src/iterative-context:check` (uv sync, import smoke, pytest subset without `TEST_REPO_*` fixtures).
- **`//:check_full`** — above plus IC `check_full` (basedpyright) and `//:repomix_fresh_check` (fails if `repomix-output.xml` is not committed at `HEAD`).

## Repomix

This repo **commits** `repomix-output.xml` as an AI-review artifact. `repomix.config.json` excludes `repomix-output.*` so the pack does not recurse.

- Pre-commit regenerates and **stages** the snapshot.
- Pre-push runs a **freshness gate** — commit the updated snapshot before pushing; hooks do not auto-amend.

For a richer one-off pack, run `repomix` manually with `--include-diffs` / `--include-logs`.

## Pkl Go bindings

After editing `configs/schema/SearchBenchRound.pkl`:

```bash
cd src/searchbench-go
pkl run package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl \
  --output-path=. \
  ../../configs/schema/SearchBenchRound.pkl
```

Opt-in Buck target `//src/searchbench-go:pkl_go_types` regenerates bindings in-tree; do not run it in parallel with `go_tests` (file races).

## golangci-lint / staticcheck

`.golangci.yml` is for **local** and CI use. Hooks do not run golangci automatically:

```bash
cd src/searchbench-go
nix develop -c golangci-lint run ./...
```

## Submodule: `src/iterative-context`

Git submodule for the Iterative Context Python tree and MCP server. Buck targets under `//src/iterative-context:*` run IC checks from the monorepo root.

Update the submodule pointer when bumping IC; run `go test` and `buck2 test //:check_full` before pushing.

## Agentic / issue workflow

Issue-first development loop (ChatGPT → GitHub issue → agent prompt → review): [reference/agentic-development-flow.md](./reference/agentic-development-flow.md).

Batch issue publishing (dev tooling): [archive/issue-wave-manifest.md](./archive/issue-wave-manifest.md).

## Related

- Root [AGENTS.md](../AGENTS.md) — short contributor contract
- [architecture.md](./architecture.md) — package boundaries summary
- [reference/build-system.md](./reference/build-system.md) — Nix cell and Buck target notes
