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

def _go_module_test_impl(ctx):
    go = ctx.attrs._go[RunInfo]
    command = cmd_args([
        go,
        "-C",
        ctx.attrs.module_dir,
    ] + _argv_parts(ctx.attrs.argv))
    return _external_test(ctx, command)

go_module_test = rule(
    impl = _go_module_test_impl,
    attrs = {
        "module_dir": attrs.string(doc = "Go module directory relative to monorepo root."),
        "argv": attrs.list(attrs.string(), doc = "Arguments after `go -C <module_dir>`."),
        "labels": attrs.list(attrs.string(), default = []),
        "contacts": attrs.list(attrs.string(), default = []),
        "_go": attrs.dep(default = "toolchains//:go", providers = [RunInfo]),
        "_inject_test_env": attrs.default_only(attrs.dep(
            default = "prelude//test/tools:inject_test_env",
            providers = [RunInfo],
        )),
        "_test_toolchain": toolchains_common.test_toolchain(),
    },
)

def _pkl_go_types_gen_impl(ctx):
    pkl = ctx.attrs._pkl[RunInfo]
    schema = ctx.attrs.schema_path
    work_dir = ctx.attrs.work_dir
    command = cmd_args([
        pkl,
        "run",
        "package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl",
        "--output-path",
        work_dir,
        schema,
    ])
    return _external_test(ctx, command)

pkl_go_types_gen = rule(
    impl = _pkl_go_types_gen_impl,
    attrs = {
        "schema_path": attrs.string(doc = "Pkl schema file relative to monorepo root."),
        "work_dir": attrs.string(
            default = "src/searchbench-go",
            doc = "Go module directory for codegen output (writes under --output-path=.).",
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

def _pkl_go_types_check_impl(ctx):
    pkl = ctx.attrs._pkl[RunInfo]
    git = ctx.attrs._git[RunInfo]
    schema = ctx.attrs.schema_path
    work_dir = ctx.attrs.work_dir
    generated_glob = ctx.attrs.generated_glob
    command = cmd_args([
        "bash",
        "-ec",
        cmd_args(
            cmd_args([pkl, "run", "package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.13.2#/gen.pkl", "--output-path", work_dir, schema]),
            " && ",
            cmd_args([git, "diff", "--quiet", "--", generated_glob]),
            " && ",
            cmd_args([git, "diff", "--quiet", "--cached", "--", generated_glob]),
            delimiter = " ",
        ),
    ])
    return _external_test(ctx, command)

pkl_go_types_check = rule(
    impl = _pkl_go_types_check_impl,
    attrs = {
        "schema_path": attrs.string(default = "configs/schema/SearchBenchRound.pkl"),
        "work_dir": attrs.string(default = "src/searchbench-go"),
        "generated_glob": attrs.string(
            default = "src/searchbench-go/internal/adapters/config/pkl/generated/",
        ),
        "labels": attrs.list(attrs.string(), default = []),
        "contacts": attrs.list(attrs.string(), default = []),
        "_pkl": attrs.dep(default = "toolchains//:pkl", providers = [RunInfo]),
        "_git": attrs.dep(default = "toolchains//:git", providers = [RunInfo]),
        "_inject_test_env": attrs.default_only(attrs.dep(
            default = "prelude//test/tools:inject_test_env",
            providers = [RunInfo],
        )),
        "_test_toolchain": toolchains_common.test_toolchain(),
    },
)

def _uv_project_test_impl(ctx):
    uv = ctx.attrs._uv[RunInfo]
    work_dir = ctx.attrs.work_dir
    sync_cmd = [uv, "sync", "--locked"]
    run_cmd = [uv]
    if work_dir != ".":
        sync_cmd.extend(["--directory", work_dir])
        run_cmd.extend(["--directory", work_dir])
    run_cmd.extend(_argv_parts(ctx.attrs.argv))
    if ctx.attrs.sync_first:
        command = cmd_args([
            "bash",
            "-ec",
            cmd_args(cmd_args(sync_cmd), " && ", cmd_args(run_cmd), delimiter = " "),
        ])
    else:
        command = cmd_args(run_cmd)
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
