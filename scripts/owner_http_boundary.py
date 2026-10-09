"""Keep business HTTP at the API and owner HTTP operational (ADR 0016)."""
import re

OWNERS = {'storefront', 'operations', 'engagement'}


def violation(path, content):
    parts = path.split('/')
    if len(parts) < 3 or parts[0] != 'services' or parts[1] not in OWNERS:
        return None
    if '/contexts/' in path and '/adaptors/http/' in path:
        return 'owner contexts receive business requests through messaging adaptors'
    # Published HTTP paths belong to the API. Wire schemas, test fixtures and
    # outbound provider integrations have separate purposes and are excluded.
    if ('/contexts/' not in path and '/integrations/' not in path
            and path.endswith(('.go', '.py', '.ts'))
            and re.search(r'''["'](?:GET |POST |PUT |PATCH |DELETE |HEAD |OPTIONS )?/+(?:api/)?v\d+/''', content)):
        return 'owner HTTP is limited to health and diagnostics; business routes belong to the API'
    return None
