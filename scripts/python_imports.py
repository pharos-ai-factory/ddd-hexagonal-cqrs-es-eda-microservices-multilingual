"""Normalise Python import targets for repository architecture checks."""

import ast
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def python_imports(path, content):
    """Yield absolute and relative Python imports, including imported child modules."""
    for node in ast.walk(ast.parse(content)):
        if isinstance(node, ast.Import):
            yield from (item.name.replace(".", "/") for item in node.names)
        elif isinstance(node, ast.ImportFrom):
            if node.level:
                base = path.parent
                for _ in range(node.level - 1):
                    base = base.parent
                target = base.relative_to(ROOT).as_posix()
                if node.module:
                    target += "/" + node.module.replace(".", "/")
            else:
                target = (node.module or "").replace(".", "/")
            yield target
            # ImportFrom may name a child module, including `from .. import collection`.
            for item in node.names:
                if item.name != "*":
                    yield target + "/" + item.name.replace(".", "/")
