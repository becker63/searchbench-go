load("@prelude//:rules.bzl", "test_suite")

# Fast gate: Go + IC smoke (pre-commit runs //:check).
test_suite(
    name = "check",
    tests = [
        "//src/searchbench-go:check",
        "//src/iterative-context:check",
    ],
)

# Full gate (pre-push): harness + IC full + docs + Pkl binding freshness.
test_suite(
    name = "check_full",
    tests = [
        "//src/searchbench-go:check",
        "//src/searchbench-go:pkl_go_types_check",
        "//src/iterative-context:check_full",
        "//docs:check",
    ],
)
