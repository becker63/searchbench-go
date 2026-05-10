# Future Integrations

SearchBench-Go should keep a strict separation between the pure SearchBench model, app orchestration, narrow external adapters, and user/developer-facing surfaces.

The goal is not maximum theoretical pluggability. The goal is to keep the Game/Round/Match model pure while making it obvious where real work happens.

```text
pure      stable SearchBench model
ports     project-owned contracts
app       orchestration and composition
adapters  concrete external/runtime boundaries
surface   CLI and human presentation
```

## Current Shape

The current package structure is:

```text
internal/
  pure/
  ports/
  app/
  adapters/
  surface/
  testing/
```

The intended dependency direction is:

```text
surface / app / adapters
    ->
ports
    ->
pure
```

`internal/pure` owns SearchBench vocabulary. `internal/adapters` translates external APIs, subprocesses, runtimes, and persistence into typed SearchBench values. `internal/surface` renders or commands the system; it does not own the model.

## Pure Model

Pure packages should define stable concepts:

```text
internal/pure/game
internal/pure/round
internal/pure/match
internal/pure/domain
internal/pure/policy
internal/pure/execution
internal/pure/score
internal/pure/report
internal/pure/optimizer
internal/pure/codegraph
internal/pure/prompts
internal/pure/usage
```

These packages may define value types, deterministic reducers, prompt contracts, and validation rules. They should not know about concrete SDKs, runtimes, filesystems, subprocesses, or tracing platforms.

## App Layer

Concrete SearchBench workflows live under `internal/app`.

Current app packages include:

```text
internal/app/round
internal/app/evaluation
internal/app/compare
internal/app/optimizer
internal/app/logging
```

The app layer composes pure values, ports, and adapters. It may orchestrate a round, resolve a Pkl manifest, compare policies, build evidence, or run the next-challenger proposal flow.

The app layer should not become the owner of core vocabulary. If a concept is stable across adapters and surfaces, move it inward.

## Adapter Layer

Adapters isolate concrete boundaries:

```text
internal/adapters/config/pkl
internal/adapters/scoring/pkl
internal/adapters/bundle/fs
internal/adapters/executor/eino
internal/adapters/pipeline/exec
```

Future adapters may include:

```text
internal/adapters/dataset/lca
internal/adapters/trace/langsmith
internal/adapters/backend/iterativecontext
internal/adapters/backend/jcodemunch
internal/adapters/providers/openai
```

Adapters may be messy. Their job is to translate the outside world into typed SearchBench values without letting SDK DTOs flow inward.

## Surface Layer

Surface packages expose SearchBench to humans and developers:

```text
internal/surface/cli
internal/surface/console
```

Surface code may render `report.RoundReport`, route commands, and display logs. It must not define artifact formats, scoring semantics, decision rules, or execution contracts.

## Integration Rules

Do not add a generic plugin system just because future integrations are likely.

Do not add a generic model-provider abstraction until more than one real provider path requires it.

Do not let adapter DTOs become SearchBench model types.

Do not create parallel report, score, match, policy, or execution models in adapters.

Do not move SDK-specific types into `internal/pure`.

## LangSmith

LangSmith can be useful as an external trace and review platform, but SearchBench remains authoritative for scoring, evidence, reports, and decisions.

LangSmith may consume:

```text
report.RoundReport
execution.ExecutedRun
execution.RunFailure
score.ScoreSet
score.RoundEvidenceDocument
domain.MatchSpec
domain.SystemRef
```

LangSmith should not define:

```text
RoundReport
RoundEvidence
Decision
NextChallenger
```

If LangSmith integration becomes substantial, place it under an adapter boundary such as `internal/adapters/trace/langsmith` unless the package is only formatting already-owned SearchBench artifacts.

## Dataset Integration

Dataset loading should translate external records into SearchBench matches before crossing inward.

The SearchBench-facing output should be:

```text
[]domain.MatchSpec
domain.NonEmpty[domain.MatchSpec]
```

External datasets may use their own words for rows or work items. Keep that vocabulary at the adapter boundary and project it into `MatchSpec` before app/pure code sees it.

## Graphing

Tree-sitter and other source-code indexing libraries are implementation details. Graphing code should produce `codegraph.Graph` or related pure values without leaking parser-specific DTOs into scoring or reporting.

## Artifact Storage

Filesystem persistence belongs at the adapter edge. Round bundles are SearchBench artifacts, but the filesystem implementation is not part of the pure model.

Canonical round bundles use:

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
```

Optimizer bundles may additionally write a next-challenger artifact and `resolved-next-challenger.json`.

## Decisions

Decision logic is SearchBench-native. It should reduce comparisons, regressions, objective results, and failures into `report.Decision`.

Decision rules may consider:

```text
primary metric improvement
regression count
protected match regressions
challenger failures
cost increases
token efficiency
composite score
```

Do not put complex decision policy directly into presentation or adapter packages.

## Guiding Principle

```text
External systems change.
SearchBench vocabulary should not.
```

Keep unstable integrations isolated. Keep the model pure.
