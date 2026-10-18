#!/usr/bin/env python3
# Copyright 2026 Marc-Antoine Ruel. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

"""Lint git-tracked files for unexpected binary data and executable files."""

import os
import subprocess
import sys

ALLOWED_BINARY_EXTENSIONS = {
    ".aac",
    ".avif",
    ".br",
    ".flac",
    ".gif",
    ".ico",
    ".jar",
    ".jpg",
    ".mov",
    ".mp3",
    ".mp4",
    ".ogg",
    ".pdf",
    ".png",
    ".svg",
    ".syso",
    ".wasm",
    ".wav",
    ".webm",
    ".webp",
    ".zst",
}
ALLOWED_EXECUTABLE_EXTENSIONS = {".desktop"}


def is_binary(path: str) -> bool:
    """Return whether path contains a null byte near its start."""
    if os.path.islink(path) or not os.path.isfile(path):
        return False
    with open(path, "rb") as f:
        return b"\0" in f.read(1024)


def has_executable_preamble(path: str) -> bool:
    """Return whether path starts with a supported executable declaration."""
    with open(path, "rb") as f:
        first_line = f.readline(256)
    # Go generator files use this shell/Go polyglot preamble to run with go run.
    return first_line.startswith((b"#!/", b"///usr/bin/true; exec /usr/bin/env go run "))


def tracked_files() -> list[tuple[str, bool]]:
    """Return tracked paths and Git executable modes, including paths with whitespace."""
    raw = subprocess.check_output(("git", "ls-files", "--stage", "-z"))
    files: list[tuple[str, bool]] = []
    for record in raw.split(b"\0"):
        if record:
            metadata, path = record.split(b"\t", 1)
            mode = metadata.split(b" ", 1)[0]
            files.append((os.fsdecode(path), mode == b"100755"))
    return files


def main() -> int:
    """Report tracked files that are binary or executable without a declaration."""
    unexpected_binaries: list[str] = []
    unexpected_executables: list[str] = []
    for path, executable in tracked_files():
        if is_binary(path) and os.path.splitext(path)[1].lower() not in ALLOWED_BINARY_EXTENSIONS:
            unexpected_binaries.append(path)
        if (
            executable
            and os.path.isfile(path)
            and not os.path.islink(path)
            and os.path.splitext(path)[1].lower() not in ALLOWED_EXECUTABLE_EXTENSIONS
            and not has_executable_preamble(path)
        ):
            unexpected_executables.append(path)

    if unexpected_binaries:
        print("Unexpected binary files:", file=sys.stderr)
        print(*(f"  {path}" for path in unexpected_binaries), sep="\n", file=sys.stderr)
    if unexpected_executables:
        print(
            "Executable files without a shebang, Go launcher preamble, or allowed extension:",
            file=sys.stderr,
        )
        print(*(f"  {path}" for path in unexpected_executables), sep="\n", file=sys.stderr)
    return int(bool(unexpected_binaries or unexpected_executables))


if __name__ == "__main__":
    sys.exit(main())
