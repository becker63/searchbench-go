"""Buck macros for repo-owned SearchBench round targets (#91)."""

load("@prelude//:rules.bzl", "sh_binary", "sh_test")

def searchbench_round_validate(name, repo_root, manifest, **kwargs):
    _ = (repo_root, manifest)
    sh_test(
        name = name,
        test = "scripts/validate.sh",
        **kwargs
    )

def searchbench_round_validate_bundle(name, repo_root, bundle_path, **kwargs):
    _ = (repo_root, bundle_path)
    sh_test(
        name = name,
        test = "scripts/validate_bundle.sh",
        **kwargs
    )

def searchbench_round_live_smoke(name, repo_root, manifest, artifact_root, bundle_path, **kwargs):
    _ = (repo_root, manifest, artifact_root, bundle_path)
    sh_test(
        name = name,
        test = "scripts/live_smoke.sh",
        **kwargs
    )

def searchbench_round_run(name, repo_root, manifest, artifact_root, **kwargs):
    _ = (repo_root, manifest, artifact_root)
    sh_binary(
        name = name,
        main = "scripts/run.sh",
        **kwargs
    )

def searchbench_round_materialize_dataset(name, repo_root, manifest_dir, **kwargs):
    _ = (repo_root, manifest_dir)
    sh_binary(
        name = name,
        main = "scripts/materialize_dataset.sh",
        **kwargs
    )

def searchbench_round_evaluate_n(name, repo_root, manifest, artifact_root, bundle_path, attempts = 3, **kwargs):
    _ = (repo_root, manifest, artifact_root, bundle_path, attempts)
    sh_binary(
        name = name,
        main = "scripts/evaluate_n.sh",
        **kwargs
    )

def searchbench_round_stability_probe(name, repo_root, manifest, artifact_root, bundle_path, attempts = 5, **kwargs):
    _ = (repo_root, manifest, artifact_root, bundle_path, attempts)
    sh_binary(
        name = name,
        main = "scripts/stability_probe.sh",
        **kwargs
    )
