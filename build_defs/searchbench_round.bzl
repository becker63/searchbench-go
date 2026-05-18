"""SearchBench repo-owned round targets as typed Buck operations (#91).

Round packages declare non-secret paths and settings as BUCK constants. The
`searchbench_round` macro expands them into `searchbench_round_op` rule targets.
Each operation invokes the private `searchbench buck` CLI with explicit args
constructed in Starlark (no shared shell executor or genrule wrappers).
"""

load("@prelude//decls:toolchains_common.bzl", "toolchains_common")
load("@prelude//test:inject_test_run_info.bzl", "inject_test_run_info")
load("@prelude//tests:re_utils.bzl", "get_re_executors_from_props")

_CLI = "//src/searchbench-go/cmd/searchbench:searchbench"
_PROJECT_ROOT_LABEL = "buck2_run_from_project_root"

_TEST_MODES = [
    "validate",
    "validate_bundle",
    "live_smoke",
]

_RUN_MODES = [
    "run",
    "materialize_dataset",
    "evaluate_n",
    "stability_probe",
]

_ALL_MODES = _TEST_MODES + _RUN_MODES

def _optional_flag(flag, value):
    if value == None or value == "":
        return []
    if type(value) == "int":
        if value == 0:
            return []
        return [cmd_args(str(value), format = flag + "={}")]
    return [cmd_args(value, format = flag + "={}")]

def _round_command(cli, attrs):
    cmd = [
        cli,
        "buck",
        "round",
        cmd_args(attrs.mode, format = "--mode={}"),
        cmd_args(attrs.repo_root, format = "--repo-root={}"),
    ]
    cmd.extend(_optional_flag("--manifest", attrs.manifest))
    cmd.extend(_optional_flag("--artifact-root", attrs.artifact_root))
    cmd.extend(_optional_flag("--bundle-path", attrs.bundle_path))
    cmd.extend(_optional_flag("--evaluate-attempts", attrs.evaluate_attempts))
    cmd.extend(_optional_flag("--stability-attempts", attrs.stability_attempts))
    return cmd_args(cmd)

def _validate_bundle_command(cli, attrs):
    return cmd_args([
        cli,
        "buck",
        "validate-bundle",
        cmd_args(attrs.bundle_path, format = "--bundle-path={}"),
    ])

def _materialize_dataset_command(cli, attrs):
    return cmd_args([
        cli,
        "buck",
        "dataset",
        "materialize-lca",
        cmd_args(attrs.manifest_dir, format = "--manifest-dir={}"),
        cmd_args(attrs.dataset_config, format = "--config={}"),
        cmd_args(attrs.dataset_split, format = "--split={}"),
        cmd_args(str(attrs.dataset_max_items), format = "--max-items={}"),
        cmd_args(str(attrs.dataset_skip), format = "--skip={}"),
    ])

def _operation_command(cli, attrs):
    mode = attrs.mode
    if mode == "validate_bundle":
        return _validate_bundle_command(cli, attrs)
    if mode == "materialize_dataset":
        return _materialize_dataset_command(cli, attrs)
    return _round_command(cli, attrs)

def _cli_executable(ctx):
    cli = ctx.attrs._cli
    if RunInfo in cli:
        return cli[RunInfo]
    outs = cli[DefaultInfo].default_outputs
    if not outs:
        fail("searchbench_round_op: CLI dep %s has no outputs (expected executable artifact)" % cli.label)
    return cmd_args(outs[0])


def _wrap_cli_command(ctx, cli, inner):
    libdir = ctx.attrs._libdir
    return cmd_args([
        "bash",
        "-ec",
        cmd_args(
            cmd_args("export LD_LIBRARY_PATH=$(cat ", libdir, ")"),
            "&&",
            inner,
            delimiter = " ",
        ),
    ])


def _searchbench_round_op_impl(ctx):
    if ctx.attrs.mode not in _ALL_MODES:
        fail("searchbench_round_op: unknown mode %r" % ctx.attrs.mode)

    cli = _cli_executable(ctx)
    command = _wrap_cli_command(ctx, cli, _operation_command(cli, ctx.attrs))
    re_executor, executor_overrides = get_re_executors_from_props(ctx)

    if ctx.attrs.mode in _TEST_MODES:
        labels = list(ctx.attrs.labels or [])
        if _PROJECT_ROOT_LABEL not in labels:
            labels.append(_PROJECT_ROOT_LABEL)
        return inject_test_run_info(
            ctx,
            ExternalRunnerTestInfo(
                type = "custom",
                command = [command],
                labels = labels,
                contacts = ctx.attrs.contacts,
                default_executor = re_executor,
                executor_overrides = executor_overrides,
                run_from_project_root = True,
                use_project_relative_paths = True,
            ),
        ) + [DefaultInfo()]

    return [
        DefaultInfo(),
        RunInfo(args = command),
    ]

searchbench_round_op = rule(
    impl = _searchbench_round_op_impl,
    attrs = {
        "mode": attrs.string(doc = "Private __buck operation mode."),
        "repo_root": attrs.string(
            default = ".",
            doc = "Monorepo root passed to the CLI (project-relative when run_from_project_root).",
        ),
        "manifest": attrs.string(default = ""),
        "manifest_dir": attrs.string(default = ""),
        "artifact_root": attrs.string(default = ""),
        "bundle_path": attrs.string(default = ""),
        "evaluate_attempts": attrs.int(default = 0),
        "stability_attempts": attrs.int(default = 0),
        "dataset_config": attrs.string(default = "py"),
        "dataset_split": attrs.string(default = "dev"),
        "dataset_max_items": attrs.int(default = 1),
        "dataset_skip": attrs.int(default = 50),
        "labels": attrs.list(attrs.string(), default = []),
        "contacts": attrs.list(attrs.string(), default = []),
        "_cli": attrs.dep(
            default = _CLI,
            doc = "Private SearchBench CLI (`go_binary` at //src/searchbench-go/cmd/searchbench:searchbench).",
        ),
        "_libdir": attrs.source(
            default = "//tools:libstdcxx_libdir",
            doc = "One-line file with libstdc++ directory (written by nix develop shellHook).",
        ),
        "_inject_test_env": attrs.default_only(attrs.dep(
            default = "prelude//test/tools:inject_test_env",
            providers = [RunInfo],
        )),
        "_test_toolchain": toolchains_common.test_toolchain(),
    },
)

def searchbench_round(
        name_prefix,
        manifest,
        manifest_dir,
        artifact_root,
        bundle_path,
        repo_root = ".",
        dataset_config = "py",
        dataset_split = "dev",
        dataset_max_items = 1,
        dataset_skip = 50,
        evaluate_attempts = 3,
        stability_probe_attempts = 5,
        **kwargs):
    """Declare the standard SearchBench round target set for one round package."""
    _ = name_prefix
    _common = {
        "repo_root": repo_root,
        "manifest": manifest,
        "manifest_dir": manifest_dir,
        "artifact_root": artifact_root,
        "bundle_path": bundle_path,
        "dataset_config": dataset_config,
        "dataset_split": dataset_split,
        "dataset_max_items": dataset_max_items,
        "dataset_skip": dataset_skip,
    }
    searchbench_round_op(
        name = "validate",
        mode = "validate",
        **kwargs
        | _common,
    )
    searchbench_round_op(
        name = "validate_bundle",
        mode = "validate_bundle",
        **kwargs
        | _common,
    )
    searchbench_round_op(
        name = "live_smoke",
        mode = "live_smoke",
        **kwargs
        | _common,
    )
    searchbench_round_op(
        name = "run",
        mode = "run",
        **kwargs
        | _common,
    )
    searchbench_round_op(
        name = "materialize_dataset",
        mode = "materialize_dataset",
        **kwargs
        | _common,
    )
    searchbench_round_op(
        name = "evaluate_n",
        mode = "evaluate_n",
        evaluate_attempts = evaluate_attempts,
        **kwargs
        | _common,
    )
    searchbench_round_op(
        name = "stability_probe",
        mode = "stability_probe",
        stability_attempts = stability_probe_attempts,
        **kwargs
        | _common,
    )
    searchbench_round_op(
        name = "e2e",
        mode = "live_smoke",
        **kwargs
        | _common,
    )
