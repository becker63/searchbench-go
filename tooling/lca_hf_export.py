#!/usr/bin/env python3
"""Export JetBrains LCA rows from Hugging Face to SearchBench JSONL."""

from __future__ import annotations

import argparse
import ast
import json
import sys
from pathlib import Path
from typing import Any

DATASET_NAME = "JetBrains-Research/lca-bug-localization"
DATASET_DIR = "JetBrains-Research_lca-bug-localization"


def normalize_changed_files(value: Any) -> list[str]:
    if value is None:
        return []
    if isinstance(value, list):
        return [str(item) for item in value]
    if isinstance(value, str):
        text = value.strip()
        if not text:
            return []
        try:
            parsed = json.loads(text)
        except json.JSONDecodeError:
            parsed = ast.literal_eval(text)
        if isinstance(parsed, list):
            return [str(item) for item in parsed]
        raise ValueError(f"changed_files string did not decode to a list: {value!r}")
    raise TypeError(f"unsupported changed_files type: {type(value).__name__}")


def normalize_repo_languages(value: Any) -> list[str] | None:
    if value is None:
        return None
    if isinstance(value, dict):
        return sorted(str(key) for key in value.keys())
    if isinstance(value, list):
        return [str(item) for item in value]
    return None


def row_to_json(row: dict[str, Any]) -> dict[str, Any]:
    out: dict[str, Any] = {
        "repo_owner": row["repo_owner"],
        "repo_name": row["repo_name"],
        "base_sha": row["base_sha"],
        "issue_title": row.get("issue_title") or "",
        "issue_body": row.get("issue_body") or "",
        "changed_files": normalize_changed_files(row.get("changed_files")),
    }
    for key in (
        "issue_url",
        "pull_url",
        "diff_url",
        "diff",
        "head_sha",
        "repo_language",
        "repo_license",
    ):
        val = row.get(key)
        if val not in (None, ""):
            out[key] = val
    if row.get("repo_stars") is not None:
        out["repo_stars"] = int(row["repo_stars"])
    langs = normalize_repo_languages(row.get("repo_languages"))
    if langs:
        out["repo_languages"] = langs
    return out


def export_slice(
    *,
    config: str,
    split: str,
    max_items: int,
    skip: int,
    output_dir: Path,
) -> Path:
    if max_items < 1:
        raise ValueError("max_items must be >= 1")

    try:
        from datasets import load_dataset
    except ImportError as exc:  # pragma: no cover - environment hint
        raise SystemExit(
            "datasets package required; use nix develop (python3Packages.datasets) "
            "or pip install datasets huggingface_hub"
        ) from exc

    dataset = load_dataset(DATASET_NAME, config, split=split, streaming=True)
    out_dir = output_dir / "datasets" / DATASET_DIR / config
    out_dir.mkdir(parents=True, exist_ok=True)
    out_path = out_dir / f"{split}.jsonl"

    skipped = 0
    count = 0
    with out_path.open("w", encoding="utf-8") as handle:
        for row in dataset:
            if skipped < skip:
                skipped += 1
                continue
            if count >= max_items:
                break
            record = row_to_json(dict(row))
            handle.write(json.dumps(record, ensure_ascii=False) + "\n")
            count += 1

    if count == 0:
        raise RuntimeError(f"no rows exported for {DATASET_NAME} config={config} split={split}")

    return out_path


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--config", default="py", help="HF dataset config (py, java, kt)")
    parser.add_argument("--split", default="dev", help="HF split (dev, train, test)")
    parser.add_argument("--max-items", type=int, default=1, help="rows to export")
    parser.add_argument(
        "--skip",
        type=int,
        default=0,
        help="streaming rows to skip before exporting (cost control for large repos)",
    )
    parser.add_argument(
        "--output-dir",
        type=Path,
        required=True,
        help="round manifest directory (parent of datasets/)",
    )
    args = parser.parse_args(argv)

    out_path = export_slice(
        config=args.config,
        split=args.split,
        max_items=args.max_items,
        skip=args.skip,
        output_dir=args.output_dir.resolve(),
    )
    print(f"exported {args.max_items} row(s) to {out_path}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
