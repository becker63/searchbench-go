package config

import (
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/artifactkind"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/backend"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/nextchallengerevidencekind"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/optimizerdeniedevidencekind"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/provider"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/reportformat"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/runmode"
)

// RoundSpec is the typed root manifest resolved from SearchBenchRound.pkl.
type RoundSpec = generated.SearchBenchRound

type Game = generated.Game
type Dataset = generated.Dataset
type Interfaces = generated.Interfaces
type Interface = generated.Interface
type Policies = generated.Policies
type System = generated.System
type PromptBundle = generated.PromptBundle
type Runtime = generated.Runtime
type Artifacts = generated.Artifacts
type PolicyArtifact = generated.PolicyArtifact
type NextChallengerArtifact = generated.NextChallengerArtifact
type CompletedRoundBundleArtifact = generated.CompletedRoundBundleArtifact
type Agents = generated.Agents
type AgentToolPolicy = generated.AgentToolPolicy
type Model = generated.Model
type AgentBounds = generated.AgentBounds
type Evaluator = generated.Evaluator
type Optimizer = generated.Optimizer
type RetryPolicy = generated.RetryPolicy
type Tracing = generated.Tracing
type Evaluation = generated.Evaluation
type EvaluationSystemBinding = generated.EvaluationSystemBinding
type ChallengerEvaluationBinding = generated.ChallengerEvaluationBinding
type ChallengerUses = generated.ChallengerUses
type Scoring = generated.Scoring
type Report = generated.Report
type Optimization = generated.Optimization
type ParentRound = generated.ParentRound
type NextChallengerTarget = generated.NextChallengerTarget
type NextChallengerEvidence = generated.NextChallengerEvidence

type RunMode = runmode.RunMode
type Backend = backend.Backend
type Provider = provider.Provider
type ReportFormat = reportformat.ReportFormat
type ArtifactKind = artifactkind.ArtifactKind
type NextChallengerEvidenceKind = nextchallengerevidencekind.NextChallengerEvidenceKind
type OptimizerDeniedEvidenceKind = optimizerdeniedevidencekind.OptimizerDeniedEvidenceKind

const (
	ModeEvaluation   = runmode.Evaluation
	ModeOptimization = runmode.Optimization

	BackendIterativeContext = backend.IterativeContext
	BackendJCodeMunch       = backend.Jcodemunch
	BackendFake             = backend.Fake

	ProviderOpenAI     = provider.Openai
	ProviderOpenRouter = provider.Openrouter
	ProviderCerebras   = provider.Cerebras
	ProviderFake       = provider.Fake

	ReportFormatJSON = reportformat.Json
	ReportFormatText = reportformat.Text

	ArtifactKindPolicy               = artifactkind.Policy
	ArtifactKindNextChallenger       = artifactkind.NextChallenger
	ArtifactKindCompletedRoundBundle = artifactkind.CompletedRoundBundle

	NextChallengerEvidenceReportSummary    = nextchallengerevidencekind.ReportSummary
	NextChallengerEvidenceRoundEvidence    = nextchallengerevidencekind.RoundEvidence
	NextChallengerEvidenceObjectiveResult  = nextchallengerevidencekind.ObjectiveResult
	NextChallengerEvidenceChallengerPolicy = nextchallengerevidencekind.ChallengerPolicy

	OptimizerDeniedGoldLabels        = optimizerdeniedevidencekind.GoldLabels
	OptimizerDeniedOracleFiles       = optimizerdeniedevidencekind.OracleFiles
	OptimizerDeniedRawDatasetAnswers = optimizerdeniedevidencekind.RawDatasetAnswers
)
