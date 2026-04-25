// Package backend defines the future runtime/session boundary for Searchbench.
//
// It owns abstract backend and session interfaces plus small tool protocol
// shapes that can later be implemented by integrations such as iterative
// context or jCodeMunch adapters.
//
// It does not implement real backend behavior in this repository pass. The
// package exists to make the eventual session boundary explicit and context-
// aware without coupling orchestration to a concrete runtime.
package backend
