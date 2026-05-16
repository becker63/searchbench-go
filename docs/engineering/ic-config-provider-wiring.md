# IC config and workspace seed provider wiring

## Split of responsibilities

| Layer | Declares |
| --- | --- |
| **Pkl** (`configs/schema/SearchBenchRound.pkl`) | Provider intent: `local_path` or `buck_descriptor` |
| **Buck** | Legal repo operations (`//src/iterative-context:optimizable_backend`, `//:check`, …) |
| **Bundle / evidence** | What happened: seed identity, validation steps, runtime identity |

Buck is **not** required for public users. Pkl remains logical and provider-neutral at the schema level.

## Schema

```pkl
typealias WorkspaceSeedProvider = "local_path" | "buck_descriptor" | "git" | "archive"

class Runtime {
  workspaceSeed: WorkspaceSeedConfig?
}

class WorkspaceSeedConfig {
  provider: WorkspaceSeedProvider = "local_path"
  localPath: String?
  buckDescriptorTarget: String?
}
```

Go validation: `internal/adapters/config/pkl/workspace_seed.go`

- `local_path` → `localPath` required
- `buck_descriptor` → `buckDescriptorTarget` required
- `git` / `archive` → rejected (reserved)

## Examples

### Public / default (`local_path`)

```pkl
system = new System {
  id = "iterative-context"
  backend = "iterative_context"
  runtime = new Runtime {
    workspaceSeed = new WorkspaceSeedConfig {
      provider = "local_path"
      localPath = "src/iterative-context"
    }
  }
}
```

### Internal (`buck_descriptor`)

```pkl
runtime = new Runtime {
  workspaceSeed = new WorkspaceSeedConfig {
    provider = "buck_descriptor"
    buckDescriptorTarget = "//src/iterative-context:optimizable_backend"
  }
}
```

See also [ic-workspace-seed-providers.md](./ic-workspace-seed-providers.md).
