"""Buck graph traversal for SearchBench BXL planners (no static registry)."""

_PROOF_GATE_SUITES = [
    "//:check",
    "//:check_full",
    "//src/searchbench-go:check",
    "//src/searchbench-go:go_native_fast",
    "//src/searchbench-go:go_native_full",
    "//src/iterative-context:check",
    "//src/iterative-context:check_full",
]

_LIVE_ROUND_MODES = [
    "live_smoke",
    "run",
    "evaluate_n",
    "stability_probe",
    "e2e",
]

def canonical_label(target):
    s = str(target.label)
    if s.startswith("root//"):
        return "//" + s[len("root//"):]
    return s

def any_path_prefix(paths, prefix):
    for p in paths:
        if p.startswith(prefix):
            return True
    return False

def dedupe_preserve(items):
    seen = {}
    out = []
    for item in items:
        key = canonical_label(item) if hasattr(item, "label") else item
        if key in seen:
            continue
        seen[key] = True
        out.append(item)
    return out

def repo_pattern(ctx):
    return ctx.unconfigured_targets(["//..."])

def discover_backends(q, ctx):
    return q.attrfilter("name", "optimizable_backend", repo_pattern(ctx))

def discover_round_ops(q, ctx):
    return q.kind("^searchbench_round_op$", repo_pattern(ctx))

def proof_gate_suites(ctx):
    return ctx.unconfigured_targets(_PROOF_GATE_SUITES)

def round_ops_for_manifest(q, ctx, manifest_path):
    if manifest_path == None or manifest_path == "":
        return []
    return q.attrfilter("manifest", manifest_path, discover_round_ops(q, ctx))

def round_ops_for_manifest_dir(q, ctx, manifest_dir):
    if manifest_dir == None or manifest_dir == "":
        return []
    return q.attrfilter("manifest_dir", manifest_dir, discover_round_ops(q, ctx))

def manifest_dir_from_path(path):
    if path.endswith("/round.pkl"):
        return path[:-len("/round.pkl")]
    if path.startswith("configs/rounds/") and path.endswith(".pkl"):
        return path.rsplit("/", 1)[0]
    if path.startswith("configs/rounds/"):
        parts = path.split("/")
        if len(parts) >= 3:
            return "/".join(parts[0:3])
    return None

def buildfile_labels_for_path(path):
    if path.endswith("/BUCK"):
        return ["//{}/BUCK".format(path[:-len("/BUCK")])]
    parts = path.split("/")
    labels = []
    for i in range(len(parts) - 1, 0, -1):
        pkg = "/".join(parts[0:i])
        if pkg:
            labels.append("//{}/BUCK".format(pkg))
    return labels

def owners_for_paths(q, paths):
    owned = []
    for path in paths:
        found = q.owner([path])
        if len(found) > 0:
            owned.extend(found)
    return dedupe_preserve(owned)

def targets_in_buildfile_for_path(q, ctx, path):
    for bf in buildfile_labels_for_path(path):
        found = q.targets_in_buildfile(bf)
        if len(found) > 0:
            return found
    return []

def seeds_for_changed_paths(q, ctx, paths):
    seeds = owners_for_paths(q, paths)
    for path in paths:
        seeds.extend(targets_in_buildfile_for_path(q, ctx, path))
    return dedupe_preserve(seeds)

def _label_in_attr_list(attr_value, needle):
    if attr_value == None:
        return False
    for item in attr_value:
        if needle in str(item):
            return True
    return False

def go_tests_referencing_label(ctx, resource_label):
    q = ctx.uquery()
    candidates = q.kind("^go_test$", ctx.unconfigured_targets(["//src/searchbench-go/..."]))
    hits = []
    for node in candidates:
        configured = ctx.configured_targets([node])
        if len(configured) == 0:
            continue
        target = configured[0]
        if _label_in_attr_list(target.get_attr("resources"), resource_label):
            hits.append(node)
        elif _label_in_attr_list(target.get_attr("deps"), resource_label):
            hits.append(node)
    return dedupe_preserve(hits)

def proof_targets_from_rdeps(q, ctx, seeds):
    suites = proof_gate_suites(ctx)
    if len(seeds) == 0:
        return []
    return q.rdeps(suites, seeds)

def round_ops_for_changed_paths(q, ctx, paths):
    ops = []
    for path in paths:
        ops.extend(round_ops_for_manifest(q, ctx, path))
        md = manifest_dir_from_path(path)
        if md != None:
            ops.extend(round_ops_for_manifest_dir(q, ctx, md))
    return dedupe_preserve(ops)

def logical_round_label(manifest_dir):
    return "//{}:round".format(manifest_dir)

def round_op_attrs(ctx, q):
    """Read manifest + manifest_dir from configured searchbench_round_op validate targets."""
    manifests = []
    manifest_dirs = []
    validate_ops = q.attrfilter("name", "validate", discover_round_ops(q, ctx))
    for node in validate_ops:
        configured = ctx.configured_targets([node])
        if len(configured) == 0:
            continue
        target = configured[0]
        manifest = target.get_attr("manifest")
        manifest_dir = target.get_attr("manifest_dir")
        if manifest != None:
            manifests.append(str(manifest))
        if manifest_dir != None:
            manifest_dirs.append(str(manifest_dir))
    return dedupe_preserve(manifests), dedupe_preserve(manifest_dirs)

def logical_rounds_in_graph(ctx, q):
    _, manifest_dirs = round_op_attrs(ctx, q)
    return dedupe_preserve([logical_round_label(md) for md in manifest_dirs])

def _normalize_repo_path(path):
    s = str(path).strip()
    prefix = "<source artifact "
    suffix = ">"
    if s.startswith(prefix) and s.endswith(suffix):
        return s[len(prefix):-len(suffix)]
    if s.startswith("root///"):
        return s[len("root///"):]
    if s.startswith("root//"):
        return s[len("root//"):]
    return s

def logical_rounds_from_test_resources(q, ctx):
    """Logical round labels for Pkl manifests wired into //:searchbench_go_test_resources."""
    rounds = []
    configured = ctx.configured_targets(
        ctx.unconfigured_targets(["//:searchbench_go_test_resources"]),
    )
    if len(configured) > 0:
        srcs = configured[0].get_attr("srcs")
        if srcs != None:
            for src in srcs:
                path = _normalize_repo_path(src)
                if path.endswith("/round.pkl") and path.startswith("configs/rounds/"):
                    rounds.append(logical_round_label(manifest_dir_from_path(path)))
    return dedupe_preserve(rounds)

def all_logical_rounds(ctx, q):
    return dedupe_preserve(logical_rounds_in_graph(ctx, q) + logical_rounds_from_test_resources(q, ctx))

def manifest_only_rounds(q, ctx, paths):
    """Logical rounds for Pkl manifests that are not backed by searchbench_round_op."""
    _, graph_dirs = round_op_attrs(ctx, q)
    rounds = []
    for path in paths:
        if not path.endswith("round.pkl"):
            continue
        md = manifest_dir_from_path(path)
        if md == None or md in graph_dirs:
            continue
        rounds.append(logical_round_label(md))
    return dedupe_preserve(rounds)

def classify_round_ops(round_ops):
    by_name = {}
    for op in round_ops:
        name = canonical_label(op).split(":")[-1]
        by_name[name] = canonical_label(op)
    minimal = []
    if "validate" in by_name:
        minimal.append(by_name["validate"])
    acceptable = dedupe_preserve(minimal + [
        by_name[k] for k in ["validate_bundle", "validate"] if k in by_name
    ])
    too_live = [by_name[k] for k in _LIVE_ROUND_MODES if k in by_name]
    return minimal, acceptable, too_live

def descriptor_path_for_backend(q, ctx, target_label):
    for node in discover_backends(q, ctx):
        if canonical_label(node) != target_label:
            continue
        owners = q.owner(["src/iterative-context/optimizable_backend.json"])
        for owner in owners:
            if canonical_label(owner) == target_label:
                return "src/iterative-context/optimizable_backend.json"
    return None

def package_prefix(label):
    if ":" in label:
        return label.split(":")[0]
    return label

def backends_for_seeds(q, ctx, seeds):
    hits = []
    for backend in discover_backends(q, ctx):
        bl = canonical_label(backend)
        pkg = package_prefix(bl)
        for seed in seeds:
            sl = canonical_label(seed)
            if sl == bl or package_prefix(sl) == pkg:
                hits.append(bl)
    return dedupe_preserve(hits)

def suites_containing_tests(ctx, test_nodes):
    """Find test_suite targets that list the given tests (reads suite attrs, not registry)."""
    if len(test_nodes) == 0:
        return []
    test_labels = [canonical_label(t) for t in test_nodes]
    hits = []
    suites = ctx.uquery().kind("^test_suite$", repo_pattern(ctx))
    for suite in suites:
        configured = ctx.configured_targets([suite])
        if len(configured) == 0:
            continue
        members = configured[0].get_attr("tests")
        if members == None:
            continue
        member_strs = [str(m) for m in members]
        for tl in test_labels:
            matched = tl in member_strs
            if not matched:
                for m in member_strs:
                    if tl in m:
                        matched = True
                        break
            if matched:
                hits.append(canonical_label(suite))
                break
    return dedupe_preserve(hits)
