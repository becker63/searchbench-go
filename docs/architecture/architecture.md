# SearchBench Architecture

**Related docs:** [Visualization](./visualization.md) · [Integration shape](./integration-shape.md) · [Package boundaries](./package-boundaries.md) · [Pkl manifests](./pkl-round-manifests.md) · [LangSmith](../integrations/langsmith-integration.md)

## Purpose

This document defines the project architecture and naming model for SearchBench.

SearchBench is a game-pluggable evaluation system for reviewing AI policy changes across datasets.

The core product shape is:

```text
Game
  → Round
      → Match
      → Evidence
      → Decision
      → NextChallenger
````

A `Game` defines the domain contract.

A `Round` compares an `IncumbentPolicy` against a `ChallengerPolicy` under that game.

A `Match` is one dataset item inside the round.

SearchBench is not primarily a tracing wrapper, prompt optimizer, benchmark script, or generic agent framework.

SearchBench owns the review structure that decides whether generated artifacts survive.

Agents are generators inside that structure.

---

## North star

SearchBench helps people review AI behavior across a dataset without needing to be data scientists.

The product answer is not a spreadsheet of metrics.

The product answer is a reviewable game:

```text
define the game
run the incumbent against the challenger
inspect the matches
read the evidence
make a decision
optionally generate the next challenger
```

For the first game, code localization:

```text
Game: Code Localization

Round 002
  IncumbentPolicy ─┐
                   ├─ Matches over dataset ── Evidence ── Decision ── NextChallenger
  ChallengerPolicy ┘
```

The important object is not a single trace.

The important object is not even a single round by itself.

The important product shape is:

```text
Game → Round → Match Review → Decision
```

---

## Product vocabulary

The following terms are project-level vocabulary and should be used in docs, reports, visualizations, high-level app packages, CLI output, and generated artifacts.

### `Game`

A `Game` is the domain contract.

It defines what kind of AI behavior is being evaluated and how humans should review it.

A game owns:

* task/match schema
* domain-specific evidence model
* scoring inputs
* allowed review panes
* replay format
* decision rule expectations
* domain language
* win/loss semantics for one match
* aggregate round semantics

A game is not one execution.

A game is the ruleset.

Examples:

```text
CodeLocalizationGame
RetrievalQAGame
SupportResponseGame
ToolUsingResearchGame
WorkflowAgentGame
```

The first game is:

```text
CodeLocalizationGame
```

It defines:

* repository/codebase as the board
* bug/task issue as the match input
* file/symbol localization as the goal
* graph traversal as the main review pane
* hop distance, file coverage, usage, and regressions as evidence

### `Round`

A `Round` is one evaluation contest under a game.

A round compares one `IncumbentPolicy` against one `ChallengerPolicy` across a fixed match slice.

A round produces:

* match executions
* match outcomes
* round evidence
* objective result
* decision
* durable round bundle
* optional next-challenger proposal

A round is immutable after completion.

### `Match`

A `Match` is one dataset item inside a round.

In old language, this is often a “task.”

In product and review language, use `Match`.

A match contains:

* one domain-specific input
* one incumbent execution
* one challenger execution
* per-side outputs
* per-side usage
* per-side failures if any
* match-level evidence
* match-level outcome

For code localization, a match is one localization problem.

For support response quality, a match might be one customer conversation.

For retrieval QA, a match might be one question with expected evidence.

### `IncumbentPolicy`

The current policy being defended.

This may be:

* a production policy
* a fixed incumbent/control
* the previous winning challenger
* a known external reference system such as jCodeMunch

The incumbent is not necessarily “old” or “bad.”

It is the policy the challenger must beat.

### `ChallengerPolicy`

The policy currently being tested.

This may include changes to:

* prompt bundle
* policy code
* retrieval strategy
* graph traversal behavior
* model/provider settings
* runtime bounds
* tool behavior

The challenger is the thing under evaluation.

### `Evidence`

The durable facts produced by a round and its matches.

Evidence is not the same thing as a trace.

A trace helps debug behavior.

Evidence supports a decision.

Evidence includes:

* per-match outcomes
* scores
* regressions
* token usage
* cost usage
* failure counts
* relevant artifact hashes
* objective inputs
* report-safe summaries
* domain-specific evidence

### `Decision`

The result of applying the game’s objective and release rule to the round evidence.

Valid decision values:

```text
PROMOTE
REVIEW
REJECT
```

Decision meaning:

* `PROMOTE`: challenger is safe to advance under the current rule.
* `REVIEW`: challenger has meaningful upside but needs human review.
* `REJECT`: challenger should not advance.

A decision is not merely a score label.

It is the explicit release judgment for the round.

### `NextChallenger`

The proposed challenger for a future round.

The optimizer does not “improve the system” directly.

The optimizer proposes a `NextChallenger`.

That proposal may become the `ChallengerPolicy` in a later round.

---

## Core hierarchy

Use this hierarchy everywhere:

```text
Game
  Round
    Match
```

### `Game = domain contract`

A game answers:

```text
What kind of AI behavior are we reviewing?
What counts as success?
What evidence matters?
What review panes make failures understandable?
```

### `Round = one evaluation contest under that contract`

A round answers:

```text
Did this ChallengerPolicy beat this IncumbentPolicy under this Game?
```

### `Match = one dataset item inside a round`

A match answers:

```text
What happened on this one case?
Which side won this case?
Why?
```

---

## First game: Code Localization

The first concrete game is `CodeLocalizationGame`.

### Goal

Given a bug report or localization prompt, identify the files, symbols, or code regions relevant to the change.

### Board

The board is a repository/code graph.

The board may include:

* files
* symbols
* functions
* classes/types
* imports
* calls
* references
* graph edges
* known target files/symbols for scoring

### Moves

Moves are the actions a policy can take through tools.

Examples:

* search code
* inspect file
* inspect symbol
* follow references
* follow calls/imports
* expand graph neighbors
* request structured context

### Evidence

Code localization evidence may include:

* gold file coverage
* predicted files
* predicted symbols
* hop distance to gold nodes
* token usage
* tool-call count
* invalid predictions
* regressions
* protected regressions
* trace/replay events
* graph traversal behavior

### Review panes

The main review pane is the side-by-side graph replay:

```text
IncumbentPolicy graph traversal
vs
ChallengerPolicy graph traversal
```

The viewer should be able to see:

* where each side searched
* which nodes/files each side reached
* how close each side got to the target
* where tokens were spent
* why the challenger won or lost

Localization is domain-specific.

The game model is not.

---

## Domain-pluggability

SearchBench should be domain-pluggable, not domain-agnostic.

“Domain-agnostic” risks vague abstractions.

“Domain-pluggable” means the core lifecycle is stable, but each game provides its own evidence and review surface.

### Shared core

The SearchBench core owns:

```text
Game registration / contract
Round lifecycle
Policy roles
Match execution structure
Evidence bundle shape
Decision shape
Lineage
NextChallenger flow
Visualization/replay envelope
```

### Game-specific layer

Each game owns:

```text
match schema
domain evidence
domain objective inputs
domain report sections
review panes
match outcome classification
domain-specific replay data
```

### Example game packs

```text
Code Localization Game
  - repository graph board
  - localization matches
  - hop-distance evidence
  - file/symbol target evidence
  - graph replay panes

Retrieval QA Game
  - question/evidence board
  - answer matches
  - citation coverage evidence
  - unsupported-claim regressions
  - source comparison panes

Support Response Game
  - conversation board
  - response matches
  - policy compliance evidence
  - escalation correctness evidence
  - conversation review panes

Tool-Using Research Game
  - research task board
  - claim/source matches
  - source trail evidence
  - tool-call timeline panes

Workflow Agent Game
  - workflow state board
  - task completion matches
  - invalid action evidence
  - step graph replay panes
```

---

## Terms to retire from public architecture

These terms may remain in low-level code temporarily during migration, but they should not define the project model.

### `incumbent`

Replace with:

```text
incumbent
```

`incumbent` sounds static. `incumbent` implies a policy defending its position.

### `challenger`

Replace with:

```text
challenger
```

`challenger` sounds generic. `challenger` implies motion and competition.

### `task`

Replace publicly with:

```text
match
```

`task` may remain in dataset adapters where the source dataset calls items tasks.

In core architecture, reports, and visualization, use `Match`.

### `decision decision`

Replace with:

```text
decision
```

Promotion is one possible decision. The object itself is the round decision.

### `local e2e`

Replace with:

```text
round
```

The current round flow is not test scaffolding. It is the main system shape.

### `projection`

Avoid as public vocabulary.

Use more specific names:

* `Evidence`
* `Report`
* `View`
* `Visualization`
* `ObjectiveResult`

A projection is how something is derived. It should not be the product-facing noun.

---

## Conceptual model

The stable architecture is:

```text
GameSpec
  ↓
RoundSpec
  ↓
ResolvedRound
  ↓
MatchExecutions
  ↓
RoundEvidence
  ↓
ObjectiveResult
  ↓
Decision
  ↓
RoundBundle
  ↓
NextChallengerProposal
```

### `GameSpec`

The domain contract for a game.

Contains:

* game id
* game kind
* domain name
* match schema reference
* evidence schema reference
* default review panes
* default objective inputs
* allowed replay formats
* match outcome semantics

### `RoundSpec`

The caller-provided intent for one round.

Contains:

* game reference
* round id/name
* incumbent policy
* challenger policy
* match slice
* objective reference
* runtime bounds
* output settings
* optional parent round reference
* optional optimizer configuration

### `ResolvedRound`

The normalized, validated, concrete round plan.

Contains:

* resolved game contract
* resolved manifest
* resolved paths
* resolved policy artifacts
* resolved match slice
* resolved objective path
* resolved bundle output path
* resolved execution settings

`ResolvedRound` is before execution.

It should be deterministic and serializable.

### `MatchExecutions`

The act of running the incumbent and challenger for each match.

Contains:

* per-match incumbent execution
* per-match challenger execution
* trace refs
* usage records
* failures
* execution timing
* domain replay refs

This is where effectful execution occurs.

### `RoundEvidence`

The report-safe durable evidence derived from match executions.

Contains:

* incumbent/challenger comparison
* per-match outcomes
* per-match score deltas
* aggregate score deltas
* usage deltas
* regressions
* protected regressions
* invalid predictions
* failure counts
* artifact refs
* game-specific evidence

This is the main scoring input.

### `ObjectiveResult`

The evaluated scoring objective.

Contains:

* named intermediate values
* penalties
* final value
* evidence refs
* bounds

The objective explains what “better” means for this game and round.

### `Decision`

The release judgment.

Contains:

* decision value: `PROMOTE`, `REVIEW`, or `REJECT`
* reason
* rule summary
* important deltas
* blocking regressions if any

### `RoundBundle`

The durable artifact directory for a completed round.

Contains:

* resolved round input
* structured report
* rendered report
* evidence
* objective result
* decision
* metadata
* optional replay/trace events
* optional game-specific visualization data
* optional next-challenger proposal

### `NextChallengerProposal`

The optimizer output.

Contains:

* proposed policy artifact
* summary of change
* evidence used
* evidence denied
* validation feedback
* target game/round hint

The optimizer proposes a next challenger.

It does not mutate the current round.

---

## Temporal architecture

Time should be modeled as lineage, not hidden procedure.

Bad mental model:

```text
evaluation service calls optimizer service
optimizer returns response
caller mutates result
```

Good mental model:

```text
Game
  Round N bundle
    ↓ evidence
  Decision
    ↓ optimizer reads bounded evidence
  NextChallenger proposal
    ↓
  Round N+1 ChallengerPolicy
```

The top-level flow should be readable as a story:

```go
game, err := game.Resolve(ctx, input.Game)
resolved, err := round.Resolve(ctx, game, input)
matches, err := round.EvaluateMatches(ctx, resolved, deps)
evidence, err := round.BuildEvidence(ctx, resolved, matches)
objective, err := round.EvaluateObjective(ctx, resolved, evidence)
decision := round.Decide(resolved, evidence, objective)
next, err := round.ProposeNextChallenger(ctx, resolved, evidence, objective, decision)
bundle, err := round.WriteBundle(ctx, resolved, evidence, objective, decision, next)
```

The implementation does not need to use exactly these functions, but the top-level app flow should preserve this shape.

---

## Package architecture

The project should keep the existing pure/effectful separation, but rename the top-level model around games, rounds, and matches.

### Target package layout

```text
internal/
  pure/
    game/
      doc.go
      id.go
      spec.go
      contract.go
      evidence_schema.go
      review_pane.go

    round/
      doc.go
      id.go
      spec.go
      resolved.go
      record.go
      lineage.go
      decision.go
      evidence.go
      next_challenger.go

    match/
      doc.go
      id.go
      spec.go
      record.go
      outcome.go

    policy/
      doc.go
      policy.go
      incumbent.go
      challenger.go
      fingerprint.go
      artifact.go

    execution/
      doc.go
      run.go
      failure.go
      phases.go

    score/
      ...

    report/
      ...

    usage/
      ...

  games/
    codelocalization/
      doc.go
      match.go
      evidence.go
      report.go
      visualization.go
      scoring.go

  app/
    round/
      ...
    game/
      doc.go
      registry.go
      resolve.go

  agents/
    evaluator/
      prompt/
      eino/
        callbacks/

      fake/

    optimizer/
      prompt/
      eino/
      bundle/
      policy/

  adapters/
    bundle/
      fs/
        ...

    config/
      pkl/
        ...

    pipeline/
      exec/
        ...

    scoring/
      pkl/
        ...

  surface/
    cli/
      ...

    console/
      ...
```

### Package decisions

#### `internal/app/round`

Rename to:

```text
internal/app/round
```

Reason:

`round` describes a local composition mechanism.

`round` names the primary contest unit.

The current fake round tests should become round tests.

Example test rename:

```text
TestFakeE2EComposesEvaluatorAndOptimizerAgents
→
TestRoundAdvancesNextChallengerFromEvidence
```

#### `internal/pure/run`

Rename conceptually to:

```text
internal/pure/execution
```

Reason:

The word `run` is overloaded.

A round is the high-level contest unit.

A match execution is one policy executing one match.

Use `execution` for low-level execution records.

Temporary migration rule:

The package may remain `pure/run` during incremental migration, but public docs should refer to these as match executions.

#### `internal/pure/domain/task.go`

Rename conceptually to:

```text
internal/pure/match
```

Reason:

The core review unit is a match.

Dataset adapters may still map external “tasks” into SearchBench matches.

Preferred target names:

```text
MatchSpec → MatchSpec
MatchID → MatchID
TaskInput → MatchInput
TaskOracle → MatchOracle
TaskSlice → MatchSlice
```

#### `internal/pure/report`

Keep package name, but rename exported concepts.

```text
RoundReport → RoundReport
Decision → Decision
DecisionPromote → DecisionPromote
DecisionReview → DecisionReview
DecisionReject → DecisionReject
```

A report belongs to a round, not to a challenger.

#### `internal/pure/score`

Keep package name.

Rename role fields from incumbent/challenger to incumbent/challenger.

```text
RoleIncumbent → RoleIncumbent
RoleChallenger → RoleChallenger
RoleCounts.Incumbent → RoleCounts.Incumbent
RoleCounts.Challenger → RoleCounts.Challenger
MetricEvidence.Incumbent → MetricEvidence.Incumbent
MetricEvidence.Challenger → MetricEvidence.Challenger
```

#### `internal/pure/optimizer`

Keep the package `optimizer` for now because the external role is understandable, but rename the core exported artifact types:

```text
NextChallengerProposal → NextChallengerProposal
NextChallengerArtifact → NextChallengerArtifact
NextChallengerResult → NextChallengerRecord
NextChallengerEvidence → NextChallengerEvidence
```

Reason:

The optimizer is the generator.

The artifact it produces is a next challenger.

#### `internal/pure/domain/pair.go`

Replace generic incumbent/challenger pair usage with named incumbent/challenger pair types where used in round/report/score contexts.

Preferred type:

```go
type PolicyPair struct {
    Incumbent  PolicyRef
    Challenger PolicyRef
}
```

For non-policy generic pairs, keep generic pair if useful.

Do not force all pair usage into the new vocabulary if it is genuinely generic.

---

## App-layer flow

The app layer should present SearchBench as a game-aware round engine.

### Target top-level API

```go
package round

type Input struct {
    ManifestPath string
    BundleRootOverride string
    Deps Deps
}

type Deps struct {
    Games GameRegistry
    Evaluator EvaluatorDeps
    Optimizer OptimizerDeps
    BundleWriter BundleWriter
    ObjectiveRunner ObjectiveRunner
    Now func() time.Time
}

type Record struct {
    Game GameContract
    Round ResolvedRound

    Matches []MatchRecord
    Evidence RoundEvidence
    Objective ObjectiveResult
    Decision Decision

    Bundle RoundBundleRef
    NextChallenger *NextChallengerProposal
}

func Run(ctx context.Context, input Input) (Record, error)
```

### Desired top-level implementation style

The top-level flow should read like this:

```go
func Run(ctx context.Context, input Input) (Record, error) {
    game, err := ResolveGame(ctx, input)
    if err != nil {
        return Record{}, err
    }

    resolved, err := ResolveRound(ctx, game, input)
    if err != nil {
        return Record{}, err
    }

    matches, err := EvaluateMatches(ctx, resolved, input.Deps)
    if err != nil {
        return Record{}, err
    }

    evidence, err := BuildEvidence(game, resolved, matches)
    if err != nil {
        return Record{}, err
    }

    objective, err := EvaluateObjective(ctx, resolved, evidence, input.Deps)
    if err != nil {
        return Record{}, err
    }

    decision := Decide(game, resolved, evidence, objective)

    next, err := ProposeNextChallenger(ctx, resolved, evidence, objective, decision, input.Deps)
    if err != nil {
        return Record{}, err
    }

    bundle, err := WriteBundle(ctx, resolved, evidence, objective, decision, next, input.Deps)
    if err != nil {
        return Record{}, err
    }

    return Record{
        Game:           game,
        Round:          resolved,
        Matches:        matches,
        Evidence:       evidence,
        Objective:      objective,
        Decision:       decision,
        Bundle:         bundle,
        NextChallenger: next,
    }, nil
}
```

This top-level code should avoid framework-shaped request/response ceremony.

Use:

```text
Input
Resolved
Record
Bundle
Proposal
```

Avoid `Request` / `Response` as architectural nouns.

---

## Naming rules

### Use these names

```text
Game
GameSpec
GameContract
GameID
GameKind
ReviewPane
Round
RoundSpec
ResolvedRound
RoundRecord
RoundBundle
RoundBundleRef
RoundLineage
RoundEvidence
RoundReport
Match
MatchSpec
MatchRecord
MatchOutcome
Decision
DecisionRule
IncumbentPolicy
ChallengerPolicy
PolicyPair
NextChallenger
NextChallengerProposal
NextChallengerArtifact
```

### Avoid these names in new high-level code

```text
Incumbent
Challenger
Task
RoundReport
Decision
Round
Projection
Request
Response
Plan
```

### Allowed transitional names

The following can remain temporarily during migration:

```text
incumbent
challenger
task
run
request
result
plan
```

but only in:

* compatibility adapters
* generated Pkl code before schema migration
* old tests not yet migrated
* dataset adapters where external datasets call items tasks
* low-level execution code where `run` means one execution

New docs, reports, CLI output, and visualization output should not introduce these names.

---

## Pkl schema migration

The Pkl configuration should move from round/incumbent/challenger vocabulary to game/round/incumbent/challenger/match vocabulary.

### Current conceptual shape

```pkl
policies {
  incumbent { ... }
  challenger { ... }
}

evaluation {
  incumbent { ... }
  challenger { ... }
}
```

### Target conceptual shape

```pkl
game {
  id = "code-localization"
  kind = "code_localization"
}

round {
  id = "round-002"

  incumbentPolicy {
    id = "jcodemunch-incumbent"
    name = "jCodeMunch incumbent"
    backend = "jcodemunch"
  }

  challengerPolicy {
    id = "iterative-context-challenger-round-002"
    name = "Iterative Context challenger round 002"
    backend = "iterative_context"

    promptBundle {
      name = "graph-lookahead"
      version = "round-002"
    }

    policy {
      id = "next-challenger-round-002"
      path = "policies/challenger_policy.py"
    }
  }

  matches {
    dataset {
      kind = "lca"
      name = "JetBrains-Research/lca-bug-localization"
      config = "py"
      split = "dev"
      maxItems = 20
    }
  }

  objective = "scoring/localization-objective.pkl"

  lineage {
    parentRound {
      id = "round-001"
      bundlePath = "../local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001"
    }
  }
}
```

### Artifact block migration

Current:

```pkl
artifacts {
  challengerPolicyRound001
  challengerPolicyRound002
  parentEvaluationRound001
}
```

Target:

```pkl
artifacts {
  incumbentPolicyRound001
  challengerPolicyRound001
  nextChallengerRound002
  parentRound001Bundle
}
```

### Generated Pkl types

Rename generated concepts after the schema changes:

```text
Round manifest concept is `Round`
ChallengerEvaluationBinding → ChallengerEvaluationBinding
ChallengerUses → ChallengerUses
NextChallengerArtifact → NextChallengerArtifact
CompletedEvaluationBundleArtifact → CompletedRoundBundleArtifact
ParentRun → ParentRound
NextChallengerTarget → NextChallengerTarget
NextChallengerEvidence → NextChallengerEvidence
Dataset.Task → Match
MatchSpec → MatchSpec
```

---

## Bundle architecture

Bundles should become game-scoped round bundles.

### Current bundle path

```text
artifacts/games/code-localization/rounds/round-001/
```

### Target bundle path

```text
artifacts/games/code-localization/rounds/round-001/
```

A shorter local path is acceptable inside a game-specific fixture, but the canonical cross-game shape includes the game id.

### Target files

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
  next-challenger.json?       # optional
  replay-events.json?         # optional
  visualization.json?         # optional derived view
```

### Canonical file decisions

The round bundle uses `resolved-round.json`, `round-report.json`, `round-report.txt`, `evidence.pkl`, `objective.json`, `decision.json`, and `metadata.json`.

### Why `evidence.pkl`

The artifact is more than a single numeric result.

It contains the structured evidence that objective Pkl evaluates.

Use:

```text
evidence.pkl
```

The objective produces:

```text
objective.json
```

The decision produces:

```text
decision.json
```

---

## Report architecture

Reports should be round reports.

### Rename

```text
RoundReport → RoundReport
RoundReport.Spec → RoundReport.Spec
RoundReport.CreatedAt → RoundReport.CreatedAt
RoundReport.Comparisons → RoundReport.Comparisons
RoundReport.Regressions → RoundReport.Regressions
RoundReport.Decision → RoundReport.Decision
```

### Report top-level shape

```go
type RoundReport struct {
    SchemaVersion string
    ReportID ReportID

    GameID GameID
    RoundID RoundID

    IncumbentPolicy PolicyRef
    ChallengerPolicy PolicyRef

    Matches []MatchSummary
    Evidence RoundEvidenceSummary
    Objective *ObjectiveResult
    Decision Decision

    CreatedAt time.Time
}
```

### Report language

Use:

```text
Game: Code Localization
Round: round-002
Decision: PROMOTE CHALLENGER
```

Avoid:

```text
challenger improved over incumbent
```

Prefer:

```text
challenger beat incumbent on localization quality
challenger regressed protected matches
challenger used fewer tokens than incumbent
```

---

## Evidence architecture

`RoundEvidence` is the scoring-facing evidence document.

### Rename

```text
RoundEvidenceDocument → RoundEvidenceDocument
ScoreEvidence → RoundEvidence
BuildRoundEvidence → BuildRoundEvidence
```

### Field rename decisions

```text
policies.incumbent → policies.incumbent
policies.challenger → policies.challenger

tasks → matches
taskId → matchId
taskRuns → matchExecutions

runCounts.incumbent → executionCounts.incumbent
runCounts.challenger → executionCounts.challenger

failureCounts.incumbent → failureCounts.incumbent
failureCounts.challenger → failureCounts.challenger

metrics[].incumbent → metrics[].incumbent
metrics[].challenger → metrics[].challenger

decisionDecision → decision
```

### Pkl evidence target

```pkl
schemaVersion = "searchbench.round_evidence.v1"
gameId = "code-localization"
roundId = "round-002"
reportId = "report-round-002"

policies {
  incumbent { ... }
  challenger { ... }
}

matchCounts {
  total = 20
  improved = 13
  neutral = 5
  regressed = 2
  protectedRegressions = 0
}

executionCounts {
  incumbent = 20
  challenger = 20
}

failureCounts {
  incumbent = 0
  challenger = 1
}

localizationDistance {
  goldHop {
    incumbent = 3.5
    challenger = 1.5
    delta = -2.0
    improved = true
  }
}

usage {
  incumbentTokens = 120000
  challengerTokens = 84000
  tokenDelta = -36000
}

regressions {
  count = 1
  severeCount = 0
  protectedCount = 0
}

decision {
  value = "PROMOTE"
  reason = "challenger improves localization quality and token usage without protected regressions"
}
```

---

## Objective architecture

The objective is the visible scoring rule for a game and round.

The objective should compare:

* current round evidence
* optional parent round evidence
* optional fixed incumbent/challenger deltas
* game-specific evidence

### Current terms

```pkl
current
parent
```

These are acceptable for round lineage.

Do not replace `parent` with incumbent.

`parent` is temporal lineage.

`incumbent` is the policy role inside a round.

They are different.

### Preferred objective variable names

```pkl
local challengerQuality = current.localizationQuality.challenger
local incumbentQuality = current.localizationQuality.incumbent
local challengerVsIncumbent = challengerQuality - incumbentQuality

local parentChallengerQuality = parent?.localizationQuality?.challenger
local improvementVsParent = challengerQuality - parentChallengerQuality
```

### Important distinction

Inside a round:

```text
incumbent vs challenger
```

Across rounds:

```text
parent round vs current round
```

Inside a game:

```text
domain-specific win condition
```

Do not collapse these relationships.

---

## Optimizer architecture

The optimizer reads bounded round evidence and proposes a next challenger.

### Role

The optimizer does not:

* own the decision
* mutate the current round
* inspect denied evidence
* become the source of truth
* decide decision
* define the game

The optimizer does:

* read allowed evidence
* inspect current challenger policy if permitted
* propose a `NextChallenger`
* include rationale and validation metadata
* preserve denied evidence boundaries

### Rename

```text
NextChallengerResult → NextChallengerRecord
NextChallengerProposal → NextChallengerProposal
NextChallengerArtifact → NextChallengerArtifact
NextChallengerTarget → NextChallengerTarget
NextChallengerEvidence → NextChallengerEvidence
```

### Target flow

```text
GameContract
+ RoundEvidence
+ ObjectiveResult
+ Decision
+ current ChallengerPolicy artifact
→ optimizer
→ NextChallengerProposal
```

### Next challenger proposal shape

```go
type NextChallengerProposal struct {
    ProposalID string

    GameID game.ID
    SourceRoundID round.ID

    Target NextChallengerTarget

    Summary string
    Rationale string

    Artifact NextChallengerArtifact

    EvidenceUsed []EvidenceRef
    EvidenceDenied []DeniedEvidenceRef

    Validation NextChallengerValidation
}
```

---

## Visualization architecture

The visualization should use the same product vocabulary as the backend.

Primary view:

```text
Round View
```

Secondary view:

```text
Analysis View
```

### Round View

Round View is the side-by-side explanatory view.

It shows:

```text
Game: Code Localization
Round 002
IncumbentPolicy vs ChallengerPolicy
```

Both policies run on the same match.

Both use game-specific replay panes.

For code localization, both panes use graph-stage replay.

Both panes show how close each policy gets to the localization target.

The winner is highlighted.

A release-report modal shows the `Decision`.

### Analysis View

Analysis View is the operator/debugging view.

It can show:

* single replay pane
* config/evidence pane
* objective values
* trace details
* replay controls
* game-specific evidence details

This mode is for serious optimization work.

### Shared bottom match bar

Both views use the same bottom match status bar.

Cell statuses:

```text
IMPROVED
NEUTRAL
REGRESSED
PENDING
EXCLUDED
```

The UI may color these green/yellow/red/gray, but the data model should use semantic status names.

Each cell is one match.

For code localization, each cell is one localization task.

### Visualization output

Avoid naming the generated type after projection in user-facing docs.

The file can be:

```text
visualization.json
```

The type should be:

```text
RoundVisualization
```

Preferred type:

```ts
type RoundVisualization = {
  game: {
    id: string
    kind: string
    name: string
  }

  round: {
    id: string
    decision: "PROMOTE" | "REVIEW" | "REJECT"
    roundHash?: string
    parentRoundHash?: string
  }

  policies: {
    incumbent: PolicySummary
    challenger: PolicySummary
    nextChallenger?: PolicySummary
  }

  objective: ObjectiveSummary
  matches: MatchRoundSummary
  representativeMatch?: RoundReplayMatch
  releaseReport: ReleaseReportSummary
}
```

---

## CLI architecture

The CLI should expose game- and round-oriented commands.

### Preferred commands

```text
searchbench game list
searchbench game inspect code-localization

searchbench round run
searchbench round inspect
searchbench round report
searchbench round visualize
```

### Avoid

```text
searchbench evaluate
searchbench optimize
searchbench local-e2e
```

These may exist as hidden/internal subcommands during migration, but the public CLI should center games and rounds.

### CLI entrypoint principle

The CLI should not own orchestration.

It should call:

```go
round.Run(ctx, input)
```

The app layer owns the lifecycle.

The CLI parses arguments and renders output.

---

## Test architecture

Tests should be renamed to assert the game/round/match model.

### Important test names

```text
TestGameDefinesMatchEvidenceAndReviewPanes
TestRoundComparesIncumbentAndChallenger
TestRoundRunsMatchesUnderGameContract
TestRoundWritesDurableBundle
TestRoundDecisionPromotesChallenger
TestRoundDecisionRejectsProtectedRegression
TestRoundEvidenceUsesIncumbentChallengerRoles
TestRoundCanProduceNextChallenger
TestRoundLineageReferencesParentRound
TestRoundVisualizationRendersMirroredReplay
TestCodeLocalizationGameBuildsHopDistanceEvidence
```

### Current fake E2E test rename

```text
TestFakeE2EComposesEvaluatorAndOptimizerAgents
→
TestRoundAdvancesNextChallengerFromEvidence
```

The test should prove:

```text
GameContract
→ ResolvedRound
→ incumbent/challenger match executions
→ RoundEvidence
→ ObjectiveResult
→ Decision
→ NextChallengerProposal
→ RoundBundle
```

---

## Migration map

### Concept rename map

```text
Round manifest               → Round
Incumbent                      → Incumbent
Challenger                     → Challenger
IncumbentPolicy                → IncumbentPolicy or IncumbentPolicy
ChallengerPolicy               → ChallengerPolicy or ChallengerPolicy
Task                          → Match
MatchSpec                      → MatchSpec
MatchID                        → MatchID
MatchExecution                       → MatchExecution
RoundReport               → RoundReport
Decision             → Decision
NextChallengerProposal                → NextChallengerProposal
NextChallengerArtifact        → NextChallengerArtifact
OptimizationResult            → NextChallengerRecord
NextChallengerEvidence          → NextChallengerEvidence
NextChallengerTarget            → NextChallengerTarget
ParentRun                     → ParentRound
CompletedEvaluationBundle     → CompletedRoundBundle
local e2e                     → round
round bundle                    → round bundle
round evidence                → round evidence
round evidence artifact       → evidence.pkl
resolved round artifact        → resolved-round.json
round report JSON artifact     → round-report.json
round report text artifact     → round-report.txt
```

### Package rename map

```text
internal/app/round         → internal/app/round
internal/pure/run             → internal/pure/execution
internal/pure/domain/task.go  → internal/pure/match
internal/pure/report          → keep, but rename exported types
internal/pure/score           → keep, but rename role fields
internal/pure/optimizer       → keep temporarily, but rename exported proposal types
internal/games/codelocalization → add for first game-specific evidence/review code
```

### Pkl rename map

```text
SearchBench round schema       → SearchBenchRound.pkl
Round manifest               → Round
dataset/task slice            → matches
policies.incumbent              → policies.incumbent
policies.challenger             → policies.challenger
evaluation.incumbent           → evaluation.incumbent
evaluation.challenger          → evaluation.challenger
challengerPolicyRound001       → challengerPolicyRound001
challengerPolicyRound002       → nextChallengerRound002
parentEvaluationRound001      → parentRound001Bundle
```

### Artifact path map

```text
artifacts/games/code-localization/rounds/               → artifacts/games/<game-id>/rounds/
example-round-001/            → round-001/
```

---

## Layering rules

### Pure packages

Pure packages may define:

* game contracts
* match specs
* round specs
* policy refs
* evidence models
* report models
* score models
* decision rules
* lineage refs
* next-challenger proposal models

Pure packages must not depend on:

* filesystem
* Pkl evaluator
* Eino
* MCP
* LangSmith
* model providers
* CLI frameworks
* network clients

### Game packages

Game packages may define:

* domain-specific match schemas
* evidence builders
* match outcome classifiers
* report sections
* visualization/replay payloads
* review pane definitions

Game packages should not own:

* round orchestration
* model provider execution
* bundle filesystem writes
* tracing integrations

### App packages

App packages coordinate:

* resolving game contracts
* resolving manifests
* executing matches
* building evidence
* evaluating objectives
* deciding outcomes
* writing bundles
* proposing next challengers

App packages may depend on ports and adapters through explicit interfaces.

### Adapter packages

Adapters own:

* Pkl loading/evaluation
* bundle filesystem writes
* model provider calls
* Eino execution
* MCP lifecycle
* LangSmith callbacks
* dataset/repo materialization

Adapters should not define core game/round semantics.

### Surface packages

Surface packages own:

* CLI parsing
* console rendering
* human output
* command wiring

Surface packages should not own round orchestration.

---

## Architectural invariants

### 1. A game defines the domain contract

The game decides what evidence matters and what review panes are meaningful.

The round executes under that contract.

### 2. A round is immutable after completion

Once `COMPLETE` exists in a round bundle, the bundle must not be mutated.

A later change creates a new round.

### 3. A match is the smallest reviewable unit

Every dataset item should become a match record.

The bottom visualization bar should be able to navigate matches.

### 4. Agents generate artifacts; SearchBench owns judgment

Evaluator agents generate predictions.

Optimizer agents generate next-challenger proposals.

SearchBench owns:

* evidence construction
* objective evaluation
* decisions
* lineage
* bundles

### 5. The UI renders evidence

The visualization must not invent evaluation state.

It renders:

* round bundles
* match replay events
* round visualization JSON
* report artifacts

### 6. Parent round and incumbent policy are different concepts

A parent round is temporal lineage.

An incumbent policy is the policy role inside the current round.

Do not collapse these.

### 7. Game and round are different concepts

A game is a domain contract.

A round is one contest under that contract.

Do not collapse these.

### 8. The decision is explicit

Every completed round should have a decision artifact.

Even if scoring fails, the round should expose a failed/incomplete state clearly.

### 9. Optimizer output points forward

Optimizer output is always a possible future challenger.

It is not a mutation of the current challenger.

---

## Non-goals

Do not turn SearchBench into:

* a generic workflow engine
* a generic DAG scheduler
* a hosted eval platform
* a trace viewer clone
* a prompt optimizer only
* an agent framework
* a live dashboard-first system
* a vague domain-agnostic abstraction framework

The project-specific spine is enough:

```text
Game
→ Round
→ Match
→ Evidence
→ Decision
→ NextChallenger
```

---

## Migration plan

### Phase 1: Documentation and terminology

* Maintain the architecture spine at `docs/architecture/architecture.md`.
* Keep `AGENTS.md` aligned with Game / Round / Match vocabulary at the repo root.
* Keep `docs/architecture/visualization.md` aligned with Game, Round View, Analysis View, and Match bar language.
* Add glossary for old/new terms.

### Phase 2: App-level rename

* Rename `internal/app/round` to `internal/app/round`.
* Rename app-level `Request`/`Result` to `Input`/`Record`.
* Rename `Plan` to `ResolvedRound` where applicable.
* Update fake E2E tests to round tests.

### Phase 3: Game model introduction

* Introduce `internal/pure/game`.
* Introduce `internal/games/codelocalization`.
* Define `GameContract`.
* Define the first game: `CodeLocalizationGame`.
* Route current localization-specific evidence through that game package.

### Phase 4: Match model introduction

* Introduce `internal/pure/match`.
* Rename task-facing core types to match-facing types.
* Preserve adapter compatibility for external datasets that call items tasks.

### Phase 5: Pure model rename

* Introduce/complete `internal/pure/round`.
* Rename `RoundReport` to `RoundReport`.
* Rename `Decision` to `Decision`.
* Rename role constants to incumbent/challenger.
* Add compatibility aliases only if needed to keep migration manageable.

### Phase 6: Bundle/schema rename

* Rename bundle request/report fields.
* Move output from `artifacts/games/code-localization/rounds` to `artifacts/games/<game-id>/rounds`.
* Rename `evidence.pkl` to `evidence.pkl`.
* Add `decision.json`.
* Update generated fixtures.

### Phase 7: Pkl schema migration

* Use `SearchBenchRound.pkl` as the round schema.
* Add top-level `game`.
* Replace incumbent/challenger schema fields with incumbent/challenger.
* Replace task slice language with match slice language.
* Regenerate Pkl bindings.
* Update example configs.

### Phase 8: Optimizer rename

* Rename optimizer artifacts to next-challenger artifacts.
* Ensure optimizer outputs point to a future round.
* Update optimizer prompt inputs and finalizer types.

### Phase 9: Visualization output

* Define `RoundVisualization`.
* Generate `visualization.json` from round bundles.
* Support side-by-side incumbent/challenger replay.
* Use the shared match status bar.
* Keep code localization graph replay as the first game-specific review pane.

---

## Acceptance criteria

The migration is successful when a new reader can understand the project from these nouns alone:

```text
Game
Round
Match
IncumbentPolicy
ChallengerPolicy
Evidence
Decision
NextChallenger
```

The code is successful when the top-level app flow reads as:

```text
resolve game
resolve round
evaluate incumbent/challenger matches
build evidence
evaluate objective
decide
propose next challenger
write round bundle
```

The artifacts are successful when a completed directory clearly represents one immutable round under one game:

```text
artifacts/games/code-localization/rounds/round-002/
  COMPLETE
  resolved-round.json
  round-report.json
  round-report.txt
  evidence.pkl
  objective.json
  decision.json
  metadata.json
  next-challenger.json
```

The visualization is successful when a viewer can understand this in under ten seconds:

```text
this is the game
the incumbent defended its position
the challenger tried to replace it
each dataset item is a match
the round produced evidence
the decision explains who won
the next challenger points forward
```

---

## One-line thesis

SearchBench turns AI model changes into reviewable games.

## Short product description

SearchBench defines a `Game`, runs an `IncumbentPolicy` against a `ChallengerPolicy` across dataset `Matches`, writes the evidence into an immutable `RoundBundle`, produces a release `Decision`, and can use that evidence to generate the `NextChallenger`.

## Load-bearing sentence

Agents generate challenger artifacts.

SearchBench owns the game, the rounds, and the evidence-backed judgment of whether those artifacts survive.
