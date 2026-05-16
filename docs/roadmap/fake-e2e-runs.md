# Fake / scripted end-to-end runs

These paths prove **no-network / no-provider-credential** round execution: bundles, evidence, objectives, and decisions are produced using in-process **fake** backends and evaluator models. They are **not** proofs that real jCodeMunch or Iterative Context MCP servers work.

## Run matrix

| Name | Command | Network | Credentials | Bundle output | Decision | Status |
| --- | --- | --- | --- | --- | --- | --- |
| Fake-local CLI round | `go run ./cmd/searchbench run --manifest=configs/rounds/fake-local-e2e/round.pkl --bundle-root=<tmp>` (see `TestSearchBenchFakeLocalRoundRunCLIE2E`) | No | No MCP launcher vars; no LLM keys | Under `--bundle-root` (games layout) | `decision.json` in bundle | Covered by root `e2e_test.go` |
| Fake-local engine round | `round.Run` with same manifest (see `TestSearchBenchFakeLocalRoundRunEngineE2E`) | No | Same | Same | Same | Covered by root `e2e_test.go` |
| Example JC/IC manifest (smoke plumbing only) | `TestSearchBenchRoundRunCLIE2E`, `TestSearchBenchRoundRunEngineE2E` | No | MCP env empty → evaluator runs **fail**; bundle still written | Yes | Yes | Documents wiring only |
| Backend guard (JC/IC require MCP env) | `cd src/searchbench-go && go test ./internal/app/round -run TestFakeE2E_LocalManifestToolFactoryFailsWithoutMCPEnv` | No | No | — | — | **Honest failure** when launcher vars unset |
| Backend audit (manifest declares JC/IC) | `cd src/searchbench-go && go test ./internal/app/round -run TestFakeE2E_LocalManifestIncumbentUsesJCodeMunchBackend` | No | No | — | — | Documents `local-ic-vs-jcodemunch` |
| Fake-local backend audit | `cd src/searchbench-go && go test ./internal/app/round -run TestFakeE2E_FakeLocalManifestUsesFakeBackends` | No | No | — | — | Both sides `backend = fake` |
| Usage honesty (evidence) | `cd src/searchbench-go && go test ./internal/app/round -run TestProjectRoundEvidenceLeavesUsageUnavailableWhenRunsOmitUsage` | No | No | — | — | Usage stays **unavailable**, not fake-nonzero |
| Optimizer smoke (continuation) | `configs/rounds/optimize-ic/round.pkl` amends a completed bundle; round tests stub `OptimizerValidateProposal` so CI does **not** execute the full IC `uv` gate suite | — | — | Parent bundle | — | Fast stub path; production CLI leaves validator unset → full pipeline (`docs/reference/optimizer-policy-validation.md`) |

## Manifests

- **`configs/rounds/fake-local-e2e/round.pkl`** — Incumbent and challenger use **`backend = "fake"`**, `game.fakeEvaluator()`, shared **symlinked** LCA JSONL / objective / challenger policy path as `local-ic-vs-jcodemunch`. Safe default for offline CI.
- **`configs/rounds/local-ic-vs-jcodemunch/round.pkl`** — Declares **jcodemunch** and **iterative_context**. Useful for artifact plumbing tests; **does not** prove MCP backends without `SEARCHBENCH_JCODEMUNCH_COMMAND` / `SEARCHBENCH_ITERATIVE_CONTEXT_COMMAND`.

## Canonical bundle artifacts

After a successful run, the root e2e helpers expect at least:

`COMPLETE`, `resolved-round.json`, `round-report.json`, `round-report.txt`, `evidence.pkl`, `objective.json`, `decision.json`, `metadata.json`.

If the writer adds more files, update expectations only when the contract intentionally changes.

## Local commands

```bash
nix develop   # optional; preferred toolchain

# Focused fake-local proof (requires `pkl` on PATH)
cd src/searchbench-go
go test ./... -run 'TestSearchBenchFakeLocal|TestFakeE2E_'

# Full gate (matches CI helpers)
nix develop -c buck2 test //:check
cd src/searchbench-go && golangci-lint run ./...
cd src/searchbench-go && staticcheck ./...
cd src/searchbench-go && go test ./internal/architecture/...
nix flake check
```

## Env vars intentionally **not** required (fake-local)

- `SEARCHBENCH_JCODEMUNCH_COMMAND`
- `SEARCHBENCH_ITERATIVE_CONTEXT_COMMAND`
- Model provider API keys (evaluator `provider = "fake"`)

## What remains unproven here

- Real jCodeMunch and Iterative Context MCP sessions and tool semantics.
- Real LLM providers and LangSmith-backed traces.
- Full parity with the legacy Python SearchBench harness (behavioral reference only).

## Next step toward real JC/IC e2e

1. Configure MCP launcher env vars and run `configs/rounds/local-ic-vs-jcodemunch/round.pkl` (or equivalent) **outside** fake-local.
2. Confirm evaluator executions succeed (`round-report.json` failures empty) and compare evidence/decisions to expectations.
