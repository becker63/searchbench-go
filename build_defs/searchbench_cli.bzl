"""Build the private SearchBench CLI binary for Buck round operations (#91)."""

def _searchbench_cli_impl(ctx):
    out = ctx.actions.declare_output("searchbench")
    go = ctx.attrs._go[RunInfo]
    pkg = ctx.attrs.go_package
    wrapper = ctx.actions.declare_output("build_searchbench.sh")
    ctx.actions.write(
        wrapper,
        cmd_args(
            "#!/usr/bin/env bash\n",
            "set -euo pipefail\n",
            "root=\"$(pwd)\"\n",
            "out=\"$1\"\n",
            "if [[ \"${out}\" != /* ]]; then out=\"${root}/${out#./}\"; fi\n",
            "exec ",
            go,
            ' -C "${root}/' + pkg + '" build -trimpath -o "${out}" ./cmd/searchbench\n',
            delimiter = "",
        ),
        is_executable = True,
    )
    ctx.actions.run(
        cmd_args(wrapper, out.as_output()),
        category = "searchbench_cli",
        local_only = True,
    )
    return [
        DefaultInfo(default_output = out),
        RunInfo(args = cmd_args(out)),
    ]

searchbench_cli = rule(
    impl = _searchbench_cli_impl,
    attrs = {
        "go_package": attrs.string(
            default = "src/searchbench-go",
            doc = "Module directory relative to the monorepo root.",
        ),
        "_go": attrs.dep(
            default = "toolchains//:go",
            providers = [RunInfo],
        ),
    },
)
