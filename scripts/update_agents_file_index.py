#!/usr/bin/env python3
# Copyright 2025 Marc-Antoine Ruel. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

"""AGENTS.md index generation and first-line summary validation.

To opt-in a directory, add these two markers to its AGENTS.md:

    <!-- BEGIN FILE INDEX -->
    <!-- END FILE INDEX -->

The script auto-discovers all non-ignored AGENTS.md files that contain the
markers, generates a file index from first-line comments, and injects it between
the markers. Repository-relative glob patterns in .agents-index-ignore exclude
content that should not require source-file summaries. Directories without
markers, including the repository root, are left unchanged.
"""

import argparse
import fnmatch
import os
import re
import subprocess
import sys

SUMMARY_SCAN_LINES = 20

INDEX_SUMMARY_GUIDANCE = f"""\
Add a one-line index summary (the first meaningful comment line):
  - Keep it under 120 characters; favor compact keywords over sentences.
  - Do not wrap it: later lines are ignored.
  - It must be within the first {SUMMARY_SCAN_LINES} lines.
  - Do not group it in license notice. Add an empty line to disambiguate.
  - Omit the filename and generic labels such as "Goal"."""

EXAMPLES = """\
Index summary examples:
  CSS:
    /* task-list panel, filters, task state */
    body { color: #fff; }
  Markdown:
    ---
    # development setup, build, test workflow
    title: Development
    ---
    # User visible title
  Python:
    #!/usr/bin/env python3
    # Copyright 2026 Example Corp. All rights reserved.

    # trace parsing, bottleneck diagnostics, report output
    '''Foo bar'''
  Go:
    // Copyright 2026 Example Corp. All rights reserved.

    // API response cache, persistent storage, TTL

    // Package cache manages persistent API response caching.
    package cache"""

HELP_EPILOG = f"{INDEX_SUMMARY_GUIDANCE}\n\n{EXAMPLES}"


def get_git_files() -> list[str]:
    """Return tracked and non-ignored untracked files, or [] on failure."""
    try:
        result = subprocess.run(
            ["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"],
            capture_output=True,
            text=True,
            check=True,
        )
        return [f for f in result.stdout.split("\0") if f and os.path.lexists(f)]
    except (subprocess.CalledProcessError, FileNotFoundError) as e:
        print(f"Error listing git files: {e}", file=sys.stderr)
        return []


def _py_docstring(lines, i):
    """Extract the description from a Python triple-quoted docstring starting at lines[i].

    Returns the description string, or "" if none found.
    """
    sline = lines[i].strip()
    quote = sline[:3]
    # Single-line docstring: """text"""
    if sline.endswith(quote) and len(sline) > 6:
        return sline[3:-3].strip()
    # Multi-line: return the first content line.
    content = sline[3:].strip()
    if content:
        return content
    # Opening quotes on their own line; use next non-empty line.
    for j in range(i + 1, len(lines)):
        if lines[j] and lines[j].strip():
            return lines[j].strip()
    return ""


def _c_style_block_comment(lines: list[str], i: int) -> str:
    """Extract the first meaningful line from a C-style block comment."""
    for line_index, line in enumerate(lines[i:]):
        text = line.strip()
        if line_index == 0:
            text = text[2:].strip()
        end = text.find("*/")
        if end >= 0:
            text = text[:end].strip()
        text = text.lstrip("*").strip()
        if text:
            return text
        if end >= 0:
            break
    return ""


def get_file_description(filepath):
    """Return the description for a file, or None if not applicable.

    Returns None if the file type has no comment convention (or is explicitly
    excluded). Returns "" if the file supports comments but has no description,
    which is treated as an error by callers.
    """
    # Glob patterns mapping filenames to supported comment styles. None skips the file.
    comment_styles = {
        "*.min.*": None,
        "*.d.ts": None,
        "pnpm-lock.yaml": None,
        "*.c": ("//", "/*"),
        "*.cc": ("//", "/*"),
        "*.cjs": ("//", "/*"),
        "*.cpp": ("//", "/*"),
        "*.css": ("/*",),
        "*.cxx": ("//", "/*"),
        "*.go": ("//", "/*"),
        "*.h": ("//", "/*"),
        "*.hh": ("//", "/*"),
        "*.hpp": ("//", "/*"),
        "*.hxx": ("//", "/*"),
        "*.js": ("//", "/*"),
        "*.kt": ("//", "/*"),
        "*.md": ("#",),
        "*.mjs": ("//", "/*"),
        "*.py": ("#",),
        "*.service": ("#",),
        "*.sh": ("#",),
        "*.swift": ("//", "/*"),
        "*.toml": ("#",),
        "*.ts": ("//", "/*"),
        "*.tsx": ("//", "/*"),
        "*.yaml": ("#",),
        "*.yml": ("#",),
        "Dockerfile*": ("#",),
        "Makefile": ("#",),
    }
    if os.path.islink(filepath):
        return None
    fname = os.path.basename(filepath)
    styles = next(
        (style for pat, style in comment_styles.items() if fnmatch.fnmatch(fname, pat)),
        None,
    )
    if not styles:
        return None  # unrecognised extension or explicitly excluded pattern
    prefix = styles[0]
    with open(filepath, encoding="utf-8") as f:
        lines = [f.readline() for _ in range(SUMMARY_SCAN_LINES)]
    in_copyright = False
    for i, line in enumerate(lines):
        if not line:
            break
        sline = line.strip()
        if not sline:
            in_copyright = False
            continue
        if fname.endswith(".py") and sline.startswith(('"""', "'''")):
            return _py_docstring(lines, i)
        if "/*" in styles and sline.startswith("/*"):
            return _c_style_block_comment(lines, i)
        # Skip common directives/metadata that aren't descriptions.
        if sline.startswith(f"{prefix}go:"):
            continue
        if sline.startswith(f"{prefix} +build"):
            continue
        if sline.startswith(f"{prefix} nolint"):
            continue
        if sline.startswith(f"{prefix} swift-tools-version:"):
            continue
        if sline.startswith(f"{prefix} ///"):
            continue
        # Skip YAML front-matter delimiters and shebangs.
        if sline == "---" or sline.startswith("#!"):
            continue
        # Skip copyright headers and their continuation lines (until blank).
        if in_copyright:
            continue
        if " copyright " in sline.lower():
            in_copyright = True
            continue
        # Skip PEP 723 inline script metadata fields (requires-python, dependencies, etc.)
        # that appear between # /// script and # /// markers.
        if fname.endswith(".py") and sline.startswith(
            (f"{prefix} requires-", f"{prefix} dependencies"),
        ):
            continue
        if sline.startswith(prefix):
            comment = sline[len(prefix) :].strip()
            if not comment:
                continue
            return comment
        # Hit code before a comment.
        return ""
    return ""


def discover_configs(all_files):
    """Auto-discover workspace roots from AGENTS.md files that contain a file index marker.

    Returns a dict mapping target AGENTS.md path to its set of excluded child directories.
    """
    candidates = sorted(f for f in all_files if os.path.basename(f) == "AGENTS.md")
    configs = {}
    for f in candidates:
        with open(f, encoding="utf-8") as fh:
            if "<!-- BEGIN FILE INDEX -->" in fh.read():
                configs[f] = set()
    # For each config, find child workspaces and add them to exclude_dirs.
    for target, exclude in configs.items():
        root = os.path.dirname(target)
        prefix = root + "/" if root else ""
        for other_target in configs:
            oroot = os.path.dirname(other_target)
            if oroot == root:
                continue
            if prefix:
                if not oroot.startswith(prefix):
                    continue
                child_rel = oroot[len(prefix) :]
            else:
                child_rel = oroot
            exclude.add(child_rel)
    return configs


def get_index_ignores() -> list[str]:
    """Read repository-relative exclusion patterns for non-source content."""
    path = ".agents-index-ignore"
    if not os.path.exists(path):
        return []
    with open(path, encoding="utf-8") as f:
        lines = (line.strip() for line in f)
        return [line for line in lines if line and not line.startswith("#")]


def generate_index(target, exclude, all_files, all_configs, ignored):
    """Generate the file index for target, returning (content, missing) where
    missing is a list of files that support comments but have no description."""
    root_dir = os.path.dirname(target)
    files_found = []
    missing = []
    for filepath in all_files:
        if any(fnmatch.fnmatch(filepath, pattern) for pattern in ignored):
            continue
        # Skip own AGENTS.md.
        if filepath == target:
            continue
        # Scope to root_dir.
        if root_dir:
            if not filepath.startswith(root_dir + "/"):
                continue
            relpath = filepath[len(root_dir) + 1 :]
        else:
            relpath = filepath
        # Check excluded subdirectories, but let sub-workspace AGENTS.md through.
        rel_parts = relpath.replace("\\", "/").split("/")
        if "testdata" in rel_parts:
            continue
        # A file is excluded if its path starts with any excluded prefix.
        excluded = any(relpath.startswith(ex + "/") or relpath == ex for ex in exclude)
        if excluded and filepath not in all_configs:
            continue
        desc = get_file_description(filepath)
        if desc is None:
            continue  # file type has no comment convention
        if desc == "":
            missing.append(relpath)
        else:
            files_found.append((relpath, desc))
    desc = "Autogenerated from first-line comments. Run scripts/update_agents_file_index.py to refresh."
    lines = ["## File Index", "", desc, ""]
    for path, comment in sorted(files_found):
        lines.append(f"- `{path}`: {comment}")
    return "\n".join(lines), missing


def update_markdown(target_file: str, content: str, check: bool) -> bool:
    """Update or check the file index in target_file. Returns True if a change was made (or needed)."""
    if not os.path.exists(target_file):
        print(f"Warning: {target_file} not found, skipping.")
        return False
    start = "<!-- BEGIN FILE INDEX -->"
    end = "<!-- END FILE INDEX -->"
    with open(target_file, encoding="utf-8") as f:
        original = f.read()
    new_section = f"{start}\n{content}\n{end}"
    if start in original and end in original:
        pattern = re.compile(f"{re.escape(start)}.*?{re.escape(end)}", re.DOTALL)
        updated = pattern.sub(new_section, original)
    else:
        if original.strip():
            updated = original.rstrip() + "\n\n" + new_section + "\n"
        else:
            updated = new_section + "\n"
    if updated == original:
        return False
    if check:
        print(
            f"Error: {target_file} file index is out of date. Run scripts/update_agents_file_index.py to fix.",
            file=sys.stderr,
        )
        return True
    with open(target_file, "w", encoding="utf-8") as f:
        f.write(updated)
    print(f"Updated: {target_file}")
    return False


def report_missing_index_summaries(missing: list[str]) -> None:
    """Print coding-agent guidance for files missing an index summary."""
    print("Error: the following files are missing an index summary:", file=sys.stderr)
    for filepath in sorted(missing):
        print(f"  {filepath}", file=sys.stderr)
    print(INDEX_SUMMARY_GUIDANCE, file=sys.stderr)
    print(
        "Run scripts/update_agents_file_index.py --help for examples.",
        file=sys.stderr,
    )


def main() -> int:
    """Parse arguments and update (or check) the file indexes."""
    parser = argparse.ArgumentParser(
        description=__doc__,
        epilog=HELP_EPILOG,
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="check that indexes are up to date without modifying files (exit 1 if not)",
    )
    args = parser.parse_args()

    all_files = get_git_files()
    if not all_files:
        print("No files found in git repository.")
        return 1
    ret = 0
    configs = discover_configs(all_files)
    ignored = get_index_ignores()
    all_missing = []
    indexes_out_of_date = False
    for target, exclude in configs.items():
        content, missing = generate_index(target, exclude, all_files, configs, ignored)
        if update_markdown(target, content, check=args.check):
            ret = 1
            indexes_out_of_date = True
        all_missing.extend(missing)
    if all_missing:
        report_missing_index_summaries(all_missing)
        ret = 1
    elif args.check and indexes_out_of_date:
        print(INDEX_SUMMARY_GUIDANCE, file=sys.stderr)
        print(
            "Run scripts/update_agents_file_index.py --help for examples.",
            file=sys.stderr,
        )
    return ret


if __name__ == "__main__":
    sys.exit(main())
