#!/bin/bash
# Copyright 2026 Marc-Antoine Ruel. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

# Checks that the staged snapshot has a current AGENTS.md file index.

set -euo pipefail

script_dir="$(CDPATH='' cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
readonly script_dir
repo_root="$(CDPATH='' cd -- "$script_dir/.." && pwd)"
readonly repo_root

cd -- "$repo_root"

if git diff --cached --quiet; then
  exit 0
fi
if ! git diff --quiet; then
  printf '%s\n' 'Stage all tracked changes before committing; the index check reads the worktree.' >&2
  exit 1
fi
untracked=$(git ls-files --others --exclude-standard)
if [[ -n "$untracked" ]]; then
  printf '%s\n' 'Stage or ignore untracked files before committing; the index check includes them:' "$untracked" >&2
  exit 1
fi
uv run --locked python scripts/lint_binaries.py
uv run --locked python scripts/update_agents_file_index.py --check
