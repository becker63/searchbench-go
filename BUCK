load("@prelude//:rules.bzl", "alias", "test_suite")

# Back-compat alias; prefer //tooling:repomix_fresh_check.
alias(
    name = "repomix_fresh_check",
    actual = "//tooling:repomix_fresh_check",
)

# Fast gate: Go + IC smoke (pre-commit runs //tooling:repomix then //:check).
test_suite(
    name = "check",
    tests = [
        "//src/searchbench-go:check",
        "//src/iterative-context:check",
    ],
)

# Full gate (pre-push): harness + IC full + docs + Pkl binding freshness + Repomix freshness.
test_suite(
    name = "check_full",
    tests = [
        "//src/searchbench-go:check",
        "//src/searchbench-go:pkl_go_types_check",
        "//src/iterative-context:check_full",
        "//docs:check",
        "//tooling:repomix_fresh_check",
    ],
)
