"""Give reviewed use-case starters a schema fragment and generated input imports."""

import re

MODULE = (
    "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront"
)


def generated_inputs(files, service, context, name, role):
    typename = name + role
    for filename, content in files.items():
        if service == "storefront" and f"type {typename} struct" in content:
            content = re.sub(r"//[^\n]*\n(type " + typename + r" struct \{[^}]*\})\n", "", content)
            content = re.sub(r"(?<![\w.])" + typename + r"\b", "inputs." + typename, content)
            content = content.replace(
                "import (", f'import (\n inputs "{MODULE}/contracts/requests/model/{context}"'
            )
        elif service == "operations" and f"class {typename}" in content:
            content = re.sub(
                r"(?:@dataclass\(frozen=True\)\n)?class "
                + typename
                + r":?\b.*?(?=class "
                + typename
                + r"Handler)",
                "",
                content,
                flags=re.S,
            )
            content = content.replace("from typing import TypedDict\n", "").replace(
                "from dataclasses import dataclass\n", ""
            )
            content = f"from operations.contracts.requests.generated.{context} import {typename}\n" + content
            content = content.replace("query.identity", "query.id")
        elif service == "engagement" and f"export type {typename} =" in content:
            content = re.sub(
                r"/\*\*[^\n]*\*/\nexport type " + typename + r" = Readonly<\{[^}]*\}>;\n", "", content
            )
            content = (
                f"import type {{ {typename} }} from '../../../../contracts/requests/generated/{context}.js';\n"
                + content
            )
        files[filename] = content
    fields = (
        "  string id = 1;\n"
        if role == "Query"
        else "  // Define the business input fields and required presence rules.\n"
    )
    files["contract.proto.txt"] = (
        f"// Add to the owning {role.lower()} schema and its request-envelope oneof.\n"
        'import "cafe/requests/v1/validation.proto";\n'
        f'message {name} {{\n  option (cafe.requests.v1.contract) = {{ owner: "{context}" visibility: PUBLIC }};\n'
        + fields
        + "}\n"
    )
    files["generation.txt"] = (
        "Choose PUBLIC for a supported external capability or PRIVATE for an owner-only durable command.\n"
        "Place private commands in the owner internal_commands.proto envelope.\n"
        "Run pnpm generate:contracts before copying the handler; import its generated payload and identifier.\n"
        "Keep semantic validation and reply mapping in owner adaptors; mechanical input mappings are generated.\n"
    )
    return files
