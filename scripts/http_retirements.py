"""Exact development-only operation retirements accepted in ADR 0016.

Keep this list fixed: a new retirement requires its own reviewed decision.
Public API routes and owner health/diagnostics retain full compatibility checks.
"""
from http_compatibility import changes

RETIRED_STOREFRONT_OPERATIONS = frozenset([
    'GET /v1/menu/drinks',
    'GET /v1/menu/drinks/{id}',
    'POST /v1/menu/drinks/{id}',
    'POST /v1/menu/drinks/{id}/revise',
    'POST /v1/menu/drinks/{id}/publish',
    'GET /v1/menu/editions',
    'GET /v1/menu/editions/{id}',
    'POST /v1/menu/editions/{id}',
    'POST /v1/menu/editions/{id}/offers',
    'POST /v1/menu/editions/{id}/prices',
    'POST /v1/menu/editions/{id}/publish',
    'GET /v1/ordering/orders',
    'GET /v1/ordering/orders/{id}',
    'POST /v1/ordering/orders/{id}',
    'POST /v1/ordering/orders/{id}/lines',
    'POST /v1/ordering/orders/{id}/quantities',
    'POST /v1/ordering/orders/{id}/place',
])


def compatible_changes(name, before, after):
    errors = changes(before, after)
    if name == 'storefront':
        retired = {operation+': removed operation' for operation in RETIRED_STOREFRONT_OPERATIONS}
        return [error for error in errors if error not in retired]
    return errors
