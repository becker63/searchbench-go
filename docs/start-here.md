# Start here

SearchBench compares agent and tool **candidates** on the same match slice and produces **release evidence**. Vocabulary: [concepts.md](./concepts.md).

## How the pieces fit

```text
Pkl declares intent.
Providers resolve source identity.
Candidate workspaces isolate mutation.
Validation proves the candidate.
Runtime launches the accepted candidate.
Bundles record what happened.
```

| Layer | Role |
| --- | --- |
| **Pkl manifests** | Round intent: policies, backends, scoring, workspace seeds |
| **Workspace seeds** | Where Iterative Context (or other backends) copy from before validation |
| **Round app** | Orchestrate compare → evidence → objective → decision → bundle |
| **Agents** | Evaluator and optimizer (prompts + Eino); propose `NextChallenger` |
| **Bundles** | Durable round output under a bundle root |

## Run one local round

Copy-paste from the repo root (after build): see [README.md § Run one local round](../README.md#run-one-local-round). Uses the **offline fake-local** path. Validation: [development.md](./development.md).

## Read next

| Doc | When |
| --- | --- |
| [concepts.md](./concepts.md) | Product vocabulary |
| [architecture.md](./architecture.md) | Package layers and boundaries |
| [development.md](./development.md) | Nix, Buck2, Go, hooks, Repomix |
| [workspace-seeds.md](./workspace-seeds.md) | `local_path` vs `buck_descriptor` for IC |
| [README.md](./README.md) | Full docs index |

Contributors and agents: read root [AGENTS.md](../AGENTS.md) for the operational contract (boundaries, validation commands, non-goals).
