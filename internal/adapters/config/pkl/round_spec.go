package config

import (
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/backend"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/nextchallengerevidencekind"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/optimizerdeniedevidencekind"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/provider"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/reportformat"
)

// RoundSpec is the typed root manifest resolved from SearchBenchRound.pkl.
type RoundSpec = generated.SearchBenchRound

type Game = generated.Game
type Dataset = generated.Dataset
type Interfaces = generated.Interfaces
type Interface = generated.Interface
type System = generated.System
type PromptBundle = generated.PromptBundle
type Runtime = generated.Runtime
type PolicyArtifact = generated.PolicyArtifact
type AgentToolPolicy = generated.AgentToolPolicy
type Model = generated.Model
type AgentBounds = generated.AgentBounds
type Evaluator = generated.Evaluator
type Optimizer = generated.Optimizer
type RetryPolicy = generated.RetryPolicy
type Scoring = generated.Scoring
type Report = generated.Report
type RoundManifest = generated.RoundManifest
type RoundPolicy = generated.RoundPolicy
type RoundChallenger = generated.RoundChallenger
type GeneratedChallenger = generated.GeneratedChallenger

type Backend = backend.Backend
type Provider = provider.Provider
type ReportFormat = reportformat.ReportFormat
type NextChallengerEvidenceKind = nextchallengerevidencekind.NextChallengerEvidenceKind
type OptimizerDeniedEvidenceKind = optimizerdeniedevidencekind.OptimizerDeniedEvidenceKind

const (
	BackendIterativeContext = backend.IterativeContext
	BackendJCodeMunch       = backend.Jcodemunch
	BackendFake             = backend.Fake

	ProviderOpenAI     = provider.Openai
	ProviderOpenRouter = provider.Openrouter
	ProviderCerebras   = provider.Cerebras
	ProviderFake       = provider.Fake

	ReportFormatJSON = reportformat.Json
	ReportFormatText = reportformat.Text

	NextChallengerEvidenceReportSummary    = nextchallengerevidencekind.ReportSummary
	NextChallengerEvidenceRoundEvidence    = nextchallengerevidencekind.RoundEvidence
	NextChallengerEvidenceObjectiveResult  = nextchallengerevidencekind.ObjectiveResult
	NextChallengerEvidenceChallengerPolicy = nextchallengerevidencekind.ChallengerPolicy

	OptimizerDeniedGoldLabels        = optimizerdeniedevidencekind.GoldLabels
	OptimizerDeniedOracleFiles       = optimizerdeniedevidencekind.OracleFiles
	OptimizerDeniedRawDatasetAnswers = optimizerdeniedevidencekind.RawDatasetAnswers
)
