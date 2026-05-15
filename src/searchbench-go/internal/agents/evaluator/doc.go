// Package evaluator is the vertical slice for the Evaluator agent: prompt types,
// Eino-backed execution, and consolidated local fakes. internal/app/round owns
// the round lifecycle and imports from here; this tree must not import app/round.
package evaluator
