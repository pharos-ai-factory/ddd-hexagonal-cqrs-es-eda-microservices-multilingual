"""Allow application dependencies on generated plain owner inputs and published events."""


def plain_request_or_event(service, owner, target):
    roots = {
        "operations": (
            "operations/contracts/events/generated",
            "services/operations/src/operations/contracts/events/generated",
            f"operations/contracts/requests/generated/{owner}",
            "operations/contracts/requests/generated/shared",
            f"services/operations/src/operations/contracts/requests/generated/{owner}",
            "services/operations/src/operations/contracts/requests/generated/shared",
        ),
        "engagement": (
            "services/engagement/src/contracts/events/generated",
            f"services/engagement/src/contracts/requests/generated/{owner}.js",
            "services/engagement/src/contracts/requests/generated/shared.js",
        ),
    }
    return any(target == root or target.startswith(root + "/") for root in roots.get(service, ()))
