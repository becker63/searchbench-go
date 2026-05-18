"""Shared helpers for SearchBench BXL planners."""

load("//tooling/bxl:registry.bzl", "BACKEND_CATALOG", "EVALUATION_COMPARISONS", "GLOBAL_PROOF", "PATH_RULES", "ROUND_CATALOG")

def _backend_resolution_entry(target, meta):
    return {
        "kind": "searchbench.backend_resolution.v1",
        "target": target,
        "resolved": True,
        "descriptor": meta["descriptor"],
        "source": {
            "kind": "local_path",
            "path": meta["source_path"],
        },
        "launcher": {
            "kind": "mcp_stdio",
            "cwd_mode": "candidate_workspace",
        },
        "candidate_validator": {
            "kind": "ic_policy_pipeline",
        },
        "runtime_admin": {
            "install_tool": "install_score",
            "verify_tool": "verify_score",
            "hidden_from_evaluator": True,
        },
        "metadata": {
            "backend_kind": meta["backend_kind"],
            "game": meta["game"],
        },
    }

def resolve_backend_doc(target):
    meta = BACKEND_CATALOG.get(target)
    if meta == None:
        return {
            "kind": "searchbench.backend_resolution.v1",
            "target": target,
            "resolved": False,
            "error": "unknown backend target; see tooling/bxl/registry.bzl BACKEND_CATALOG",
        }
    return _backend_resolution_entry(target, meta)

def backend_inventory_doc():
    backends = []
    for target, meta in BACKEND_CATALOG.items():
        backends.append({
            "target": target,
            "backend_kind": meta["backend_kind"],
            "game": meta["game"],
            "descriptor": meta["descriptor"],
        })
    return {
        "kind": "searchbench.backend_inventory.v1",
        "backends": backends,
    }

def proof_plan_doc(target):
    round_meta = ROUND_CATALOG.get(target)
    if round_meta == None:
        return {
            "kind": "searchbench.proof_plan.v1",
            "target": target,
            "goal": "validate target (generic)",
            "minimal_targets": GLOBAL_PROOF["minimal"],
            "acceptable_targets": GLOBAL_PROOF["acceptable"],
            "fallback_targets": GLOBAL_PROOF["fallback"],
            "too_live_for_default_gate": GLOBAL_PROOF["too_live"],
            "note": "target not in ROUND_CATALOG; using global gates only",
        }

    minimal = []
    if round_meta.get("validate"):
        minimal.append(round_meta["validate"])
    if not minimal:
        minimal = ["//:check"]

    acceptable = minimal + [x for x in [
        round_meta.get("validate_bundle"),
        "//src/searchbench-go:check",
        "//:check",
    ] if x != None]

    too_live = []
    for key in ["live_smoke", "run", "evaluate_n"]:
        t = round_meta.get(key)
        if t:
            too_live.append(t)
    too_live.extend(GLOBAL_PROOF["too_live"])

    return {
        "kind": "searchbench.proof_plan.v1",
        "goal": "validate round manifest at {}".format(round_meta["manifest"]),
        "target": target,
        "manifest": round_meta["manifest"],
        "minimal_targets": dedupe_preserve(minimal),
        "acceptable_targets": dedupe_preserve(acceptable + GLOBAL_PROOF["acceptable"]),
        "fallback_targets": GLOBAL_PROOF["fallback"],
        "too_live_for_default_gate": dedupe_preserve(too_live),
    }

def dedupe_preserve(items):
    seen = {}
    out = []
    for item in items:
        if item in seen:
            continue
        seen[item] = True
        out.append(item)
    return out

def affected_plan_doc(changed_files_raw):
    changed = []
    if changed_files_raw:
        for line in changed_files_raw.splitlines():
            line = line.strip()
            if line:
                changed.append(line)
    affected_backends = []
    affected_rounds = []
    proof_targets = []
    possibly_docs = []

    for path in changed:
        for rule in PATH_RULES:
            matched = False
            for prefix in rule["prefixes"]:
                if path.startswith(prefix):
                    matched = True
                    break
            if not matched:
                continue
            affected_backends.extend(rule.get("backends", []))
            affected_rounds.extend(rule.get("rounds", []))
            proof_targets.extend(rule.get("proof_targets", []))
            if rule.get("possibly_affected_docs"):
                possibly_docs.append(path)

    return {
        "kind": "searchbench.affected_plan.v1",
        "changed_files": changed,
        "affected_backends": dedupe_preserve(affected_backends),
        "affected_rounds": dedupe_preserve(affected_rounds),
        "proof_targets": dedupe_preserve(proof_targets),
        "fallback_targets": GLOBAL_PROOF["fallback"],
        "possibly_affected_docs": dedupe_preserve(possibly_docs),
        "heuristic": True,
    }

def evaluation_matrix_doc():
    rounds = dedupe_preserve(ROUND_CATALOG.keys())
    backends = dedupe_preserve(BACKEND_CATALOG.keys())
    return {
        "kind": "searchbench.evaluation_matrix.v1",
        "rounds": rounds,
        "backends": backends,
        "comparisons": EVALUATION_COMPARISONS,
    }

def target_summary_doc(target):
    resolved = target in BACKEND_CATALOG or target in ROUND_CATALOG
    kind = "unknown"
    if target in BACKEND_CATALOG:
        kind = "optimizable_backend"
    elif target in ROUND_CATALOG:
        kind = "round"
    elif target in ("//:check", "//:check_full"):
        kind = "repo_gate"

    return {
        "kind": "searchbench.bxl_target_summary.v1",
        "target": target,
        "resolved": resolved,
        "metadata": {
            "label": target,
            "kind": kind,
        },
    }
