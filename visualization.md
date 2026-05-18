# SearchBench Visualization Plan

**Related docs:** [Architecture](./docs/architecture.md)

## Current status

The "SAT" / graph-stage visualization already exists.

It is currently a full-page graph proof-trace demo, but that is mostly a layout and product-shape issue. The core interaction is already built:

- GraphEvent-driven replay
- React Flow graph stage
- Framer Motion edge and node animation
- RxJS event pacing
- Jotai projection into UI state
- deterministic fixtures
- panning/follow behavior
- token/context visual hints
- tests around layout, panning, and animation behavior

So this plan should not treat the graph stage as a new system.

The real work now is to place it inside a broader SearchBench product story.

---

## Core product idea

The real value of SearchBench is not only showing an agent searching code.

The real value is showing that a new policy can challenge the current one, be evaluated on the same task conditions, and be reduced into a clear release decision.

The visualization should communicate this:

    IncumbentPolicy
    vs
    ChallengerPolicy
    on the same task
    across the same dataset
    reduced into a Decision
    and, optionally, a NextChallenger

This should feel like a round-based comparison system, not just a trace viewer.

---

## Two visualization products

There are really two related visualization products here.

### 1. Round View

This is the explanatory / cinematic / homepage / video-friendly view.

Its job is to make the system instantly legible.

It should answer:

    What is competing?
    What changed?
    Why did one side win?
    Should the challenger advance?

This is the side-by-side "race" view.

### 2. Analysis View

This is the serious operator / optimization workbench.

Its job is to help someone tune prompts, policies, and retrieval behavior.

It should answer:

    What is happening inside this one run?
    What changed in the config?
    How can I improve the challenger?

This is the single-pane, more technical view.

Both modes should share a common projection model and a common replay/status bar.

---

# Round View

## High-level concept

Round View should lean hard into the competitive story.

Two mirrored panes sit side by side:

- left = IncumbentPolicy
- right = ChallengerPolicy

Both receive the same task.
Both run at the same time.
Both navigate the same code universe independently.
The viewer sees how each one behaves and how close each one gets to the localization target.

At the end of the round, SearchBench overlays a release-report modal explaining the outcome.

This should feel like a polished native app experience, even if it is implemented on the web.

Not a dashboard first.

A comparison viewer first.

## Core message

The Round View should communicate:

    the current policy defends its position
    the challenger tries to replace it
    both are evaluated under the same conditions
    the evidence produces a Decision

---

## Round View layout

### 1. Match header

A compact top header should establish the round immediately.

Fields:

- Round ID / number
- task title or issue summary
- IncumbentPolicy name
- ChallengerPolicy name
- current status:
  - LIVE
  - REPLAY
  - EVALUATING
  - DECISION READY

Possible visual shape:

    Round 002
    IncumbentPolicy: jCodeMunch
    vs
    ChallengerPolicy: Iterative Context + Lookahead

This should feel like a match card.

### 2. Side-by-side mirrored panes

The center of the UI is the most important part.

There should be two mirrored panes:

- left pane = incumbent
- right pane = challenger

Both panes should use the existing SAT / graph-stage interaction.

Each pane should show the same categories of information so that the comparison is immediate:

- task / issue being localized
- current anchor(s)
- graph expansion
- explored nodes
- pending nodes
- resolved evidence
- chosen context
- token movement
- tool-call behavior

The panes should be structurally similar, not two unrelated views.

The viewer should feel like they are watching two policies compete in the same arena.

### 3. Visual closeness to the target

A key addition:

show how close each side is to the correct localization target.

This is important because it turns the scoring logic into something visible.

Possible ways to express it:

- lay out the graph/tree so the target region exists in a visible spatial location
- pan the view over the code graph to show how far each system is from the correct neighborhood
- visually indicate hop distance to the relevant nodes/files
- show explored paths diverging toward or away from the correct cluster
- highlight correct target nodes/files once the replay completes, or faintly reveal them as a reference layer

This is where the explanatory power really increases.

Instead of only saying:

    challenger improved hop distance

the viewer can see:

    challenger moved into the correct neighborhood
    incumbent wandered elsewhere or stopped early

### 4. Live comparison ribbon

A comparison strip should sit between or below the panes.

It should update live during the replay and make the "race" feeling explicit.

Possible fields:

- token usage
- tool calls
- current hop distance
- files localized
- localization quality
- current lead / trailing status

Example shape:

    Tokens:        12.4k   vs   8.9k
    Files found:   4/6     vs   5/6
    Hop distance:  3.1     vs   1.8
    Status:        trailing vs leading

This is not the final score.

It is the live comparison layer that helps the user read the match.

### 5. Final release-report modal

When both runs complete, the visualization should freeze into a judgment moment.

At that point:

- the winning side gets a green highlight or glow
- the losing side becomes visually quieter
- the background dims slightly
- a release-report modal overlays the content

This modal is one of the most important product surfaces.

It should not feel like a generic pop-up.

It should feel like a release judgment.

Fields:

- Decision: PROMOTE / REVIEW / REJECT
- IncumbentPolicy
- ChallengerPolicy
- summary of what changed
- what improved
- what regressed
- average score delta
- token delta
- cost delta
- protected regressions
- plain-language decision rule

Example:

    Decision: PROMOTE CHALLENGER

    Why:
    - better localization quality
    - lower token usage
    - no protected regressions

    Margin:
    +0.12 final score
    -18% token cost

If the round is part of an evolving series, the modal can also optionally show:

    Challenger promoted
    -> becomes next round's IncumbentPolicy

That makes the temporal spine visible.

---

# Analysis View

## High-level concept

The Analysis View is the more technical, workbench-like interface.

It is for users who want to actually optimize prompts, policies, and retrieval behavior.

This mode should not be forced to serve the homepage story.

It can be denser and more operational.

## Core message

The Analysis View should communicate:

    here is one run
    here is the controlled configuration
    here is how the policy behaved
    here is the score/evidence you would use to improve it

## Analysis View layout

This mode can use a single main graph pane plus supporting context.

Suggested regions:

### 1. Main graph pane

Use the existing SAT / graph-stage demo as the primary focus.

### 2. Config / evidence pane

A left or side pane can show:

- IncumbentPolicy summary
- ChallengerPolicy summary
- what changed
- task slice / dataset summary
- objective values
- score/evidence summary
- lineage / prior round
- policy fingerprint / prompt fingerprint

This should present meaning first, not raw config first.

It can visually reference PKL, but should not be a raw config dump.

### 3. Event / trace detail

Optional operator-focused detail:

- current tool call
- current context set
- event log
- symbol/file details
- selected node metadata

This mode is where serious users inspect and improve behavior.

---

# Shared bottom status bar

Both Round View and Analysis View should have a shared bottom bar.

This should feel like a Git status / CI / replay strip.

It is a core unifying surface.

## Purpose

The bottom bar should let the user:

- browse all dataset items
- replay specific runs
- see outcome distribution at a glance
- navigate red/yellow/green cases
- jump directly to interesting regressions

## Semantics

Each cell represents one dataset item.

Cell semantics:

- green  = challenger improved
- yellow = tie / ambiguous / review threshold
- red    = challenger regressed
- gray   = pending / excluded / not run

Protected regressions should be visually distinct.

Possibilities:

- stronger red border
- warning glyph
- separate lane
- tooltip explaining why the case is protected

This bar is important because it makes the system feel like:
- a dataset viewer
- a replay chooser
- a release engineering surface
- a durable evaluation artifact

---

# Product language

The visualization should consistently use the following names:

- IncumbentPolicy
- ChallengerPolicy
- Round
- Decision
- NextChallenger

These names should appear in:

- UI copy
- visualizations
- reports
- docs
- narration
- future videos

Avoid centering older language like:

- incumbent
- challenger
- projection

Those terms can still exist internally where useful, but the public/explanatory product language should be temporal and story-shaped.

---

# Architecture

The visualization should remain static/projection-first.

Do not build a live dashboard first.

The graph panes can animate, but the source of truth should still be static artifacts and replayable event streams.

Flow:

    SearchBench round bundle
      -> visualization projection
      -> visualization.json
      -> website renderer
      -> optional video/animation renderer

The UI should not become another source of truth.

The UI renders durable evidence.

---

# Artifact inputs

The visualization should be built from the round bundle.

Expected inputs:

- metadata.json
- resolved-round.json
- round-report.json
- round-report.txt
- objective.json
- evidence.pkl
- optional trace-events.json
- optional representative-match trace
- optional per-side graph replay events for incumbent and challenger

Important rule:

    the visualization renders the round bundle;
    it does not invent evaluation state

---

# Projection output

The first milestone should still be a single presentation-facing JSON file.

Suggested output:

    visualization.json

But the shape should now reflect the round-based comparison model.

Minimum contents:

- round id
- round hash
- parent round hash if available
- Decision
- IncumbentPolicy summary
- ChallengerPolicy summary
- what changed
- dataset summary
- per-item result cells
- aggregate counts
- objective values
- replay events for both sides
- representative task metadata
- release-report summary

The website should consume only this projection.

---

# Suggested projection shape

```ts
type VisualizationProjection = {
  round: {
    id: string
    createdAt?: string
    decision: "PROMOTE" | "REVIEW" | "REJECT"
    roundHash?: string
    parentRoundHash?: string
  }

  policies: {
    incumbent: PolicySummary
    challenger: PolicySummary
    nextChallenger?: PolicySummary
  }

  change?: {
    summary: string
    incumbentFingerprint?: string
    challengerFingerprint?: string
  }

  objective: {
    finalScore: number
    values: ObjectiveValue[]
    decisionRule: string
  }

  dataset: {
    name: string
    config?: string
    split?: string
    itemCount: number
    cells: DatasetCell[]
    summary: DatasetSummary
  }

  representativeTask?: {
    taskId: string
    title?: string
    prompt?: string

    incumbentReplay?: {
      events: GraphEvent[]
    }

    challengerReplay?: {
      events: GraphEvent[]
    }

    target?: {
      filePaths?: string[]
      nodeIds?: string[]
      hopAnnotations?: HopAnnotation[]
    }
  }

  releaseReport: {
    summary: string
    improvedCount: number
    neutralCount: number
    regressedCount: number
    protectedRegressionCount: number
    averageScoreDelta?: number
    tokenDelta?: number
    costDelta?: number
    recommendation?: string
  }
}

type DatasetCell = {
  id: string
  index: number
  status: "IMPROVED" | "NEUTRAL" | "REGRESSED" | "PENDING"
  protected?: boolean
  scoreDelta?: number
  tokenDelta?: number
}

type PolicySummary = {
  id: string
  name: string
  backend?: string
  promptBundleName?: string
  promptBundleVersion?: string
  policyId?: string
  summary?: string
}

type ObjectiveValue = {
  name: string
  kind: "intermediate" | "penalty" | "final"
  value: number | string | boolean
}

type DatasetSummary = {
  improved: number
  neutral: number
  regressed: number
  pending: number
  protectedRegressions: number
}

type HopAnnotation = {
  nodeId: string
  hopDistance: number
}
````

---

# Implementation priorities

## Milestone 1: Projection and replay foundation

* define the new round-oriented visualization projection
* adapt current graph-stage data into a reusable replay payload
* support two independent replay streams in one layout
* build the shared bottom status bar
* render a static Decision summary

## Milestone 2: Round View

* side-by-side mirrored incumbent/challenger panes
* live comparison ribbon
* visual closeness / hop-distance cues
* winner highlight
* release-report modal
* polished animation / match-like flow

## Milestone 3: Analysis View

* single-pane serious-user view
* config/evidence side panel
* more detailed trace inspection
* run replay and case switching from the shared bottom bar

## Milestone 4: Homepage and video usage

* use Round View as the main explanatory artifact
* use Analysis View for deeper technical demos
* ensure both read clearly in narration and recording
* ensure the final flow communicates:

  * what changed
  * why the challenger won or lost
  * whether it should advance

---

# Final product thesis

The visualization should make SearchBench feel like a round-based comparison system for agentic code-search policies.

Not just:

```
here is an agent trace
```

But:

```
here is an IncumbentPolicy
here is a ChallengerPolicy
here is the Round where they compete
here is the Decision
and here is the NextChallenger
```

That is the story the visualization should tell.
