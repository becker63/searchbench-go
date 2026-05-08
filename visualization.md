# SearchBench Visualization Plan

## Current status

The “cool right pane” already exists.

It is currently a full-page graph proof-trace demo, but that is mostly a styling/layout issue. The core interaction is already built:

- GraphEvent-driven replay
- React Flow graph stage
- Framer Motion edge and node animation
- RxJS event pacing
- Jotai projection into UI state
- deterministic fixtures
- panning/follow behavior
- token/context visual hints
- tests around layout, panning, and animation behavior

So the visualization plan should not treat the right pane as a new system.

The plan is to embed the existing graph-stage demo as the right pane inside a broader SearchBench evaluation visualization.

## Core product idea

The real value of SearchBench is not only showing an agent searching code.

The real value is showing that a candidate prompt can be tested against a baseline across a dataset, using deterministic tooling, and reduced into a clear release decision.

The visualization should communicate this:

    one prompt change
    plus deterministic tooling
    plus dataset-wide evaluation
    equals an inspectable release decision

The homepage version should make this obvious to business-facing visitors.

The deeper case-study version can expose more technical detail.

## High-level layout

The final visualization should have three major regions.

### 1. Left pane: evaluation surface

The left pane shows the controlled change and the scoring logic.

It should answer:

    What changed?
    What stayed constant?
    How are we judging it?

It should include:

- baseline system
- candidate system
- candidate prompt or policy fingerprint
- dataset name/config/split
- objective summary
- score intermediates
- token/cost summary
- promotion/review/reject rule
- lineage: parent run and current run when available

This pane should feel like a polished projection of the run artifacts, not a raw config dump.

It can visually reference PKL, but it should present the meaning first.

### 2. Right pane: existing proof-trace demo

The right pane is the already-built graph-stage visualization.

Its job is to provide the emotional/intuitive proof of how the system works on one representative task.

It should show:

- task prompt or issue
- anchors
- graph expansion
- pending nodes
- resolved evidence
- context inclusion
- token movement
- causal edge animations

This is the “look, the system is actually doing structured work” pane.

Important: this pane does not need to be live-streamed in the homepage version.

It can replay from static GraphEvent JSON.

The existing full-page demo becomes a component inside the larger evaluation layout.

### 3. Bottom pane: dataset result grid

The bottom pane spans the full width below the left and right panes.

This is the most important business-facing piece.

It should look like a GitHub status / uptime grid.

Each cell represents one dataset item.

Cell semantics:

    green  = candidate improved over baseline
    yellow = tied, ambiguous, or within review threshold
    red    = candidate regressed
    gray   = pending, excluded, or not run

Protected regressions should be visually distinct.

Possibilities:

- red cell with stronger border
- red cell with warning glyph
- separate protected-case lane
- tooltip explaining why it matters

This pane answers:

    Did the candidate prompt actually hold up across the dataset?

## Final decision surface

Above or beside the grid, show a compact summary.

Example fields:

- Decision: PROMOTE / REVIEW / REJECT
- Improved: N
- Neutral: N
- Regressed: N
- Protected regressions: N
- Average score delta
- Token delta
- Cost delta

The decision rule should be visible in plain language.

Example:

    Promote only if average score improves, token cost stays within budget,
    regression count stays under threshold, and protected regressions are zero.

This is the part that makes the tool feel governable instead of magical.

## Architecture

The visualization should be static/projection-first.

Do not build a live dashboard first.

The right pane can animate, but the source of truth should be static artifacts.

Flow:

    SearchBench run bundle
      -> visualization projection
      -> visualization.json
      -> website renderer
      -> optional video/animation renderer

The UI should not become another source of truth.

The UI renders the evidence.

## Artifact inputs

The projection should be built from the run bundle.

Expected inputs:

- metadata.json
- resolved.json
- report.json
- objective.json
- score.pkl
- optional trace-events.json
- optional representative-task trace

The exact artifact names can evolve, but the important rule is:

    the visualization renders the run bundle;
    it does not invent evaluation state.

## Projection output

The first milestone should be a single presentation-facing JSON file.

Suggested output:

    visualization.json

Minimum contents:

- run id
- run hash
- parent run hash if available
- decision
- baseline summary
- candidate summary
- prompt/policy summary
- dataset summary
- per-item cells
- aggregate counts
- objective values
- representative trace events for the right pane

The website should consume only this projection.

## Suggested projection shape

type VisualizationProjection = {
  run: {
    id: string
    createdAt?: string
    decision: "PROMOTE" | "REVIEW" | "REJECT"
    runHash?: string
    parentRunHash?: string
  }

  systems: {
    baseline: SystemSummary
    candidate: SystemSummary
  }

  promptChange?: {
    baselineFingerprint?: string
    candidateFingerprint?: string
    summary: string
  }

  objective: {
    finalScore: number
    values: ObjectiveValue[]
    promotionRule: PromotionRuleSummary
  }

  dataset: {
    name: string
    config?: string
    split?: string
    itemCount: number
    cells: DatasetCell[]
    summary: DatasetSummary
  }

  representativeTrace?: {
    taskId: string
    title?: string
    prompt?: string
    events: GraphEvent[]
  }
}

type DatasetCell = {
  id: string
  index: number
  status: "improved" | "neutral" | "regressed" | "pending" | "excluded"
  protected?: boolean

  baselineScore?: number
  candidateScore?: number
  delta?: number

  baselineTokens?: number
  candidateTokens?: number
  tokenDelta?: number

  reason?: string
}

## Relationship to the existing right pane

The existing right-pane demo should become a reusable component.

Current role:

    full-page proof-trace demo

Future role:

    representative-trace pane inside SearchBench evaluation hero

The component should accept a trace/event list and render the same proof-trace behavior it already supports.

The work is mostly:

- layout adaptation
- prop/interface cleanup
- projection JSON input
- styling integration with the two-pane + bottom-grid layout
- making sure the graph remains legible at pane size

The right pane should not be redesigned from scratch.

## Homepage use

Homepage headline options:

    Making AI workflows inspectable

or:

    From one prompt change to a release decision

or:

    Test whether a prompt actually beats the baseline

Suggested subhead:

    SearchBench compares a candidate prompt against a baseline across a dataset,
    then turns traces, scores, regressions, and token costs into a decision surface
    teams can inspect.

The homepage animation should show:

    prompt change
    -> representative trace
    -> dataset cells filling in
    -> final decision

## Animation sequence

The motion should be explanatory, not the source of truth.

Suggested sequence:

1. Left pane shows baseline and candidate.
2. Candidate prompt/policy becomes highlighted.
3. Right pane replays one representative GraphEvent trace.
4. Bottom grid starts gray.
5. Cells fill in green/yellow/red as dataset results “complete.”
6. Summary counters update.
7. Objective values resolve.
8. Final decision appears.

This can all be driven from static visualization.json.

No streaming required.

## Why static-first matters

The strongest architecture is:

    the run is static
    the evidence is static
    the projection is static
    the motion is explanation

That matches the SearchBench thesis.

Truth should live in inspectable artifacts, not inside a running dashboard.

## What not to build first

Do not start with:

- a generic live dashboard
- mandatory streaming
- a frontend event bus beyond what already exists for the graph replay
- trace-provider lock-in
- UI-owned evaluation state
- hand-authored one-off animation data
- a second graph event protocol

The existing graph-event protocol should be reused for the right pane.

The dataset grid should come from per-item evaluation results.

The left pane should come from the run/objective/config artifacts.

## Required backend/schema gap

The biggest likely backend gap is per-item comparison data.

The bottom grid requires report data at dataset-item granularity.

The report should preserve, per item:

- item id
- baseline score
- candidate score
- score delta
- baseline token count
- candidate token count
- token delta
- status classification
- protected-case flag if applicable
- short reason or failure/regression summary

Without this, the UI can show aggregate score but not the dataset-wide status map.

## First milestone

Build a static visualization fixture.

Input:

    one completed run bundle

Output:

    visualization.json

Minimum useful fields:

- run id
- baseline id
- candidate id
- final decision
- final score
- dataset cells
- aggregate counts
- representative GraphEvent trace

Then render a static page from that JSON.

## Second milestone

Embed the existing graph-stage demo as the right pane.

Tasks:

- extract the current full-page demo into a reusable pane component
- feed it representativeTrace.events from visualization.json
- keep the existing animation/replay behavior
- tune sizing so it works inside a pane
- avoid major graph redesign

## Third milestone

Build the dataset grid.

Tasks:

- render one cell per dataset item
- support green/yellow/red/gray statuses
- show aggregate counts
- show final decision
- add hover/click detail later, but not in v1

## Fourth milestone

Add homepage polish.

Tasks:

- integrate left pane, right pane, and bottom grid into the site hero
- add short business-facing copy
- make the animation loop cleanly
- optionally export the same flow as a video

## Acceptance criteria

The visualization succeeds when a non-specialist can understand this in under ten seconds:

1. A candidate prompt was tested against a baseline.
2. It was tested across many tasks, not one cherry-picked demo.
3. The graph pane shows how one representative run works.
4. The bottom grid shows whether the candidate generalized.
5. The final decision comes from inspectable artifacts, not vibes.

## One-line thesis

SearchBench turns prompt changes into release decisions.
