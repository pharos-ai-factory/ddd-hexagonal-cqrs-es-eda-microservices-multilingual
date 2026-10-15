package realtime

import (
	"google.golang.org/protobuf/encoding/protowire"
	"testing"
)

func TestRootlessRejectionRetainsOwnerAndPayloadChecks(t *testing.T) {
	var header []byte
	for _, field := range []struct {
		number protowire.Number
		value  string
	}{{1, "publication-id"}, {3, "ordering"}, {4, "order"}, {5, "missing-root"}} {
		header = protowire.AppendString(protowire.AppendTag(header, field.number, protowire.BytesType), field.value)
	}
	header = protowire.AppendVarint(protowire.AppendTag(header, 2, protowire.VarintType), 1)
	// Omitted revision is zero, as encoded for an outcome without a root.
	rejected := protowire.AppendBytes(protowire.AppendTag(append([]byte{}, header...), 18, protowire.BytesType), []byte{10, 1, 'x'})
	if !validEnvelope(rejected, "ordering", "publication-id") {
		t.Fatal("rootless rejection refused")
	}
	if validEnvelope(rejected, "menu", "publication-id") || validEnvelope(rejected, "ordering", "other-id") {
		t.Fatal("rejection bypassed envelope identity")
	}
	snapshot := protowire.AppendBytes(protowire.AppendTag(header, 12, protowire.BytesType), []byte{10, 1, 'x'})
	if validEnvelope(snapshot, "ordering", "publication-id") {
		t.Fatal("zero-revision snapshot accepted")
	}
}
