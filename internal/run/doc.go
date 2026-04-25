// Package run defines the Searchbench run lifecycle.
//
// It owns planned, prepared, and executed run records plus the report-facing
// failure model used when execution or scoring cannot produce a successful run.
//
// It does not own metric computation or release decisions. A run becomes a
// scored run in package score, and a comparison becomes a release artifact in
// package report.
package run
