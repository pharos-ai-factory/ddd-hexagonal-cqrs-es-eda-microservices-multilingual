package realtime

import (
	"regexp"

	"google.golang.org/protobuf/encoding/protowire"
)

// Validate only the published transport envelope. Snapshot business values remain
// opaque and the original persisted bytes are forwarded without re-encoding.
// Field numbers belong to contracts/shared/realtime/v1/realtime.proto.
func validatedChannel(body []byte, owner, id string) (string, bool) {
	if id == "" {
		return "", false
	}
	strings := map[protowire.Number]string{}
	numbers := map[protowire.Number]uint64{}
	seen := map[protowire.Number]bool{}
	var target []byte
	var payload protowire.Number
	for len(body) > 0 {
		field, kind, n := protowire.ConsumeTag(body)
		if n < 0 || seen[field] {
			return "", false
		}
		seen[field] = true
		body = body[n:]
		switch field {
		case 1, 3, 4, 5:
			if kind != protowire.BytesType {
				return "", false
			}
			value, consumed := protowire.ConsumeString(body)
			if consumed < 0 {
				return "", false
			}
			strings[field], n = value, consumed
		case 2, 6:
			if kind != protowire.VarintType {
				return "", false
			}
			value, consumed := protowire.ConsumeVarint(body)
			if consumed < 0 {
				return "", false
			}
			numbers[field], n = value, consumed
		case 7:
			if kind != protowire.BytesType {
				return "", false
			}
			value, consumed := protowire.ConsumeBytes(body)
			if consumed < 0 {
				return "", false
			}
			target, n = value, consumed
		default:
			if field >= 10 && field <= 18 {
				if kind != protowire.BytesType || payload != 0 {
					return "", false
				}
				payload = field
			}
			n = protowire.ConsumeFieldValue(field, kind, body)
			if n < 0 {
				return "", false
			}
		}
		body = body[n:]
	}
	if strings[1] != id || strings[3] != owner || strings[4] == "" || strings[5] == "" || numbers[2] != 1 || payload == 0 || (numbers[6] == 0 && payload != 18) {
		return "", false
	}
	return targetChannel(target, owner, payload)
}

func targetChannel(body []byte, owner string, payload protowire.Number) (string, bool) {
	field, kind, n := protowire.ConsumeTag(body)
	if n < 0 || kind != protowire.BytesType {
		return "", false
	}
	value, consumed := protowire.ConsumeString(body[n:])
	if consumed < 0 || n+consumed != len(body) {
		return "", false
	}
	switch field {
	case 1:
		if value != owner || !ownerPayload(owner, payload) {
			return "", false
		}
		return "cafe:" + owner, true
	case 2:
		if payload != 18 || !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(value) {
			return "", false
		}
		return "principal:" + value, true
	default:
		return "", false
	}
}

func ownerPayload(owner string, payload protowire.Number) bool {
	allowed := map[string]map[protowire.Number]bool{
		"menu":          {10: true, 11: true},
		"ordering":      {12: true},
		"preparation":   {13: true},
		"collection":    {14: true},
		"loyalty":       {15: true, 16: true},
		"communication": {17: true},
	}
	return allowed[owner][payload]
}
