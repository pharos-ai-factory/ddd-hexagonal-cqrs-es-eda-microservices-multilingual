import pytest
from operations.contexts.collection.adaptors.http import inputs as collection_inputs

@pytest.mark.parametrize("value", [None, [], {}, {"code": None}, {"code": 123},
                                 {"code": "ABC123", "extra": True}])
def test_untyped_collection_bodies_cannot_enter_the_application(value: object) -> None:
    with pytest.raises(ValueError):
        collection_inputs.collect_order(value)
