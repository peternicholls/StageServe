#!/usr/bin/env python3
"""Capture real, read-only CLI evidence; never starts or changes a runtime."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys


def capture(binary, output):
    commands = {
        "version": ["--version"],
        "help": ["--help"],
        "run-help": ["run", "--help"],
        "inspect-help": ["inspect", "--help"],
        "network-help": ["network", "--help"],
        "network-create-help": ["network", "create", "--help"],
        "volume-help": ["volume", "--help"],
        "system-status": ["system", "status", "--format", "json"],
        "list": ["list", "--all", "--format", "json"],
        "network-list": ["network", "list", "--format", "json"],
        "volume-list": ["volume", "list", "--format", "json"],
    }
    # Refuse to overwrite prior evidence, including a symlink destination.
    output.mkdir(mode=0o700, parents=True, exist_ok=False)
    records = []
    for name, args in commands.items():
        try:
            result = subprocess.run([binary, *args], capture_output=True, timeout=30)
            stdout, stderr, code = result.stdout, result.stderr, result.returncode
        except subprocess.TimeoutExpired as error:
            stdout, stderr, code = error.stdout or b"", error.stderr or b"", None
        record = {"command": args, "exit_code": code,
                  "timed_out": code is None}
        for suffix, data in (("stdout", stdout), ("stderr", stderr)):
            filename = f"{name}.{suffix}"
            path = output / filename
            path.write_bytes(data)
            path.chmod(0o600)
            record[suffix] = {"file": filename,
                              "sha256": hashlib.sha256(data).hexdigest()}
        records.append(record)
    metadata = {"captured_at": datetime.datetime.now(datetime.timezone.utc).isoformat(),
                "host": " ".join(os.uname()[:3]), "architecture": os.uname().machine,
                "binary": binary, "commands": records,
                "qualification": "unqualified; help/inventory capture only"}
    path = output / "capture.json"
    path.write_text(json.dumps(metadata, indent=2) + "\n")
    path.chmod(0o600)
    return all(record["exit_code"] == 0 for record in records)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", default="container")
    parser.add_argument("--output", required=True, type=Path,
                        help="new private directory; review before committing fixtures")
    args = parser.parse_args()
    binary = shutil.which(args.binary)
    if binary is None:
        parser.error("Apple container CLI not found; no evidence captured")
    os.umask(0o077)
    try:
        success = capture(binary, args.output)
    except OSError as error:
        print(f"capture failed: {error}", file=sys.stderr)
        return 1
    print(f"Evidence saved to {args.output}; review for sensitive data before sharing.")
    return 0 if success else 1


if __name__ == "__main__":
    sys.exit(main())
