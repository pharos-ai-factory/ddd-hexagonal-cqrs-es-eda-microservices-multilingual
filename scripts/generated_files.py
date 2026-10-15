"""Require repository-generated files to identify their ownership in the path."""

from pathlib import Path
import re

# Package managers and Next.js require these conventional filenames.
STANDARD_NAMES = {"pnpm-lock.yaml", "services/operations/uv.lock", "services/web/next-env.d.ts"}
HEADER = re.compile(
    r"^(?://|#|/\*+|\*)\s*(?:(?:Code )?[Gg]enerated (?:by|from)\b|[^\n]*DO NOT EDIT)",
    re.MULTILINE,
)


def clear_location(path):
    path = Path(path)
    return "generated" in path.name or "generated" in path.parts[:-1]


def violation(path, source, classified=False):
    """Check classified outputs and recognise generated headers outside the inventory."""
    header = "\n".join(source.splitlines()[:5])
    generated = classified or HEADER.search(header) or header.startswith('"""Generated plain published')
    if (
        generated
        and "generated" in Path(path).parts[:-1]
        and re.search(r"(^|[_.-])generated([_.-]|$)", Path(path).name)
    ):
        return "generated folder already identifies ownership; remove the redundant filename marker"
    if generated and path not in STANDARD_NAMES and not clear_location(path):
        return "generated file must have 'generated' in its filename or live under a 'generated' folder"
    return None
