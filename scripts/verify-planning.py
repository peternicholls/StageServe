#!/usr/bin/env python3
"""Read-only structural checks for spec 012; never certifies runtime readiness."""
from __future__ import annotations

import argparse
import hashlib
import json
import re
import subprocess
import sys
from pathlib import Path


def task_refs(text: str) -> set[str]:
    refs: set[str] = set()
    for match in re.finditer(r"T(\d{3})(?:[–-]T(\d{3}))?", text):
        first = int(match[1])
        last = int(match[2]) if match[2] else first
        if last < first:
            raise ValueError(f"Reversed task range: {match[0]}")
        refs.update(f"T{n:03d}" for n in range(first, last + 1))
    return refs


def verify(root: Path) -> list[str]:
    errors: list[str] = []
    feature = root / "specs/012-apple-only-experience"
    required = ["spec.md", "plan.md", "tasks.md", "data-model.md", "research.md",
                "quickstart.md", "analysis.md", "review-resolution.md", "ux-validation.md",
                "contracts/application-contract.md", "contracts/project-runtime-contract.md",
                "contracts/release-qualification.md", "checklists/requirements.md"]
    for rel in required:
        if not (feature / rel).is_file():
            errors.append(f"Missing artifact: {rel}")
    if errors:
        return errors
    task_text = (feature / "tasks.md").read_text()
    tasks: dict[str, set[str]] = {}
    for line in task_text.splitlines():
        match = re.match(r"- \[[ x]\] (T\d{3})\b", line)
        if not match:
            continue
        tid = match[1]
        if tid in tasks:
            errors.append(f"Duplicate task: {tid}")
        dep = re.search(r"depends on ([^)]+)\)", line)
        if not dep:
            errors.append(f"Missing explicit dependencies: {tid}")
        tasks[tid] = task_refs(dep[1]) if dep else set()
    for tid, deps in tasks.items():
        for dep in deps - tasks.keys():
            errors.append(f"Unknown dependency {dep} in {tid}")
    active: set[str] = set()
    done: set[str] = set()

    def visit(tid: str) -> None:
        if tid in active:
            errors.append(f"Dependency cycle at {tid}")
            return
        if tid in done or tid not in tasks:
            return
        active.add(tid)
        for dep in tasks[tid]:
            visit(dep)
        active.remove(tid)
        done.add(tid)

    for tid in tasks:
        visit(tid)
    spec_text = (feature / "spec.md").read_text()
    requirements = set(re.findall(r"\*\*(FR-\d{3})\*\*:", spec_text))
    coverage: dict[str, tuple[set[str], set[str]]] = {}
    for match in re.finditer(r"^\| (FR-\d{3}) \| ([^|]+) \| ([^|]+) \|", task_text, re.M):
        if match[1] in coverage:
            errors.append(f"Duplicate coverage row: {match[1]}")
        coverage[match[1]] = (task_refs(match[2]), task_refs(match[3]))
    if requirements != coverage.keys():
        errors.append(f"Requirement coverage mismatch: {sorted(requirements ^ coverage.keys())}")
    for requirement, lanes in coverage.items():
        for lane in lanes:
            if not lane or lane - tasks.keys():
                errors.append(f"Invalid implementation/verification mapping: {requirement}")
    outcomes = set(re.findall(r"\*\*(SC-\d{3})\*\*:", spec_text))
    for outcome in outcomes:
        if outcome.replace("-", "") not in task_text:
            errors.append(f"Success criterion lacks task mapping: {outcome}")

    docs = list(feature.rglob("*.md")) + [root / p for p in [
        "docs/product-direction.md", "docs/roadmap.md", "specs/README.md",
        "mockups/README.md", "archive/2026-09-11-pre-apple-only/README.md"]]
    for path in docs:
        text = path.read_text()
        if "[NEEDS CLARIFICATION" in text:
            errors.append(f"Unresolved clarification: {path.relative_to(root)}")
        for link in re.findall(r"\]\(([^)]+)\)", text):
            if "://" in link or link.startswith(("#", "mailto:")):
                continue
            target = link.split("#", 1)[0]
            if not (path.parent / target).exists():
                errors.append(f"Broken link in {path.relative_to(root)}: {link}")
    archive = root / "archive/2026-09-11-pre-apple-only"
    content = json.loads((archive / "manifest.json").read_text())["files"]
    original = json.loads((archive / "filesystem-manifest.json").read_text())["files"]
    metadata = json.loads((archive / "local-metadata-manifest.json").read_text())["files"]
    if len(content) != 88 or len(original) != 90 or len(metadata) != 2:
        errors.append("Archive inventory counts must be 88 content + 2 metadata = 90 snapshot")
    by_path = {entry["archived"]: entry for entry in original}
    if len(by_path) != len(original):
        errors.append("Duplicate original archive paths")
    combined = {entry["archived"]: entry for entry in content + metadata}
    if len(combined) != len(content) + len(metadata) or combined != by_path:
        errors.append("Split inventories do not preserve the original snapshot entries")
    for entry in content:
        path = (root / entry["archived"]).resolve()
        if not path.is_relative_to(archive.resolve()):
            errors.append(f"Archive path escapes archive: {entry['archived']}")
            continue
        if not path.is_file() or hashlib.sha256(path.read_bytes()).hexdigest() != entry["sha256"]:
            errors.append(f"Archive content missing/changed: {entry['archived']}")
    # Metadata is optional in fresh checkouts. If present, it must still be original.
    for entry in metadata:
        path = root / entry["archived"]
        if path.exists() and hashlib.sha256(path.read_bytes()).hexdigest() != entry["sha256"]:
            errors.append(f"Local metadata changed: {entry['archived']}")
    ignored = subprocess.run(["git", "check-ignore", "--stdin"], cwd=root,
                             input="\n".join(e["archived"] for e in content) + "\n",
                             capture_output=True, text=True)
    if ignored.returncode not in (0, 1):
        errors.append("Cannot validate archive Git eligibility: " + ignored.stderr.strip())
    elif ignored.stdout.strip():
        errors.append("Content inventory includes Git-ignored files: " + ignored.stdout.strip())
    for name in ("docs-contract.instructions.md", "runtime-go.instructions.md"):
        text = (root / ".github/instructions" / name).read_text()
        if "specs/012-apple-only-experience" not in text:
            errors.append(f"Missing current authority: {name}")
        if re.search(r"active (?:project |runtime )?compose file", text, re.I):
            errors.append(f"Obsolete Compose authority: {name}")
    docs_instruction = (root / ".github/instructions/docs-contract.instructions.md").read_text()
    scope = re.search(r"^applyTo: (.*)$", docs_instruction, re.M)
    if not scope or "archive/" in scope[1]:
        errors.append("Active docs instruction must exclude archive scope")
    archive_instruction = (root / ".github/instructions/archive.instructions.md").read_text()
    if '"archive/**,previous-version-archive/**"' not in archive_instruction:
        errors.append("Archive instruction does not cover both archive roots")
    print(f"Checked {len(tasks)} tasks, {len(requirements)} requirements, {len(outcomes)} outcomes, "
          f"{len(content)} content hashes and {len(docs)} linked documents.")
    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    args = parser.parse_args()
    try:
        errors = verify(args.root.resolve())
    except (OSError, ValueError, KeyError) as exc:
        errors = [str(exc)]
    if errors:
        print("\n".join("FAIL: " + error for error in errors), file=sys.stderr)
        return 1
    print("PASS: planning structure/provenance checks; runtime and user-study gates are NOT certified.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
