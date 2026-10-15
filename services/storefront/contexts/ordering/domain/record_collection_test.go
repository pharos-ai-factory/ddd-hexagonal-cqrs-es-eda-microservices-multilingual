package domain

import "testing"

func TestCollectionRequiresPlacementAndMatchingCustomer(t *testing.T) {
	order, _ := New(id, id, id, "EUR")
	assertRejected(t, order.RecordCollection(id), "order_not_placed")
	_ = order.AddLine(lineID, Selection{"C1", "Coffee", 300}, 1)
	_ = order.Place()
	order, _ = Restore(order.Snapshot())
	assertRejected(t, order.RecordCollection(lineID), "customer_mismatch")
	if order.Snapshot().CollectionRecorded || len(order.Events()) != 0 {
		t.Fatal("rejection changed the order")
	}
	if err := order.RecordCollection(id); err != nil {
		t.Fatal(err)
	}
	if !order.Snapshot().CollectionRecorded || len(order.Events()) != 1 {
		t.Fatal("collection fact missing")
	}
	restored, err := Restore(order.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	if err = restored.RecordCollection(id); err != nil || len(restored.Events()) != 0 {
		t.Fatal("repeated collection changed the order")
	}
	assertRejected(t, restored.RecordCollection(lineID), "customer_mismatch")
	state := order.Snapshot()
	state.Status = "draft"
	if _, err = Restore(state); err == nil {
		t.Fatal("corrupt collection state accepted")
	}
}
