// Package eino implements the bounded Eino-backed Optimizer execution path used
// to propose NextChallenger artifacts. Prompt rendering delegates to agents/optimizer/prompt,
// retries and final JSON parsing stay in this package alongside pipeline validation hooks.
//
// Unlike the evaluator eino subtree, callbacks are inlined at the optimizer layer;
// evolve shared callback wiring intentionally if observability parity is required later.
package eino
