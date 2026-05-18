"""Static SearchBench work-graph registry for BXL planners (#94).

BXL reads this catalog and optional on-disk descriptors; it does not execute
rounds, optimizers, or provider calls.
"""

# Known optimizable backends (expand when jCodeMunch gets optimizable_backend).
BACKEND_CATALOG = {
    "//src/iterative-context:optimizable_backend": {
        "descriptor": "src/iterative-context/optimizable_backend.json",
        "backend_kind": "mcp_stdio",
        "game": "code-localization",
        "source_path": "src/iterative-context",
    },
}

# Round manifests and associated Buck proof targets where modeled.
ROUND_CATALOG = {
    "//configs/rounds/local-ic-vs-jcodemunch:round": {
        "manifest": "configs/rounds/local-ic-vs-jcodemunch/round.pkl",
        "round_dir": "configs/rounds/local-ic-vs-jcodemunch",
        "validate": "//configs/rounds/live-ic-vs-jcodemunch:validate",
        "validate_bundle": "//configs/rounds/live-ic-vs-jcodemunch:validate_bundle",
        "live_smoke": "//configs/rounds/live-ic-vs-jcodemunch:live_smoke",
        "game": "code-localization",
    },
    "//configs/rounds/optimize-ic:round": {
        "manifest": "configs/rounds/optimize-ic/round.pkl",
        "round_dir": "configs/rounds/optimize-ic",
        "validate": "//configs/rounds/live-ic-vs-jcodemunch:validate",
        "note": "optimize-ic uses continuation manifest; validate via shared schema + live round validate until dedicated BUCK package exists",
        "game": "code-localization",
    },
    "//configs/rounds/live-ic-vs-jcodemunch:round": {
        "manifest": "configs/rounds/live-ic-vs-jcodemunch/round.pkl",
        "round_dir": "configs/rounds/live-ic-vs-jcodemunch",
        "validate": "//configs/rounds/live-ic-vs-jcodemunch:validate",
        "validate_bundle": "//configs/rounds/live-ic-vs-jcodemunch:validate_bundle",
        "live_smoke": "//configs/rounds/live-ic-vs-jcodemunch:live_smoke",
        "run": "//configs/rounds/live-ic-vs-jcodemunch:run",
        "game": "code-localization",
    },
}

GLOBAL_PROOF = {
    "minimal": ["//:check"],
    "acceptable": ["//:check", "//:check_full"],
    "fallback": ["//:check_full"],
    "too_live": [
        "//configs/rounds/live-ic-vs-jcodemunch:live_smoke",
        "//configs/rounds/live-ic-vs-jcodemunch:run",
        "//configs/rounds/live-ic-vs-jcodemunch:evaluate_n",
    ],
}

# Path-prefix heuristics for affected_plan (conservative; false positives OK).
PATH_RULES = [
    {
        "prefixes": ["src/iterative-context/"],
        "backends": ["//src/iterative-context:optimizable_backend"],
        "proof_targets": [
            "//src/iterative-context:check",
            "//src/iterative-context:check_full",
        ],
    },
    {
        "prefixes": ["configs/rounds/optimize-ic/"],
        "rounds": ["//configs/rounds/optimize-ic:round"],
        "proof_targets": [
            "//configs/rounds/live-ic-vs-jcodemunch:validate",
            "//:check",
        ],
    },
    {
        "prefixes": ["configs/rounds/local-ic-vs-jcodemunch/", "configs/rounds/live-ic-vs-jcodemunch/"],
        "rounds": [
            "//configs/rounds/local-ic-vs-jcodemunch:round",
            "//configs/rounds/live-ic-vs-jcodemunch:round",
        ],
        "proof_targets": [
            "//configs/rounds/live-ic-vs-jcodemunch:validate",
            "//configs/rounds/live-ic-vs-jcodemunch:validate_bundle",
        ],
    },
    {
        "prefixes": ["src/searchbench-go/"],
        "proof_targets": [
            "//src/searchbench-go:go_native_fast",
            "//src/searchbench-go:check",
        ],
    },
    {
        "prefixes": ["docs/"],
        "possibly_affected_docs": True,
    },
]

EVALUATION_COMPARISONS = [
    {
        "round": "//configs/rounds/local-ic-vs-jcodemunch:round",
        "incumbent": "//src/jcodemunch:optimizable_backend",
        "challenger": "//src/iterative-context:optimizable_backend",
        "note": "jCodeMunch optimizable_backend target not modeled yet; incumbent label is aspirational",
    },
    {
        "round": "//configs/rounds/live-ic-vs-jcodemunch:round",
        "incumbent": None,
        "challenger": "//src/iterative-context:optimizable_backend",
        "note": "live round uses IC challenger via buck_descriptor in round.pkl",
    },
]
