You are operating as the GitHub coordination/control-plane agent for SearchBench-Go.

Your job is NOT to implement product code. Your job is to inspect the current SearchBench-Go repository, inspect the old Python SearchBench harness, inspect the `iterative-context` submodule/nested repo if present, inspect existing GitHub issues, then create a clean GitHub Issues control plane for the next implementation wave.

This is a single consolidated prompt. It contains both:

1. the GitHub issue/branch publishing instructions
2. the architectural issue inventory/spec for the next implementation wave

Do not begin creating issues until you have read and reconciled:

1. this entire prompt
2. the full current SearchBench-Go repo
3. the old Python SearchBench harness at `https://github.com/becker63/searchbench`
4. the current `iterative-context` submodule/nested repo state, if present
5. existing open GitHub issues
6. existing labels
7. current branch and working tree state

## Role

You are the GitHub issue/branch publishing operator.

You may:

- inspect the repo
- inspect the old Python harness
- inspect submodule/nested repo state
- inspect existing issues
- create missing labels
- create parent tracking issues
- create child implementation issues
- add issue comments
- create linked branches for parallel-safe issues

You must not:

- implement product code
- make large source changes
- modify pure modeling semantics
- create hidden automation
- create agent lifecycle/worktree orchestration inside the repo
- duplicate existing issues
- create branches for unclear or shared-core tasks unless explicitly safe
- vendor `iterative-context` code into SearchBench-Go
- silently ignore submodule/gitignore problems

## Architecture context

SearchBench-Go has largely completed its pure modeling phase. The next phase is real adapter/integration buildout.

Preserve this boundary:

- Pure core/model/scoring code should remain stable.
- Effectful adapters should be built at the edges.
- GitHub Issues are the control plane.
- Branches are implementation lanes.
- PRs/commits are implementation evidence.
- The repo itself should not own Cursor/Codex/agent lifecycle.
- The repo itself should not own worktree lifecycle for coding agents.
- Parallel agents should only be assigned to isolated conflict domains.

The target implementation wave moves SearchBench-Go toward:

```text
Pkl round manifest
→ resolved round
→ LCA match slice
→ repo materialization at base SHA
→ static graph indexing
→ incumbent jCodeMunch backend execution
→ challenger Iterative Context backend execution
→ deterministic harness-owned IC score/policy installation
→ evaluator finalization
→ localization-distance scoring
→ objective evaluation
→ decision
→ immutable round bundle
→ optional next-challenger proposal
```

## Initial checks

Use the current checked-out SearchBench-Go repository.

First run:

```bash
gh auth status
gh repo view --json nameWithOwner,defaultBranchRef
git status --short
git branch --show-current
```

If authentication is missing or the repo cannot be identified, stop and report exactly what is missing.

Do not continue if the working tree has uncommitted changes that could be confused with your own work. Report the dirty state and ask the human to resolve it.

## Required repository ingestion

Before creating or editing any GitHub issues, you must read the whole current SearchBench-Go repository and inspect the old Python SearchBench harness.

Do not rely only on shallow `find`, `grep`, or `head` output. Those commands are useful orientation checks, but they are not enough for this task.

### Read the full current SearchBench-Go repo

Use Repomix to produce a full AI-readable pack of the current checked-out SearchBench-Go repository into `/tmp` so the working tree is not modified:

```bash
nix develop -c repomix --style xml --compress -o /tmp/searchbench-go-repomix.xml .
```

If the Nix shell is unavailable, try:

```bash
repomix --style xml --compress -o /tmp/searchbench-go-repomix.xml .
```

Then inspect the pack:

```bash
wc -c /tmp/searchbench-go-repomix.xml
sed -n '1,220p' /tmp/searchbench-go-repomix.xml
grep -n "<directory_structure>" /tmp/searchbench-go-repomix.xml || true
grep -n "<file path=" /tmp/searchbench-go-repomix.xml | head -250
```

Use the packed repo as the main source of truth for package names, existing implementations, docs, tests, commands, and architecture boundaries.

If Repomix is unavailable, fall back to exhaustive tracked-file inspection:

```bash
git ls-files | sort > /tmp/searchbench-go-files.txt
cat /tmp/searchbench-go-files.txt
```

Then inspect relevant files directly. Do not proceed from only a shallow directory listing.

### Read the old Python SearchBench harness

Use Repomix remote mode to pack the old Python harness from GitHub:

```bash
repomix --remote https://github.com/becker63/searchbench --style xml --compress -o /tmp/searchbench-python-repomix.xml
```

If running inside the Nix dev shell is required:

```bash
nix develop -c repomix --remote https://github.com/becker63/searchbench --style xml --compress -o /tmp/searchbench-python-repomix.xml
```

Then inspect the pack:

```bash
wc -c /tmp/searchbench-python-repomix.xml
sed -n '1,220p' /tmp/searchbench-python-repomix.xml
grep -n "<directory_structure>" /tmp/searchbench-python-repomix.xml || true
grep -n "<file path=" /tmp/searchbench-python-repomix.xml | head -300
```

Use the old Python harness only as a reference for proven e2e behavior. Port useful behavior, not accidental architecture.

Pay special attention to old Python areas related to:

- LCA task loading and canonical task identity
- repo materialization/worktree cache
- jCodeMunch backend adapter
- Iterative Context backend adapter
- MCP tool conversion and dispatch
- localization finalization
- token usage extraction
- scoring/reducers
- CLI guardrails
- telemetry/session ownership
- optimization/writer loop

### Reconciliation rule

After reading both repo packs, create or reuse issues based on the actual gap between:

```text
old Python behavior that proved real e2e execution
current SearchBench-Go implementation
current open GitHub issues
this issue inventory/spec
```

Do not create issues for things already implemented in SearchBench-Go unless the issue is explicitly framed as hardening, integration, or parity work.

Do not create issues for old Python behavior that should intentionally not be ported.

## Required submodule / nested-repo handling

This repository may contain `iterative-context` as a git submodule or nested git repository used by the SearchBench-Go Iterative Context adapter work.

You must inspect it before creating IC-related issues.

Run:

```bash
cat .gitmodules 2>/dev/null || true
git submodule status || true
git ls-files --stage iterative-context 2>/dev/null || true
git check-ignore -v iterative-context iterative-context/ 2>/dev/null || true
git status --short --ignored iterative-context 2>/dev/null || true

if [ -d iterative-context/.git ] || [ -f iterative-context/.git ]; then
  git -C iterative-context status --short
  git -C iterative-context branch --show-current
  git -C iterative-context remote -v
  git -C iterative-context log --oneline -5
fi
```

Treat `iterative-context` as a separate repo lifecycle.

If an issue requires changes to Iterative Context itself, the issue body must say so explicitly and include a submodule workflow.

The implementation agent must:

1. enter the submodule/nested repo
2. create a branch inside `iterative-context`
3. make and commit the Iterative Context change there
4. push the Iterative Context branch if it has a remote
5. return to the SearchBench-Go parent repo
6. update the parent repo’s submodule pointer/gitlink
7. commit the parent repo change referencing the SearchBench-Go issue

Example workflow:

```bash
cd iterative-context
git status --short
git checkout -b ic/score-install-api-issue-<number>
# make IC server/API changes
git add .
git commit -m "Add session score install API for SearchBench-Go issue #<number>"
git push -u origin ic/score-install-api-issue-<number>

cd ..
git status --short
git add iterative-context
git commit -m "Update iterative-context submodule for issue #<number>"
```

If `iterative-context` is gitignored and `git add iterative-context` refuses to stage the submodule pointer, the implementation agent must not hack around it silently.

It should inspect:

```bash
git check-ignore -v iterative-context iterative-context/
git ls-files --stage iterative-context
cat .gitmodules 2>/dev/null || true
```

If the path is intended to be a tracked submodule but is blocked by `.gitignore`, the agent should propose or make the smallest parent-repo metadata fix needed to track the submodule pointer, such as:

```bash
git add -f .gitmodules iterative-context
```

or a precise `.gitignore` exception, depending on the actual repo state.

Do not vendor the submodule contents into the parent repository.

Do not copy Iterative Context code into SearchBench-Go.

Do not make SearchBench-Go own the Iterative Context server implementation.

SearchBench-Go may depend on a committed Iterative Context submodule revision, but the Iterative Context change itself must be committed in the Iterative Context repo first.

## Additional targeted current-repo inspection

After full Repomix ingestion, run additional targeted inspection commands to avoid hallucinating file paths or package names and to verify details that will appear in issue bodies.

Run commands like:

```bash
find . -maxdepth 3 -type f | sort | sed 's#^\./##' | head -300

grep -R "type .*Task\|type .*Score\|type .*Report\|package .*model\|package .*scoring\|package .*artifact\|package .*adapter" \
  -n . --include='*.go' | head -150

go list ./... 2>/dev/null || true
```

Use actual paths/packages discovered from the repo in issue bodies.

Do not invent package names if the repo does not contain them yet. In that case, phrase paths as proposed locations and make that explicit.

## Inspect existing issues

Before creating anything, inspect open issues:

```bash
gh issue list --state open --limit 150 --json number,title,labels,url
```

Also inspect labels:

```bash
gh label list --limit 200
```

Avoid duplicates.

If an existing open issue is equivalent to one requested by this prompt, reuse it. You may comment on it or include it in the parent tracking issue instead of creating a duplicate.

## Label taxonomy

Create missing labels only. Do not duplicate existing labels with slightly different names.

Start from this baseline taxonomy, adapting only if the repo already has a better convention:

- `kind/adapter`
- `kind/integration`
- `kind/contract`
- `kind/tooling`
- `area/dataset`
- `area/materialization`
- `area/codegraph`
- `area/scoring`
- `area/mcp`
- `area/eino`
- `area/telemetry`
- `area/artifacts`
- `area/report`
- `area/round`
- `area/cli`
- `area/github-control-plane`
- `parallel-safe`
- `blocked`
- `risk/shared-core`
- `touches/submodule`
- `touches/searchbench-go`
- `touches/iterative-context`

Use reasonable colors and descriptions.

Keep the taxonomy small. Do not create labels for every tiny subtopic.

## Existing Go implementation assumptions to verify

Do not blindly trust this list. Verify it by reading the repo.

The current Go repo likely already has:

- Pkl round manifests and schema under `configs/schema/` and `configs/rounds/`
- Pkl config adapter under `internal/adapters/config/pkl`
- Pkl scoring adapter under `internal/adapters/scoring/pkl`
- filesystem round bundle adapter under `internal/adapters/bundle/fs`
- text report adapter under `internal/adapters/report/text`
- fake evaluator/optimizer support under `internal/agents/*/fake`
- Eino evaluator/optimizer packages under `internal/agents/*/eino`
- app-level round orchestration under `internal/app/round`
- pure models under `internal/pure/*`
- code localization game package under `internal/games/codelocalization`
- codegraph pure package under `internal/pure/codegraph`
- architecture docs emphasizing Game / Round / Match / Evidence / Decision / NextChallenger

The next phase is not more pure modeling.

The next phase is real adapter/e2e buildout.

## Core architectural rule

Preserve the Go repo’s pure/effectful separation.

Pure packages may define:

- game/round/match/policy vocabulary
- execution records and failure kinds
- score/evidence/report models
- code graph models
- usage accounting models
- next-challenger proposal models

Pure packages must not import:

- Pkl runtime
- Eino
- MCP
- LangSmith/Langfuse
- filesystem writers
- os/exec
- network clients
- model provider SDKs

Adapters own real-world effects:

- dataset loading
- repo materialization
- tree-sitter/source parsing
- MCP client/session behavior
- Pkl loading/evaluation
- filesystem bundle writing
- provider calls
- trace/callback integrations
- CLI presentation

The app layer composes the round lifecycle.

The repo should not own Cursor/Codex/agent worktree lifecycle.

GitHub Issues are the control plane.

## Important existing issue drafts

The publisher should treat these as existing planned issues. If equivalent open GitHub issues already exist, reuse them. If they do not exist, create them or adapt them as child issues.

### Existing issue A: Client-owned IC score installation before evaluator runs

Purpose:

SearchBench-Go must install/configure the Iterative Context score or selection policy into the IC MCP session before the evaluator agent starts.

The evaluator agent must receive an already-prepared tool surface.

The evaluator must not install, select, mutate, verify, or hash scoring policy artifacts.

Required phase flow:

```text
prepare_score_artifact
start_ic_session
install_score
verify_score
expose_evaluator_tools
run_evaluator
finalize_prediction
complete
```

Failure kinds should include:

```text
score_artifact_invalid
ic_session_failed
score_install_failed
score_verify_failed
evaluator_tool_setup_failed
evaluator_failed
evaluator_tool_call_failed
finalization_failed
invalid_prediction
```

This issue is a dependency for the real Iterative Context backend/evaluator issue.

### Existing issue B: Materialize LCA task repositories for harness-side evaluation

Purpose:

Given a SearchBench LCA match/task, materialize the referenced GitHub repository at `base_sha` into a local read-only snapshot suitable for graph indexing and scoring.

This is harness-side infrastructure.

It must not own:

- scoring
- tree-sitter indexing
- evaluator execution
- writer execution
- MCP lifecycle
- model calls
- tracing SDKs

It should define a materialization package such as:

```text
internal/adapters/materialize/git
```

or another repo-appropriate adapter location.

It should have offline tests using local git fixtures or scripted command runners.

Failure kinds should include:

```text
invalid_repo_url
clone_failed
fetch_failed
checkout_failed
missing_base_sha
cache_permission_failed
invalid_task_repo
filesystem_error
```

This issue is a dependency for static graph indexing and real e2e scoring.

### Existing issue C: Tree-sitter/static code graph indexing for materialized LCA repositories

Purpose:

Given a materialized repo, build a static graph supporting deterministic localization-distance scoring.

This should use the existing pure codegraph model if present.

Minimum graph nodes:

- file nodes
- Python function nodes
- Python class nodes if cheap

Minimum edges:

- file contains symbol
- symbol belongs to file
- import/reference edges if cheap
- call edges only if reliable enough

Initial deterministic metric:

```text
localization_distance
```

Suggested scalar:

```text
best_gold_hop_distance
```

or:

```text
average_gold_nearest_prediction_distance
```

The issue must make the choice explicit.

Token efficiency is secondary/diagnostic and must not be mixed into graph distance.

This issue depends on materialization.

## Important submodule dependency: Iterative Context lives in `iterative-context`

Some real IC adapter work may require changes inside the `iterative-context` repo that is present inside SearchBench-Go as a submodule or nested git repository.

The issue publisher must inspect the submodule state and make IC-related issues explicit about whether they touch:

```text
SearchBench-Go parent repo only
Iterative Context submodule only
both repos
```

If an issue requires both repos, it must include a cross-repo/submodule commit plan.

The correct lifecycle is:

```text
change iterative-context
→ commit inside iterative-context
→ update SearchBench-Go submodule pointer/gitlink
→ commit parent SearchBench-Go change
```

The implementation agent must not vendor `iterative-context` code into SearchBench-Go.

The implementation agent must not ignore the submodule pointer update.

If `iterative-context` is currently gitignored, the issue should instruct the agent to inspect whether it is intentionally ignored, accidentally ignored, or already tracked as a submodule. If it is intended to be a real submodule, the agent should make the smallest safe parent-repo metadata fix needed to track the submodule pointer.

## Parent tracking issue

Create or reuse a parent issue titled:

```text
Real e2e adapter buildout for SearchBench-Go
```

The parent issue should explain:

- the pure core is now stable enough to build adapters around it
- this wave ports the useful old Python e2e behavior into Go
- issues are split by adapter/conflict domain
- parallel agents should work only on parallel-safe issues
- shared app/runtime wiring should be serialized
- first milestone is a tiny real LCA round, not a large benchmark run
- Iterative Context may require submodule/nested-repo changes that must be committed separately

The parent issue should include this wave plan:

```text
Wave 0: audit and reconcile existing issues
Wave 1: parallel-safe substrates/adapters
Wave 2: backend/evaluator execution
Wave 3: first real round e2e
Wave 4: optimizer and coordination hardening
```

The parent issue should include:

- wave plan
- child issue checklist
- dependency notes
- parallelization notes
- validation expectations
- submodule/nested-repo notes
- explicit non-goals

After child issues are created or identified, update the parent issue body or add a comment with links to the child issues.

## Wave 0: Audit/reconcile current state

Create an issue if no equivalent exists:

```text
Audit existing Go implementation against old Python e2e path
```

Labels:

- `kind/contract`
- `area/github-control-plane`
- `risk/shared-core`

Purpose:

Before launching many agents, compare:

- current SearchBench-Go repo
- old Python SearchBench e2e behavior
- current `iterative-context` submodule/nested repo state
- existing GitHub issues
- architecture docs

The output should be a checked-in doc or issue comment mapping:

```text
already implemented
partially implemented
missing adapter
missing integration
requires iterative-context change
should not port
```

This is not a product-code issue unless a tiny doc file is needed.

Acceptance criteria:

- identifies existing Go packages that already cover bundle/config/scoring/fake round flow
- identifies Python features that should be ported
- identifies Python features that should not be ported
- identifies IC work requiring submodule changes
- maps open issues to the new wave plan
- flags duplicate/stale issues that can be closed or consolidated
- does not implement e2e runtime behavior

Parallel safety:

- not parallel-safe with issue creation/publishing
- safe before implementation starts

Suggested branch:

```text
contract/e2e-gap-audit-issue-<number>
```

## Wave 1: Parallel-safe substrates/adapters

These issues can mostly run in parallel if they avoid shared app-layer wiring.

### Issue 1: Implement LCA dataset adapter for SearchBench matches

Labels:

- `area/dataset`
- `kind/adapter`
- `parallel-safe`
- `touches/searchbench-go`

Conflict domain:

```text
dataset loading / match materialization
```

Goal:

Implement the adapter that turns the configured LCA dataset slice from a Pkl round manifest into SearchBench match/task values.

The current Pkl helpers already express:

```pkl
game.lca("py", "dev", 5)
```

The adapter should translate that into the existing pure/domain match/task model used by the Go app.

Use the old Python LCA task model and selection behavior as reference, but do not copy its hosted/Langfuse-first assumptions unless explicitly needed.

Scope:

- read resolved dataset config from the Go round manifest path
- support JetBrains LCA dataset identity/config/split/maxItems
- produce deterministic match/task values
- normalize repo owner/name, base SHA, issue title/body, changed files
- preserve gold labels as scoring-only data, never prompt-visible data
- include deterministic IDs
- include offline fixtures/tests

Non-goals:

- no repo cloning
- no graph indexing
- no scoring
- no MCP
- no evaluator/model execution
- no LangSmith/Langfuse
- no large hosted dataset abstraction
- no agent lifecycle

Acceptance criteria:

- Pkl LCA dataset config can resolve to a bounded match slice
- deterministic sorting/windowing is implemented or explicitly deferred
- gold changed files are retained for scoring but excluded from evaluator prompt payloads
- tests use fixtures or fake loaders by default
- no default test requires network access
- pure model packages do not import dataset adapter packages

Suggested branch:

```text
adapter/lca-dataset-issue-<number>
```

### Issue 2: Reuse/create materialization issue

Use Existing issue B.

If no equivalent issue exists, create it.

Labels:

- `area/materialization`
- `kind/adapter`
- `parallel-safe`
- `touches/searchbench-go`

Conflict domain:

```text
repo materialization / git cache
```

Dependency:

- can start once LCA task/match repo identity is clear
- should not require full dataset adapter to be complete if tests use direct materialization requests

Important adaptation:

Place materialization under adapter/effectful code, not pure core.

Suggested branch:

```text
adapter/lca-repo-materialization-issue-<number>
```

### Issue 3: Reuse/create static graph indexing and localization-distance scoring issue

Use Existing issue C.

If no equivalent issue exists, create it.

Labels:

- `area/codegraph`
- `area/scoring`
- `kind/adapter`
- `parallel-safe`
- `blocked`
- `touches/searchbench-go`

Conflict domain:

```text
static graph indexing / graph-distance scoring
```

Dependencies:

- materialized repo substrate
- existing pure codegraph package audit

Important adaptation:

The current Go repo may already have `internal/pure/codegraph`.

Do not create a parallel `internal/codegraph` model if the pure codegraph package already provides the model. Extend or adapt the existing pure package and put tree-sitter/indexing effects at an adapter/builder boundary.

Suggested branch:

```text
adapter/static-graph-localization-score-issue-<number>
```

### Issue 4: Implement real jCodeMunch backend adapter

Labels:

- `area/mcp`
- `kind/adapter`
- `parallel-safe`
- `touches/searchbench-go`

Conflict domain:

```text
jcodemunch backend adapter
```

Goal:

Implement the real jCodeMunch backend adapter for SearchBench-Go.

The old Python harness had a `JCodeMunchBackend` that listed MCP tools, initialized a repo via local path or remote repo, converted MCP tools into model tool specs, and dispatched tool calls.

The Go implementation should fit the existing evaluator/backend abstractions instead of copying Python structure.

Scope:

- start/acquire jCodeMunch MCP client/session
- initialize/index a materialized repo path
- expose only evaluator-allowed tools
- convert tool specs into the Eino/tool surface expected by the evaluator
- dispatch tool calls and normalize results into SearchBench execution records
- return typed setup/tool-call failures
- support fake or scripted MCP tests

Non-goals:

- no IC policy installation
- no dataset loading
- no repo materialization
- no graph scoring
- no optimizer
- no hosted tracing dependency
- no agent lifecycle/worktree orchestration

Acceptance criteria:

- adapter can expose a stable tool surface from a fake/scripted jCodeMunch MCP server
- adapter can initialize against a local materialized repo path
- admin/setup behavior is not exposed as evaluator behavior unless intentionally allowed
- typed failures distinguish setup failure from tool-call failure
- default tests do not require live jCodeMunch server or network

Suggested branch:

```text
adapter/jcodemunch-mcp-issue-<number>
```

### Issue 5: Add Iterative Context MCP score install/verify API

Labels:

- `area/mcp`
- `kind/adapter`
- `blocked`
- `touches/submodule`
- `touches/iterative-context`
- `touches/searchbench-go`

Conflict domain:

```text
iterative-context MCP server API / score install and verify
```

Repo touched:

```text
iterative-context submodule / nested repo
SearchBench-Go parent repo only for submodule pointer update
```

Goal:

Add or verify the Iterative Context MCP server API needed for SearchBench-Go to install and verify a session-bound score/selection policy before evaluator execution.

Scope:

- expose a deterministic install/configure score or policy operation
- expose a deterministic verify-active-score operation
- bind the installed policy to the active IC session/runtime
- ensure evaluator-visible tools use the installed score internally
- ensure admin/install/verify tools can be hidden from the evaluator-facing tool surface
- add IC-side tests for install-before-use behavior
- commit the IC change inside the `iterative-context` repo
- update the SearchBench-Go parent repo submodule pointer

Non-goals:

- no SearchBench-Go evaluator wiring
- no Go provider/model work
- no dataset loading
- no graph scoring
- no optimizer loop
- no vendoring IC code into SearchBench-Go

Acceptance criteria:

- IC exposes install/configure and verify operations
- active score identity can be checked deterministically
- evaluator tools can operate against the installed score
- IC tests cover missing/mismatched score behavior
- SearchBench-Go parent repo records the updated submodule pointer
- issue body includes submodule workflow

Suggested branch inside `iterative-context`:

```text
ic/score-install-api-issue-<number>
```

Suggested parent branch if needed:

```text
adapter/iterative-context-submodule-issue-<number>
```

Submodule workflow to include in the issue:

```bash
cd iterative-context
git status --short
git checkout -b ic/score-install-api-issue-<number>
# make IC server/API changes
git add .
git commit -m "Add score install API for SearchBench-Go issue #<number>"
git push -u origin ic/score-install-api-issue-<number>

cd ..
git status --short
git add iterative-context
git commit -m "Update iterative-context submodule for issue #<number>"
```

### Issue 6: Implement real Iterative Context backend adapter with harness-owned score install seam

Labels:

- `area/mcp`
- `kind/adapter`
- `parallel-safe`
- `blocked`
- `touches/searchbench-go`
- `touches/submodule`

Conflict domain:

```text
iterative-context backend adapter / score install seam
```

Goal:

Implement the real Iterative Context backend adapter in Go, using the score install/verify API from `iterative-context`.

This issue may require coordinated changes in both:

- SearchBench-Go parent repo
- the `iterative-context` submodule/nested repo

The publisher must inspect the current `iterative-context` state before creating this issue. If the IC MCP server does not yet expose the needed install/verify API, the publisher should create the separate Iterative Context API issue above and mark this one blocked by it.

Scope:

- start/acquire IC MCP session for a materialized repo
- install/configure challenger selection policy before evaluator starts
- verify active score/policy identity
- expose only evaluator-safe localization tools
- hide install/verify/admin tools from the evaluator agent
- dispatch evaluator tool calls
- surface typed failures for session, install, verify, tool setup, and tool calls
- support fake/scripted MCP tests

Non-goals:

- no evaluator prompt changes unless required by tool schema
- no optimizer generation
- no dataset loading
- no graph scoring
- no model provider implementation
- no agent lifecycle/worktree orchestration
- no vendoring IC code into SearchBench-Go

Dependencies:

- Existing issue A: client-owned IC score installation
- Iterative Context MCP score install/verify API
- materialization may be needed for real repo path, but tests can use fake sessions

Acceptance criteria:

- score/policy install happens deterministically before evaluator execution
- active score/policy identity is verified before tool exposure
- evaluator-visible tool list excludes admin/install/verify operations
- failure kinds are observable and typed
- fake/scripted tests prove install-before-run ordering
- default tests do not require live IC server or network
- issue body includes whether it touches parent repo only or both parent and submodule

Suggested branch:

```text
adapter/iterative-context-mcp-issue-<number>
```

### Issue 7: Implement provider/model adapter for real evaluator calls

Labels:

- `area/eino`
- `kind/adapter`
- `parallel-safe`
- `touches/searchbench-go`

Conflict domain:

```text
model provider adapter / evaluator model calls
```

Goal:

Wire real model-provider calls into the existing Eino evaluator path without making provider DTOs part of the SearchBench model.

Scope:

- resolve provider/model settings from the existing Pkl evaluator config
- support at least one real provider path configured by environment
- preserve fake/scripted model tests
- collect model usage when provider returns it
- expose missing usage as unavailable, not zero
- avoid making LangSmith/Langfuse the source of truth for token usage
- keep provider SDK details out of pure packages

Non-goals:

- no MCP backend implementation
- no scoring
- no repo materialization
- no optimizer provider unless already shared cleanly
- no pricing/projection system unless the repo already has a clear place for it

Acceptance criteria:

- fake provider tests continue to pass
- real provider wiring is isolated behind an adapter/boundary
- usage accounting records available/unavailable status explicitly
- no provider SDK types leak into pure model packages
- validation passes without credentials by default

Suggested branch:

```text
adapter/evaluator-provider-issue-<number>
```

### Issue 8: Implement LangSmith/trace callback adapter boundary

Labels:

- `area/telemetry`
- `area/eino`
- `kind/adapter`
- `parallel-safe`
- `touches/searchbench-go`

Conflict domain:

```text
callbacks / tracing / usage observation
```

Goal:

Implement or harden the tracing/callback adapter boundary for real evaluator runs.

The old Python harness used Langfuse heavily, but SearchBench-Go should keep SearchBench artifacts authoritative.

Tracing is useful as an external observation surface.

It must not own:

- scoring
- evidence
- decision
- bundle identity
- token truth

Scope:

- attach Eino callbacks for model/tool/runtime observations
- capture phase transitions and failure categories
- attach usage details if available
- preserve local SearchBench usage accounting
- allow tests to run without network/credentials
- keep trace SDKs out of pure packages

Non-goals:

- no scoring logic
- no report generation
- no dataset loading
- no hosted dataset store
- no mandatory tracing for local fake runs

Acceptance criteria:

- evaluator callbacks can record model/tool observations
- tracing can be disabled/no-op in tests
- missing credentials do not fail normal unit tests
- SearchBench bundle/report/evidence remains source of truth
- trace adapter package does not define core report/evidence types

Suggested branch:

```text
adapter/langsmith-callbacks-issue-<number>
```

## Wave 2: Backend/evaluator execution wiring

These issues are more integration-heavy. They should generally be serialized or handled by one agent at a time.

### Issue 9: Wire real evaluator runtime across prepared backend tools

Labels:

- `area/eino`
- `kind/integration`
- `blocked`
- `touches/searchbench-go`

Conflict domain:

```text
evaluator runtime orchestration
```

Goal:

Connect the real backend adapters to the existing evaluator runtime so the evaluator can run one SearchBench match against either incumbent or challenger policy.

Scope:

- select backend from resolved round system backend id
- prepare backend/session/tool surface
- run evaluator over allowed tools
- finalize localization prediction using strict structured output
- record usage, tool calls, observations, failures, phases
- normalize final prediction into existing domain prediction types
- preserve retry/finalization behavior already modeled in Go

Non-goals:

- no dataset loading
- no materialization implementation
- no static graph scoring
- no optimizer loop
- no custom state machine if Eino/app flow already provides orchestration
- no agent/worktree lifecycle

Dependencies:

- jCodeMunch backend adapter
- Iterative Context backend adapter
- provider/model adapter
- callback boundary, if tracing is required

Acceptance criteria:

- one fake/scripted backend can run through the real evaluator runtime
- finalization produces typed predictions
- failures are classified by phase
- evaluator does not install IC scores or mutate policy artifacts
- app/pure packages retain proper dependency direction

Suggested branch:

```text
integration/evaluator-runtime-issue-<number>
```

### Issue 10: Build round match execution path for incumbent vs challenger

Labels:

- `area/round`
- `kind/integration`
- `blocked`
- `touches/searchbench-go`

Conflict domain:

```text
app round match execution
```

Goal:

Wire a resolved round so each selected match runs both:

```text
IncumbentPolicy
ChallengerPolicy
```

under the configured code-localization game.

Scope:

- take resolved round + match slice
- materialize repo per match or reuse cache
- prepare per-policy backend execution
- run incumbent and challenger executions
- collect match records
- preserve deterministic ordering
- support bounded parallelism only if existing app conventions already support it
- keep failures as match execution evidence rather than panics when possible

Non-goals:

- no new generic scheduler
- no agent lifecycle/worktree ownership
- no optimizer loop
- no large benchmark run
- no hosted platform dependency

Dependencies:

- dataset adapter
- materialization
- evaluator runtime
- backend adapters

Acceptance criteria:

- fake/scripted round can execute incumbent/challenger over at least one match
- match execution records include policy role, prediction, usage, failure, and timing where available
- deterministic ordering is preserved
- validation passes without real network/model credentials

Suggested branch:

```text
integration/round-match-execution-issue-<number>
```

### Issue 11: Build localization evidence from match executions

Labels:

- `area/scoring`
- `area/report`
- `kind/integration`
- `blocked`
- `touches/searchbench-go`

Conflict domain:

```text
round evidence construction
```

Goal:

Build code-localization `RoundEvidence` from incumbent/challenger match executions.

Scope:

- compare incumbent/challenger predictions against gold changed files
- include localization-distance components from static graph scoring
- include usage summaries if available
- count invalid predictions
- count failures
- identify regressions/severe/protected regressions if the existing model supports them
- emit scoring-facing evidence compatible with the existing Pkl objective shape

Non-goals:

- no tree-sitter indexing implementation
- no evaluator execution
- no bundle writer implementation from scratch
- no objective Pkl runtime changes unless required by evidence shape

Dependencies:

- static graph/localization scoring
- round match execution path
- existing score/report models

Acceptance criteria:

- evidence can be built from fixture match records
- evidence includes incumbent/challenger role fields
- evidence includes match counts, failure counts, usage summaries, localization distance
- missing graph/gold/prediction data is represented explicitly
- Pkl objective fixture can consume the evidence shape

Suggested branch:

```text
integration/localization-evidence-issue-<number>
```

## Wave 3: First real round e2e

These issues should be serialized.

### Issue 12: Add tiny real LCA round smoke command

Labels:

- `area/cli`
- `kind/integration`
- `blocked`
- `touches/searchbench-go`

Conflict domain:

```text
CLI / real e2e smoke path
```

Goal:

Add a tiny command/path that runs a real or near-real LCA round through the Go app.

The milestone is not performance.

The milestone is an end-to-end artifact-producing round.

Target command shape should follow the project’s game/round vocabulary, for example:

```text
searchbench round run --manifest configs/rounds/local-ic-vs-jcodemunch/round.pkl
```

or the closest current CLI surface.

Scope:

- load Pkl round manifest
- resolve LCA match slice with small maxItems
- materialize repo(s)
- run incumbent/challenger using fake/scripted or real backends depending on available config
- score/build evidence
- evaluate objective
- decide
- write round bundle
- print bundle path and decision summary

Non-goals:

- no large benchmark
- no optimizer loop
- no remote hosted dataset requirement
- no mandatory live model call in default tests
- no hidden background automation

Dependencies:

- dataset adapter
- materialization
- evaluator runtime
- match execution
- evidence/scoring
- bundle alignment

Acceptance criteria:

- one command can run a tiny bounded round
- default test path can use fake/scripted providers/backends
- output bundle is written
- command refuses unsafe unbounded runs if applicable
- CLI remains thin over app layer

Suggested branch:

```text
integration/tiny-real-round-smoke-issue-<number>
```

### Issue 13: Align round bundle artifacts with canonical real-e2e bundle shape

Labels:

- `area/artifacts`
- `kind/integration`
- `blocked`
- `touches/searchbench-go`

Conflict domain:

```text
round bundle artifact compatibility
```

Goal:

Harden or adjust the existing filesystem bundle writer so real e2e rounds produce the canonical durable bundle shape.

Do not reimplement the bundle writer from scratch if the existing `internal/adapters/bundle/fs` package already covers most of this.

Canonical target:

```text
artifacts/games/<game-id>/rounds/<round-id>/
  COMPLETE
  resolved-round.json
  round-report.json
  round-report.txt
  evidence.pkl
  objective.json
  decision.json
  metadata.json
  next-challenger.json?
  replay-events.json?
  visualization.json?
```

Scope:

- compare current bundle outputs against canonical docs
- add missing decision artifact if needed
- ensure metadata captures artifact inventory/hashes where existing conventions support it
- ensure bundle is immutable after COMPLETE where practical
- ensure reports and objective refs point to the right files

Non-goals:

- no scoring logic
- no evaluator execution
- no optimizer implementation
- no trace platform dependency

Dependencies:

- evidence/objective/decision output shape
- existing bundle writer

Acceptance criteria:

- fixture round can write canonical bundle artifacts
- metadata includes the written artifact inventory
- COMPLETE marker behavior is clear
- writer tests cover canonical filenames
- no pure package imports filesystem writer

Suggested branch:

```text
integration/canonical-round-bundle-issue-<number>
```

### Issue 14: Produce release-report alpha from real round evidence

Labels:

- `area/report`
- `kind/integration`
- `blocked`
- `touches/searchbench-go`

Conflict domain:

```text
round report / decision rendering
```

Goal:

Produce the first useful release-report alpha from real or scripted round evidence.

Scope:

- render round identity, game identity, incumbent/challenger policies
- render match count and outcome summary
- render localization-distance comparison
- render token/usage comparison if available
- render regressions/failures/invalid predictions
- render objective final value and important intermediate values
- render decision: PROMOTE / REVIEW / REJECT
- link report to bundle artifacts

Non-goals:

- no visualization UI
- no graph replay UI
- no new scoring engine
- no hosted traces requirement

Dependencies:

- evidence construction
- objective result
- decision artifact
- canonical bundle alignment

Acceptance criteria:

- text report explains who won and why
- JSON report is stable enough for visualization later
- report remains generated from evidence/objective/decision, not ad hoc runtime state
- tests use fixtures

Suggested branch:

```text
integration/release-report-alpha-issue-<number>
```

## Wave 4: Optimizer and coordination hardening

These should wait until the first e2e path exists.

### Issue 15: Connect optimizer to completed round bundle as NextChallenger proposal

Labels:

- `area/eino`
- `area/artifacts`
- `kind/integration`
- `blocked`
- `touches/searchbench-go`

Conflict domain:

```text
optimizer / next-challenger proposal
```

Goal:

Connect the optimizer to completed round evidence so it can propose a `NextChallenger`, not mutate the current round.

Scope:

- read allowed evidence from parent/current round bundle
- deny gold labels/oracle files/raw answers
- pass objective result, report summary, current challenger policy as allowed
- produce next-challenger proposal artifact
- validate generated policy artifact shape
- write next-challenger artifact into bundle or next-round staging location according to current architecture

Non-goals:

- no automatic promotion
- no mutation of completed rounds
- no immediate recursive optimization loop
- no agent lifecycle orchestration
- no hidden worktree management

Dependencies:

- completed round bundle
- release-report alpha
- optimizer fake/eino path audit

Acceptance criteria:

- optimizer can run from fixture bundle evidence
- denied evidence is not included in prompt input
- proposal is clearly future-oriented
- generated artifact is validated
- output is represented as `NextChallengerProposal`

Suggested branch:

```text
integration/next-challenger-proposal-issue-<number>
```

### Issue 16: Add GitHub issue manifest and gh publisher script

Labels:

- `area/github-control-plane`
- `kind/tooling`
- `touches/searchbench-go`

Conflict domain:

```text
external coordination tooling
```

Goal:

Add a small, explicit coordination tool for publishing future issue waves through the GitHub CLI.

This is not SearchBench runtime code.

Scope:

- define a manifest format for issue waves
- support dry-run mode
- support duplicate detection by title
- call `gh issue create`
- optionally call `gh issue develop` for parallel-safe issues
- document how to run it

Non-goals:

- no Cursor/Codex lifecycle management
- no worktree orchestration
- no autonomous background agent runner
- no product runtime dependency
- no hidden GitHub writes

Acceptance criteria:

- dry-run prints planned labels/issues/branches
- publish mode creates missing issues
- duplicate open issues are reused/skipped
- script is clearly dev tooling
- docs warn that GitHub Issues are the control plane, not the SearchBench runtime

Suggested branch:

```text
tooling/issue-manifest-publisher-issue-<number>
```

## Issue creation priorities

Create/reuse issues in this order:

1. Parent tracking issue
2. Audit/reconcile issue
3. Existing issue A: IC score installation
4. Existing issue B: LCA repo materialization
5. Existing issue C: static graph/localization scoring
6. LCA dataset adapter
7. jCodeMunch MCP adapter
8. Iterative Context MCP score install/verify API
9. Iterative Context Go backend adapter
10. provider/model adapter
11. telemetry/callback adapter
12. evaluator runtime wiring
13. round match execution
14. localization evidence construction
15. tiny real round smoke command
16. canonical round bundle alignment
17. release-report alpha
18. optimizer next-challenger proposal
19. issue manifest / gh publisher script

If this is too many issues for one publishing pass, create:

- parent issue
- audit issue
- Wave 1 issues
- Wave 2 blockers

Then stop and report the remaining planned issues.

## Parallelization guidance

Safe to run in parallel after issue creation:

```text
LCA dataset adapter
repo materialization
jCodeMunch MCP adapter
provider/model adapter
telemetry/callback adapter
```

Conditionally parallel:

```text
static graph indexing
Iterative Context MCP score install/verify API
Iterative Context Go backend adapter
```

These may touch shared score/codegraph/backend/submodule surfaces and should be assigned carefully.

Do not parallelize initially:

```text
evaluator runtime wiring
round match execution
localization evidence construction
tiny real e2e smoke command
canonical bundle alignment
release-report alpha
optimizer connection
```

Those are shared integration surfaces.

## Branch creation guidance

Create linked branches only for clearly parallel-safe issues.

Create parent SearchBench-Go branches for:

```text
adapter/lca-dataset-issue-<number>
adapter/lca-repo-materialization-issue-<number>
adapter/jcodemunch-mcp-issue-<number>
adapter/evaluator-provider-issue-<number>
adapter/langsmith-callbacks-issue-<number>
```

For Iterative Context issues, do not assume one branch is enough.

If the issue touches the submodule, include the submodule branch plan in the issue body. If appropriate, create the parent repo branch only for the parent gitlink update, and instruct the implementation agent to create the submodule branch inside `iterative-context`.

Do not create branches by default for:

```text
integration/evaluator-runtime-issue-<number>
integration/round-match-execution-issue-<number>
integration/tiny-real-round-smoke-issue-<number>
integration/release-report-alpha-issue-<number>
```

unless you have inspected the repo and confirmed they are safe to start.

Prefer:

```bash
gh issue develop <issue-number> --name <branch-name> --base <default-branch>
```

Do not checkout branches unless necessary.

Do not make implementation commits.

If `gh issue develop` fails, fall back to recording the suggested branch name in the issue body or a comment. Do not manually create branches unless you are confident the repo policy allows it.

## Dependency handling

Mark issues as `blocked` when they depend on unresolved adapter contracts, submodule API work, or shared integration work.

Do not create linked branches for blocked integration issues unless this prompt explicitly says they are safe to start and you have confirmed the repo structure supports it.

Parallel-safe adapter issues should have clear conflict domains.

Examples of conflict domains:

- dataset loading / match materialization
- repo materialization / git cache
- static graph indexing / graph-distance scoring
- baseline MCP adapter
- candidate MCP adapter
- Iterative Context submodule API
- callback telemetry / token accounting
- provider/model config
- CLI integration

## Child issue body requirements

For each issue requested by this prompt:

1. Search existing open issues for an equivalent issue.
2. Reuse equivalent issues instead of duplicating them.
3. Create missing issues with precise, implementation-ready bodies.
4. Add appropriate labels.
5. Include dependency/blocked notes.
6. Include parallel-safety notes.
7. Include a suggested branch name.
8. Include validation commands.
9. Include repo-touch scope: SearchBench-Go, iterative-context, or both.
10. Include submodule workflow when relevant.

Each child issue must include:

- Goal
- Background
- Scope
- Explicit non-goals
- Expected files/packages, based on actual repo inspection
- Acceptance criteria
- Validation commands
- Parallelization/conflict-domain notes
- Suggested branch name
- Repo touched
- Agent instructions
- Submodule workflow, if applicable

Every child issue must tell the future implementation agent to:

- read the issue with `gh issue view <number>`
- create/use the suggested branch
- keep changes small
- avoid pure-core semantic changes unless the issue explicitly allows them
- run validation before reporting done
- commit with a message referencing the issue number
- report any architectural mismatch instead of papering over it
- for submodule work, commit inside `iterative-context` first, then update the parent repo pointer

## Validation expectations

Each issue should include validation commands discovered from the repo.

Prefer current repo commands such as:

```bash
nix develop --command go test ./...
nix flake check
go test ./...
golangci-lint run
```

Only include commands that appear appropriate after inspecting the repo.

Every issue should say:

- default tests must not require network access
- real provider/backend tests must be opt-in or skipped without credentials
- pure packages must not import adapters/agents/surface
- implementation agents should report architecture mismatches instead of forcing code through

For `iterative-context` submodule issues, include validation commands discovered from inside the submodule, for example:

```bash
cd iterative-context
# run the submodule's own tests/checks discovered from its repo
```

Do not invent submodule validation commands without inspecting its files.

## Safety and correctness rules

Do not claim something exists unless you inspected it.

Do not claim an issue, label, or branch was created unless the command succeeded.

Do not silently skip requested issues. If you skip one, explain why.

Do not create duplicate issues.

Do not create broad vague issues like “finish integrations.” Split by adapter boundary and conflict domain.

Do not create issues that require multiple parallel agents to edit the same core files at the same time.

Do not mutate product code.

Do not create a custom agent orchestrator inside the SearchBench-Go repo.

Do not create a worktree lifecycle manager inside SearchBench-Go.

Do not vendor `iterative-context` into SearchBench-Go.

Do not ignore submodule pointer updates.

Do not treat `iterative-context` as ordinary ignored files if it is intended to be a submodule/nested repo.

Do not port old Python behavior that is accidental, over-generalized, or contrary to the Go architecture.

## Non-goals for the whole wave

Do not create issues that turn SearchBench-Go into:

- a generic workflow engine
- a generic agent framework
- a hosted eval platform
- a trace viewer clone
- a background agent scheduler
- a worktree lifecycle manager
- a Cursor/Codex orchestration system inside the product repo

The project spine remains:

```text
Game
→ Round
→ Match
→ Evidence
→ Objective
→ Decision
→ RoundBundle
→ NextChallenger
```

## Final desired state

This implementation wave is successful when SearchBench-Go can run a tiny real or scripted LCA round through the Go architecture and produce a durable bundle like:

```text
artifacts/games/code-localization/rounds/round-001/
  COMPLETE
  resolved-round.json
  round-report.json
  round-report.txt
  evidence.pkl
  objective.json
  decision.json
  metadata.json
```

A human should be able to inspect the bundle and understand:

```text
this was the game
this was the round
these were the matches
this was the incumbent
this was the challenger
these were the predictions
this was the evidence
this was the objective score
this was the decision
this is what should happen next
```

## Final report

At the end, report:

- parent issue number and URL
- child issue numbers and URLs
- which requested issues were created
- which requested issues were reused
- labels created or reused
- branches created or skipped
- blocked/dependent issues
- submodule-related issues
- any commands that failed
- any assumptions made because repo structure was ambiguous
- remaining planned issues not created in this pass

Do not over-explain. The final report should be operationally useful to the human who will launch Cursor/Codex agents against these issues.

## Load-bearing instruction

Use this prompt as the desired issue inventory, but do not blindly create duplicates.

Inspect the current repository, the old Python harness, the `iterative-context` submodule/nested repo, and open GitHub issues.

If a requested issue is already implemented or already open, reuse/comment/link it instead of creating a duplicate.

If current Go code already partially implements a requested area, write the issue as a hardening/integration issue rather than a greenfield implementation issue.

Prefer small, bounded, adapter-shaped issues over broad “finish e2e” issues.

The control plane is GitHub Issues.

The implementation evidence is branches, commits, PRs, tests, and bundles.

SearchBench-Go owns the game, the rounds, and the evidence-backed judgment.

Agents generate artifacts. SearchBench decides whether those artifacts survive.
