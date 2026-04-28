// Package scoring executes human-authored Pkl objective files against
// harness-owned score.pkl evidence artifacts and returns validated
// score.ObjectiveResult values.
//
// This package is effectful and intentionally lives outside internal/score.
// The pure score package owns evidence and objective result models. This
// package owns only the Pkl execution seam.
package scoring
