"""Decision 0029 records existing successful-response presence guarantees once."""

from copy import deepcopy

# These fields were always supplied by the owner reply mappers. The new annotation
# records that producer guarantee; field identity and wire representation stay checked.
REQUIRED_OUTPUTS = {
    **{
        f"cafe.{owner}.requests.v1.Loaded{resource}": "3"
        for owner, resource in (
            ("menu", "Drink"),
            ("menu", "Edition"),
            ("ordering", "Order"),
            ("preparation", "Ticket"),
            ("collection", "Pickup"),
            ("loyalty", "Account"),
            ("loyalty", "Reward"),
            ("communication", "Notification"),
        )
    },
    "cafe.ordering.requests.v1.Line": "2",
}
REQUIRED_OUTPUT_TRUE = "d8f31801"  # FieldOptions extension 51003, boolean true.


def accepted_guarantees(old, new):
    """Accept only the initial annotation on the nine reviewed existing fields."""
    result = deepcopy(old)
    for name, number in REQUIRED_OUTPUTS.items():
        before = result.get(name, {}).get("fields", {}).get(number)
        after = new.get(name, {}).get("fields", {}).get(number)
        if before and after and before["options"] == "" and after["options"] == REQUIRED_OUTPUT_TRUE:
            before["options"] = REQUIRED_OUTPUT_TRUE
    return result
