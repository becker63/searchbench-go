load("@prelude//:rules.bzl", "sh_test", "test_suite")

sh_test(
    name = "repomix_fresh_check",
    test = "repomix_fresh_check.sh",
)

# Fast gate: Go module tests + CLI build + Iterative Context `check` (import smoke + pytest; no Repomix).
test_suite(
    name = "check",
    tests = [
        "//src/searchbench-go:check",
        "//src/iterative-context:check",
    ],
)

# Full gate: Go `check` + Iterative Context `check_full` (adds basedpyright) + Repomix snapshot freshness (pre-push / manual).
test_suite(
    name = "check_full",
    tests = [
        "//src/searchbench-go:check",
        "//src/iterative-context:check_full",
        "//docs:check",
        ":repomix_fresh_check",
    ],
)
