"""Allow application dependencies on generated plain owner inputs and published events."""


def plain_request_or_event(service, owner, target):
    roots = {
        "operations": (
            "operations/contracts/events/generated",
            "services/operations/src/operations/contracts/events/generated",
            f"operations/contracts/requests/generated/{owner}",
            f"operations/contracts/requests/generated/{owner}_responses",
            "operations/contracts/requests/generated/shared_responses",
            "operations/contracts/requests/generated/shared",
            f"services/operations/src/operations/contracts/requests/generated/{owner}",
            f"services/operations/src/operations/contracts/requests/generated/{owner}_responses",
            "services/operations/src/operations/contracts/requests/generated/shared_responses",
            "services/operations/src/operations/contracts/requests/generated/shared",
        ),
        "engagement": (
            "services/engagement/src/contracts/events/generated",
            f"services/engagement/src/contracts/requests/generated/{owner}.js",
            f"services/engagement/src/contracts/requests/generated/{owner}-responses.js",
            "services/engagement/src/contracts/requests/generated/shared-responses.js",
            "services/engagement/src/contracts/requests/generated/shared.js",
        ),
    }
    return any(target == root or target.startswith(root + "/") for root in roots.get(service, ()))
