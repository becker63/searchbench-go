# GitHub issue wave manifest (publisher tooling)

SearchBench’s **control plane** is GitHub Issues; this repository also carries a small **publisher** so operators can batch-create coordinated issues from JSON. This is **not** SearchBench runtime code and is never imported by product packages.

## Manifest format

JSON object:

| Field | Required | Meaning |
| --- | --- | --- |
| `repo` | No | `OWNER/REPO` passed to `gh -R`. If omitted, `gh` uses the current repository context. |
| `issues` | Yes | Non-empty array of issue entries. |

Each **issue** entry:

| Field | Required | Meaning |
| --- | --- | --- |
| `title` | Yes | Issue title (duplicate open titles are skipped after whitespace trim). |
| `body` | No* | Markdown body inline. |
| `body_file` | No* | Path to body file, **relative to the git repo root**. |
| `labels` | No | Array of label names (must already exist on GitHub or `gh` fails). |
| `parallel_safe` | No | Default `false`. When `true` and `--develop` is passed, runs `gh issue develop` after create. |
| `develop_branch_name` | No | Branch name for `gh issue develop`; default `issue-<num>-develop`. |

\* Provide at most one of `body` or `body_file`.

## Commands

From `nix develop` (or any environment with the flake tools on `PATH`):

```bash
# Preview — no GitHub writes
nix develop -c searchbench-publish-issue-wave --dry-run tooling/issue-wave.example.json

# Create missing issues (skips open duplicates by title)
nix develop -c searchbench-publish-issue-wave tooling/my-wave.json

# Optional: gh issue develop for entries with parallel_safe: true
nix develop -c searchbench-publish-issue-wave --develop tooling/my-wave.json
```

Flake app shortcut:

```bash
nix run .#publish-issue-wave -- --dry-run tooling/issue-wave.example.json
```

## Safety notes

- **`gh` authentication** is required for publish mode; dry-run still lists open issues via the API.
- **Duplicate detection** compares trimmed titles against **open** issues only.
- **`gh issue develop`** may fail depending on org/repo settings; the script prints a warning and continues.
- Prefer **`--dry-run`** in CI or unfamiliar repos — there is **no hidden write mode**.

See also: Issue 16 scope in `prompts.md` (original intent) and umbrella tracking issue `#38` / `#39` history on GitHub.
