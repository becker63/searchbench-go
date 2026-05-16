# Start here

SearchBench tests **agent environments**: which tools, search interfaces, and configuration surfaces help the same agent perform better on a fixed benchmark slice.

Vocabulary and definitions: [concepts.md](./concepts.md). Product hook on the repo: [README.md](../README.md).

## What this repo is (today)

- A **research / product workbench** for controlled rounds and durable **bundles**
- **Infra-native** validation (Nix, Buck2, Pkl) for contributors
- **First game:** code localization — bug-localization slices to compare symbol/code-search interfaces with lookahead

Not the default story yet: autonomous week-long interface optimization loops. Long term, agents may propose interface changes and SearchBench may evaluate them; today the focus is **comparable candidates** and **trustworthy evidence**.

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
| **Round app** | Compare incumbent vs challenger → evidence → decision → bundle |
| **Agents** | Evaluator and optimizer; optional `NextChallenger` proposal |
| **Bundles** | Durable round output; input for reports and visualization |

## Run one local round

Copy-paste from the repo root (after build): [README § Run one local round](../README.md#run-one-local-round). Offline fake-local path. Validation: [development.md](./development.md).

## Read next

| Doc | When |
| --- | --- |
| [concepts.md](./concepts.md) | Game, Interface, dataset slice, Round, bundle |
| [architecture.md](./architecture.md) | Package layers |
| [development.md](./development.md) | Nix, Buck2, hooks |
| [workspace-seeds.md](./workspace-seeds.md) | IC workspace providers |
| [README.md](./README.md) | Docs index |

Contributors: [AGENTS.md](../AGENTS.md).
