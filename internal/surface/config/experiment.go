package config

import (
	"github.com/becker63/searchbench-go/internal/surface/config/generated"
	"github.com/becker63/searchbench-go/internal/surface/config/generated/backend"
	"github.com/becker63/searchbench-go/internal/surface/config/generated/provider"
	"github.com/becker63/searchbench-go/internal/surface/config/generated/reportformat"
	"github.com/becker63/searchbench-go/internal/surface/config/generated/runmode"
)

type Experiment = generated.Experiment
type Dataset = generated.Dataset
type Systems = generated.Systems
type System = generated.System
type PromptBundle = generated.PromptBundle
type Runtime = generated.Runtime
type Policy = generated.Policy
type Model = generated.Model
type AgentBounds = generated.AgentBounds
type Evaluator = generated.Evaluator
type RetryPolicy = generated.RetryPolicy
type Writer = generated.Writer
type PipelineProfile = generated.PipelineProfile
type PipelineStep = generated.PipelineStep
type Scoring = generated.Scoring
type Output = generated.Output
type Tracing = generated.Tracing

type RunMode = runmode.RunMode
type Backend = backend.Backend
type Provider = provider.Provider
type ReportFormat = reportformat.ReportFormat

const (
	ModeEvaluatorOnly       = runmode.EvaluatorOnly
	ModeWriterOptimization  = runmode.WriterOptimization
	ModeOptimizationKickoff = runmode.OptimizationKickoff

	BackendIterativeContext = backend.IterativeContext
	BackendJCodeMunch       = backend.Jcodemunch
	BackendFake             = backend.Fake

	ProviderOpenAI     = provider.Openai
	ProviderOpenRouter = provider.Openrouter
	ProviderCerebras   = provider.Cerebras
	ProviderFake       = provider.Fake

	ReportFormatPretty = reportformat.Pretty
	ReportFormatJSON   = reportformat.Json
	ReportFormatBoth   = reportformat.Both
)
