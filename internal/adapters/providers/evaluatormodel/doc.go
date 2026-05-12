// Package evaluatormodel resolves manifest evaluator model settings plus process
// environment into an Eino ToolCallingChatModel factory. Provider SDK types stay
// in this adapter; pure packages receive only usage records and execution results.
//
// Environment (OpenAI-compatible HTTP):
//
//	openai:     OPENAI_API_KEY, optional OPENAI_BASE_URL
//	openrouter: OPENROUTER_API_KEY (or OPENAI_API_KEY), optional OPENROUTER_BASE_URL (default https://openrouter.ai/api/v1)
//	cerebras:   CEREBRAS_API_KEY, optional CEREBRAS_BASE_URL (default https://api.cerebras.ai/v1)
//
// Manifest provider "fake" (or missing/empty provider) uses the local deterministic
// fake model. If a cloud provider is selected but credentials are missing, the
// factory falls back to the fake model so tests and CI stay offline by default.
package evaluatormodel
