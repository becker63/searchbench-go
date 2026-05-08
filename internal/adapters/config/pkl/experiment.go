package config

import (
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/artifactkind"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/backend"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/optimizerdeniedevidencekind"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/optimizerevidencekind"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/provider"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/reportformat"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/runmode"
)

type Experiment = generated.Experiment
type Dataset = generated.Dataset
type Interfaces = generated.Interfaces
type Interface = generated.Interface
type Systems = generated.Systems
type System = generated.System
type PromptBundle = generated.PromptBundle
type Runtime = generated.Runtime
type Artifacts = generated.Artifacts
type PolicyArtifact = generated.PolicyArtifact
type PolicyProposalArtifact = generated.PolicyProposalArtifact
type CompletedEvaluationBundleArtifact = generated.CompletedEvaluationBundleArtifact
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
type CandidateEvaluationBinding = generated.CandidateEvaluationBinding
type CandidateUses = generated.CandidateUses
type Scoring = generated.Scoring
type Report = generated.Report
type Optimization = generated.Optimization
type ParentRun = generated.ParentRun
type OptimizationTarget = generated.OptimizationTarget
type OptimizationEvidence = generated.OptimizationEvidence

type RunMode = runmode.RunMode
type Backend = backend.Backend
type Provider = provider.Provider
type ReportFormat = reportformat.ReportFormat
type ArtifactKind = artifactkind.ArtifactKind
type OptimizerEvidenceKind = optimizerevidencekind.OptimizerEvidenceKind
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

	ArtifactKindPolicy                    = artifactkind.Policy
	ArtifactKindPolicyProposal            = artifactkind.PolicyProposal
	ArtifactKindCompletedEvaluationBundle = artifactkind.CompletedEvaluationBundle

	OptimizerEvidenceReportSummary   = optimizerevidencekind.ReportSummary
	OptimizerEvidenceScoreEvidence   = optimizerevidencekind.ScoreEvidence
	OptimizerEvidenceObjectiveResult = optimizerevidencekind.ObjectiveResult
	OptimizerEvidenceCandidatePolicy = optimizerevidencekind.CandidatePolicy

	OptimizerDeniedGoldLabels        = optimizerdeniedevidencekind.GoldLabels
	OptimizerDeniedOracleFiles       = optimizerdeniedevidencekind.OracleFiles
	OptimizerDeniedRawDatasetAnswers = optimizerdeniedevidencekind.RawDatasetAnswers
)
