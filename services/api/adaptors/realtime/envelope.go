package realtime

import "google.golang.org/protobuf/encoding/protowire"

// Validate only the published transport envelope. Snapshot business values remain
// opaque and the original persisted bytes are forwarded without re-encoding.
// Field numbers belong to contracts/realtime/proto/cafe/realtime/v1/realtime.proto.
func validEnvelope(body []byte, owner, id string) bool {
	if id == "" {
		return false
	}
	strings := map[protowire.Number]string{}
	numbers := map[protowire.Number]uint64{}
	seen := map[protowire.Number]bool{}
	snapshots := 0
	for len(body) > 0 {
		field, kind, n := protowire.ConsumeTag(body)
		if n < 0 || seen[field] {
			return false
		}
		seen[field] = true
		body = body[n:]
		switch field {
		case 1, 3, 4, 5:
			if kind != protowire.BytesType {
				return false
			}
			value, consumed := protowire.ConsumeString(body)
			if consumed < 0 {
				return false
			}
			strings[field], n = value, consumed
		case 2, 6:
			if kind != protowire.VarintType {
				return false
			}
			value, consumed := protowire.ConsumeVarint(body)
			if consumed < 0 {
				return false
			}
			numbers[field], n = value, consumed
		default:
			if field >= 10 && field <= 17 {
				if kind != protowire.BytesType {
					return false
				}
				snapshots++
			}
			n = protowire.ConsumeFieldValue(field, kind, body)
			if n < 0 {
				return false
			}
		}
		body = body[n:]
	}
	return strings[1] == id && strings[3] == owner && strings[4] != "" && strings[5] != "" && numbers[2] == 1 && numbers[6] > 0 && snapshots == 1
}
