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
	base := append([]byte{}, header...)
	principal := "00000000-0000-4000-8000-000000000001"
	target := protowire.AppendString(protowire.AppendTag(nil, 2, protowire.BytesType), principal)
	header = protowire.AppendBytes(protowire.AppendTag(header, 7, protowire.BytesType), target)
	// Omitted revision is zero, as encoded for an outcome without a root.
	rejected := protowire.AppendBytes(protowire.AppendTag(append([]byte{}, header...), 18, protowire.BytesType), []byte{10, 1, 'x'})
	if channel, ok := validatedChannel(rejected, "ordering", "publication-id"); !ok || channel != "principal:"+principal {
		t.Fatal("rootless rejection refused")
	}
	if _, ok := validatedChannel(rejected, "menu", "publication-id"); ok {
		t.Fatal("rejection bypassed envelope owner")
	}
	if _, ok := validatedChannel(rejected, "ordering", "other-id"); ok {
		t.Fatal("rejection bypassed envelope identity")
	}
	withRevision := protowire.AppendVarint(protowire.AppendTag(append([]byte{}, header...), 6, protowire.VarintType), 1)
	snapshot := protowire.AppendBytes(protowire.AppendTag(withRevision, 12, protowire.BytesType), []byte{10, 1, 'x'})
	if _, ok := validatedChannel(snapshot, "ordering", "publication-id"); ok {
		t.Fatal("principal-targeted snapshot accepted")
	}
	contextTarget := protowire.AppendString(protowire.AppendTag(nil, 1, protowire.BytesType), "ordering")
	contextHeader := protowire.AppendBytes(protowire.AppendTag(base, 7, protowire.BytesType), contextTarget)
	contextRejection := protowire.AppendBytes(protowire.AppendTag(contextHeader, 18, protowire.BytesType), []byte{10, 1, 'x'})
	if _, ok := validatedChannel(contextRejection, "ordering", "publication-id"); ok {
		t.Fatal("context-targeted rejection accepted")
	}
}
