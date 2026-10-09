package backend

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	mapping "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/http/backend/generated"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1"
)

// requestError carries a transport-safe failure through HTTP request translation.
type requestError struct {
	code   string
	status int
}

func (e requestError) Error() string { return e.code }
func newID() string {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		panic(err)
	}
	id[6] = id[6]&15 | 64
	id[8] = id[8]&63 | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[:4], id[4:6], id[6:8], id[8:10], id[10:])
}
func translate(r *http.Request, owner, path string, operation *openapi3.Operation) (pb.Request, error) {
	request, err := pb.NewRequest(owner, newID())
	if err != nil {
		return nil, err
	}
	if r.Method == http.MethodPost {
		id, key, correlation := r.PathValue("id"), r.Header.Get("Idempotency-Key"), r.Header.Get("X-Correlation-ID")
		if correlation == "" {
			correlation = key
		}
		if !rpc.ValidID(id) || !rpc.ValidID(key) || !rpc.ValidID(correlation) {
			return nil, requestError{"invalid_id", 400}
		}
		version, err := strconv.ParseUint(strings.Trim(r.Header.Get("If-Match"), `"`), 10, 64)
		if err != nil {
			return nil, requestError{"expected_version_required", 428}
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, requestError{"invalid_request", 400}
		}
		var object map[string]any
		if err = json.Unmarshal(data, &object); err != nil || object == nil {
			return nil, requestError{"invalid_request", 400}
		}
		schema := operation.RequestBody.Value.Content["application/json"].Schema.Value
		if err = schema.VisitJSON(object); err != nil {
			return nil, requestError{"invalid_request", 400}
		}
		command, err := mapping.Command(operation.OperationID, data, &pb.CommandMetadata{CommandId: key, AggregateId: id, ExpectedVersion: &version, CorrelationId: correlation})
		if err != nil {
			return nil, requestError{"invalid_request", 400}
		}
		if err = rpc.SetPayload(request, "command", command); err != nil {
			return nil, err
		}
	} else {
		var id string
		var page *pb.PageRequest
		if strings.Contains(path, "{id}") {
			id = r.PathValue("id")
			if !rpc.ValidID(id) {
				return nil, requestError{"invalid_id", 400}
			}
		} else {
			values, err := pagination(r, strings.TrimPrefix(path, "/api"))
			if err != nil {
				return nil, err
			}
			if values != nil {
				page = &pb.PageRequest{Limit: uint32(values["limit"].(int))}
				if after, ok := values["after"].(string); ok {
					page.After = &after
				}
			}
		}
		query, err := mapping.Query(operation.OperationID, id, page)
		if err != nil {
			return nil, err
		}
		if err = rpc.SetPayload(request, "query", query); err != nil {
			return nil, err
		}
	}
	return request, nil
}
func pagination(r *http.Request, resource string) (map[string]any, error) {
	values := r.URL.Query()
	if !values.Has("limit") && !values.Has("cursor") {
		return nil, nil
	}
	invalid := requestError{"invalid_pagination", 400}
	raw := values.Get("limit")
	limit, err := strconv.Atoi(raw)
	if err != nil || strconv.Itoa(limit) != raw || limit < 1 || limit > 100 || len(values["limit"]) != 1 || len(values["cursor"]) > 1 {
		return nil, invalid
	}
	page := map[string]any{"limit": limit}
	if values.Has("cursor") {
		cursor := values.Get("cursor")
		if len(cursor) > 1024 {
			return nil, invalid
		}
		data, err := base64.RawURLEncoding.DecodeString(cursor)
		prefix := "1|" + resource + "|"
		if err != nil || base64.RawURLEncoding.EncodeToString(data) != cursor || !bytes.HasPrefix(data, []byte(prefix)) {
			return nil, invalid
		}
		after := string(data[len(prefix):])
		if !rpc.ValidID(after) {
			return nil, invalid
		}
		page["after"] = after
	}
	return page, nil
}
func translateReply(request pb.Request, reply pb.Reply, path string) (int, any, error) {
	if reply == nil || reply.GetContractVersion() != 1 || reply.GetContext() != request.GetContext() || reply.GetRequestId() != request.GetRequestId() {
		return 0, nil, fmt.Errorf("invalid reply identity")
	}
	if rejected := reply.GetError(); rejected != nil {
		status := 400
		switch rejected.Code {
		case "not_found":
			status = 404
		case "temporarily_unavailable":
			status = 503
		case "invalid_request", "invalid_pagination", "invalid_id":
		default:
			return 0, nil, fmt.Errorf("unknown technical failure")
		}
		object := map[string]any{"code": rejected.Code}
		if rejected.Message != "" {
			object["message"] = rejected.Message
		}
		return status, object, nil
	}
	if pb.Command(request) != nil {
		if reply.GetOutcome() == nil || reply.GetOutcome().AggregateId != pb.Command(request).GetMetadata().AggregateId {
			return 0, nil, fmt.Errorf("invalid command outcome")
		}
		status := 200
		outcome := reply.GetOutcome()
		object := map[string]any{"aggregateId": outcome.AggregateId, "version": outcome.Version, "status": outcome.Status}
		rejection := outcome.Rejection
		if rejection != nil {
			object["rejection"] = map[string]any{"code": rejection.Code, "message": rejection.Message}
		}
		if rejection != nil {
			status = 422
			switch rejection.Code {
			case "version_conflict", "idempotency_conflict":
				status = 409
			case "not_found":
				status = 404
			}
		}
		return status, object, nil
	}
	object, paged, after, err := mapping.QueryReply(request, reply)
	if err != nil {
		return 0, nil, err
	}
	if !paged {
		return 200, object, nil
	}
	result := map[string]any{"items": object, "nextCursor": nil}
	if after != nil {
		if !rpc.ValidID(*after) {
			return 0, nil, fmt.Errorf("invalid next identity")
		}
		result["nextCursor"] = base64.RawURLEncoding.EncodeToString([]byte("1|" + strings.TrimPrefix(path, "/api") + "|" + *after))
	}
	return 200, result, nil
}
