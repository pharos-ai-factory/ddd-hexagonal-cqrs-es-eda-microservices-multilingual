"""Keep public response shapes schema-owned and composition bound to their signatures."""

import re


def violation(path, content):
    if "generated" in path or "test" in path or not path.startswith("services/"):
        return None
    if "/application/" in path or "/foundation/" in path:
        shape = r"(?:class\s+\w*(?:View|Response|Outcome)\s*\(TypedDict\)|type\s+\w*(?:View|Response|Outcome)\s+(?:struct\s*\{|=\s*(?:Readonly<)?\{)|interface\s+\w+(?:View|Response)\s*\{)"
        if re.search(shape, content):
            return "response types and published read views must use generated contracts"
    if "/apps/" in path and re.search(r"\b(?:registry\.query|requests\.Query)\s*\(", content):
        return "query composition must use its generated response-bound registration"
    return None
