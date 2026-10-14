"""Select repository source consistently for formatting and edit checks."""

from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[1]
SERVICES = ("storefront", "api", "operations", "engagement", "web")
FORMAT_SUFFIXES = {
    ".go",
    ".py",
    ".pyi",
    ".ts",
    ".tsx",
    ".mjs",
    ".mts",
    ".cts",
    ".json",
    ".yaml",
    ".yml",
    ".css",
    ".md",
}


def git_paths(*arguments, root=ROOT):
    output = subprocess.check_output(["git", *arguments, "-z"], cwd=root)
    return {name.decode() for name in output.split(b"\0") if name}


def inventory(root=ROOT):
    """Include new files; ignore local tools, dependencies and build artefacts."""
    return git_paths("ls-files", "--cached", "--others", "--exclude-standard", root=root)


def changed(root=ROOT):
    return git_paths("diff", "--name-only", "HEAD", root=root) | git_paths(
        "ls-files", "--others", "--exclude-standard", root=root
    )


def handwritten(name, root=ROOT):
    from check_architecture import GENERATED

    path = root / name
    if not path.is_file() or path.suffix not in FORMAT_SUFFIXES:
        return False
    if name in GENERATED or "generated" in path.parts or path.name == "next-env.d.ts":
        return False
    first_line = path.read_text().splitlines()[:1]
    return not first_line or "DO NOT EDIT" not in first_line[0]


def affected_services(names):
    """Shared contracts/tooling can affect every service; deletions count too."""
    result = set()
    for name in names:
        parts = Path(name).parts
        if len(parts) > 1 and parts[0] == "services" and parts[1] in SERVICES:
            result.add(parts[1])
        elif Path(name).suffix != ".md":
            return list(SERVICES)
    return [service for service in SERVICES if service in result]
