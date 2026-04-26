// Package modeltest provides deterministic model-testing fixtures.
//
// SearchBench-Go uses three testing tiers for model-backed code:
//
//   - Tier 1: an in-process scripted Eino-compatible chat model
//   - Tier 2: a tiny local OpenAI-compatible httptest server
//   - Tier 3: opt-in live model tests only
//
// Default tests must remain offline, deterministic, and safe to run without
// real API keys or paid provider calls.
//
// See docs/engineering/model-testing.md for usage guidance, current helper
// capabilities, and how later executor issues should use Tier 1 and Tier 2.
package modeltest
