"""Reusable Buck operations over Nix-backed toolchain executables (#92).

Targets declare operation shape in Starlark attrs; rules invoke tools directly
via cmd_args (no repo-owned .sh entrypoints).
"""

load("@prelude//decls:toolchains_common.bzl", "toolchains_common")
load("@prelude//test:inject_test_run_info.bzl", "inject_test_run_info")
load("@prelude//tests:re_utils.bzl", "get_re_executors_from_props")

_PROJECT_ROOT_LABEL = "buck2_run_from_project_root"


def _argv_parts(argv):
    return [arg for arg in argv if arg != ""]


def _shell_quote(arg):
    return "'" + arg.replace("'", "'\"'\"'") + "'"


def _external_test(ctx, command):
    re_executor, executor_overrides = get_re_executors_from_props(ctx)
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


def _uv_project_test_impl(ctx):
    work_dir = ctx.attrs.work_dir
    argv = _argv_parts(ctx.attrs.argv)
    dir_prefix = ("--directory " + _shell_quote(work_dir) + " ") if work_dir != "." else ""
    run_args = " ".join([_shell_quote(a) for a in argv])
    if ctx.attrs.sync_first:
        script = "uv {d}sync --locked && uv {d}{args}".format(d=dir_prefix, args=run_args)
    else:
        script = "uv {d}{args}".format(d=dir_prefix, args=run_args)
    command = cmd_args(["bash", "-ec", script])
    return _external_test(ctx, command)


uv_project_test = rule(
    impl = _uv_project_test_impl,
    attrs = {
        "work_dir": attrs.string(
            default = ".",
            doc = "Project directory relative to monorepo root (`.` = repo root).",
        ),
        "sync_first": attrs.bool(
            default = True,
            doc = "When true, run `uv sync --locked` before argv (via shell-safe quoting).",
        ),
        "argv": attrs.list(attrs.string(), doc = "Arguments after optional `uv sync --locked`."),
        "labels": attrs.list(attrs.string(), default = []),
        "contacts": attrs.list(attrs.string(), default = []),
        "_uv": attrs.dep(default = "toolchains//:uv", providers = [RunInfo]),
        "_inject_test_env": attrs.default_only(attrs.dep(
            default = "prelude//test/tools:inject_test_env",
            providers = [RunInfo],
        )),
        "_test_toolchain": toolchains_common.test_toolchain(),
    },
)


def _pkl_go_codegen_cmd(pkl, output_path, schema):
    return cmd_args([
        pkl,
        "run",
        "package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl",
        "--output-path",
        output_path,
        schema,
    ])


def _pkl_go_types_gen_impl(ctx):
    pkl = ctx.attrs._pkl[RunInfo]
    schema = ctx.attrs.schema_path
    generated_dir = ctx.attrs.generated_dir
    temp_dir = ctx.attrs.temp_dir
    temp_generated_dir = temp_dir + "/github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated"
    command = cmd_args([
        "bash",
        "-ec",
        cmd_args(
            cmd_args(["rm", "-rf", temp_dir]),
            " && ",
            cmd_args(["mkdir", "-p", temp_dir]),
            " && ",
            _pkl_go_codegen_cmd(pkl, temp_dir, schema),
            " && ",
            cmd_args(["test", "-d", temp_generated_dir]),
            " && ",
            cmd_args(["mkdir", "-p", generated_dir]),
            " && ",
            cmd_args(["find", generated_dir, "-type", "f", "-name", "*.go", "-delete"]),
            " && ",
            cmd_args(["cp", "-R", temp_generated_dir + "/.", generated_dir + "/"]),
            " && ",
            cmd_args(["rm", "-rf", temp_dir]),
            delimiter = " ",
        ),
    ])
    return [
        DefaultInfo(),
        RunInfo(args = command),
    ]


pkl_go_types_gen = rule(
    impl = _pkl_go_types_gen_impl,
    attrs = {
        "schema_path": attrs.string(doc = "Pkl schema file relative to monorepo root."),
        "generated_dir": attrs.string(
            default = "src/searchbench-go/internal/adapters/config/pkl/generated",
            doc = "Checked-in generated Go bindings directory relative to monorepo root.",
        ),
        "temp_dir": attrs.string(
            default = ".tmp/pkl-go-types-gen",
            doc = "Ignored scratch directory used while regenerating checked-in Go bindings.",
        ),
        "_pkl": attrs.dep(default = "toolchains//:pkl", providers = [RunInfo]),
    },
)


def _pkl_go_types_check_impl(ctx):
    pkl = ctx.attrs._pkl[RunInfo]
    schema = ctx.attrs.schema_path
    generated_dir = ctx.attrs.generated_dir
    temp_dir = ctx.attrs.temp_dir
    temp_generated_dir = temp_dir + "/github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated"
    command = cmd_args([
        "bash",
        "-ec",
        cmd_args(
            cmd_args(["rm", "-rf", temp_dir]),
            " && ",
            cmd_args(["mkdir", "-p", temp_dir]),
            " && ",
            _pkl_go_codegen_cmd(pkl, temp_dir, schema),
            " && ",
            cmd_args(["test", "-d", temp_generated_dir]),
            " && ",
            cmd_args(["diff", "-ru", "--exclude", "BUCK", temp_generated_dir, generated_dir]),
            " && ",
            cmd_args(["rm", "-rf", temp_dir]),
            delimiter = " ",
        ),
    ])
    return _external_test(ctx, command)


pkl_go_types_check = rule(
    impl = _pkl_go_types_check_impl,
    attrs = {
        "schema_path": attrs.string(default = "configs/schema/SearchBenchRound.pkl"),
        "generated_dir": attrs.string(
            default = "src/searchbench-go/internal/adapters/config/pkl/generated",
        ),
        "temp_dir": attrs.string(
            default = ".tmp/pkl-go-types-check",
        ),
        "labels": attrs.list(attrs.string(), default = []),
        "contacts": attrs.list(attrs.string(), default = []),
        "_pkl": attrs.dep(default = "toolchains//:pkl", providers = [RunInfo]),
        "_inject_test_env": attrs.default_only(attrs.dep(
            default = "prelude//test/tools:inject_test_env",
            providers = [RunInfo],
        )),
        "_test_toolchain": toolchains_common.test_toolchain(),
    },
)


def _npm_project_test_impl(ctx):
    npm = ctx.attrs._npm[RunInfo]
    script = ctx.attrs.npm_script
    command = cmd_args([
        "bash",
        "-c",
        cmd_args(npm, " ci && ", npm, " run ", script, delimiter = ""),
    ])
    return _external_test(ctx, command)


npm_project_test = rule(
    impl = _npm_project_test_impl,
    attrs = {
        "npm_script": attrs.string(doc = "npm run <script> after npm ci."),
        "labels": attrs.list(attrs.string(), default = []),
        "contacts": attrs.list(attrs.string(), default = []),
        "_npm": attrs.dep(default = "toolchains//:npm", providers = [RunInfo]),
        "_inject_test_env": attrs.default_only(attrs.dep(
            default = "prelude//test/tools:inject_test_env",
            providers = [RunInfo],
        )),
        "_test_toolchain": toolchains_common.test_toolchain(),
    },
)


def _npm_project_run_impl(ctx):
    npm = ctx.attrs._npm[RunInfo]
    script = ctx.attrs.npm_script
    command = cmd_args([
        "bash",
        "-c",
        cmd_args(npm, " ci && ", npm, " run ", script, delimiter = ""),
    ])
    return [
        DefaultInfo(),
        RunInfo(args = command),
    ]


npm_project_run = rule(
    impl = _npm_project_run_impl,
    attrs = {
        "npm_script": attrs.string(doc = "npm run <script> after npm ci."),
        "_npm": attrs.dep(default = "toolchains//:npm", providers = [RunInfo]),
    },
)
