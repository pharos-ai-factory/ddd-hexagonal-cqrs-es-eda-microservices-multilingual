"""Keep feature composition on typed, direct message-to-handler registrations."""

import re


def violation(path: str, content: str) -> str | None:
    roots = (
        "services/storefront/apps/storefront/di/",
        "services/operations/src/operations/apps/composition/",
        "services/engagement/src/apps/composition/",
    )
    services = (
        "services/operations/src/operations/apps/service.py",
        "services/engagement/src/apps/service.ts",
    )
    if not (path.startswith(roots) or path in services):
        return None
    if path.endswith(("_test.go", ".test.ts")) or path.rsplit("/", 1)[-1].startswith("test_"):
        return None
    if re.search(r"\.queries\s*\(|\b\w*QueryEndpoints\b|\bRegisterQueries\b|\b\w+\.Queries\s*\(", content):
        return "register each query identifier, owner parser, handler and reply mapper directly"
    if re.search(
        r"(?:\.request|\bRegister(?:Projection)?|\.Query)\s*\([^,]+,\s*[\"']|\.query\s*\(\s*[\"']", content
    ):
        return "use generated command and query identifiers in registrations"
    if re.search(r"[\"'][a-z]+\.[a-z]+(?:-[a-z]+)+(?:\.domain)?[\"']", content):
        return "use generated event and subscription identifiers in composition"
    patterns = {
        ".py": r"\b(?:CommandExecutor|command_subscription)\b|\.command\s*\(",
        ".ts": r"\b(?:CommandExecutor|bindCommand|commandSubscription)\b|\.command\s*\(",
        ".go": r"\b\w+\.(?:Executor\s*\[|Bind(?:Projection)?\s*\(|Command\s*\()",
    }
    suffix = "." + path.rsplit(".", 1)[-1]
    if suffix in patterns and re.search(patterns[suffix], content):
        return "register commands with plain handler factories through the shared command registration API"
    return None
