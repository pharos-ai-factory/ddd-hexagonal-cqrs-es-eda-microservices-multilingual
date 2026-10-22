package realtime

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/openapi"
	check "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/tests/httpcontract"

	"google.golang.org/protobuf/encoding/protowire"
)

func envelope(owner, id string) []byte {
	var body []byte
	for _, field := range []struct {
		number protowire.Number
		value  string
	}{{1, id}, {3, owner}, {4, "root"}, {5, "aggregate-id"}} {
		body = protowire.AppendString(protowire.AppendTag(body, field.number, protowire.BytesType), field.value)
	}
	body = protowire.AppendVarint(protowire.AppendTag(body, 2, protowire.VarintType), 1)
	body = protowire.AppendVarint(protowire.AppendTag(body, 6, protowire.VarintType), 2)
	target := protowire.AppendString(protowire.AppendTag(nil, 1, protowire.BytesType), owner)
	body = protowire.AppendBytes(protowire.AppendTag(body, 7, protowire.BytesType), target)
	fields := map[string]protowire.Number{"menu": 10, "ordering": 12, "preparation": 13, "collection": 14, "loyalty": 15, "communication": 17}
	return protowire.AppendBytes(protowire.AppendTag(body, fields[owner], protowire.BytesType), []byte{10, 1, 'x'})
}

func publication(owner, id string) string {
	return fmt.Sprintf(`{"b64data":"%s","idempotency_key":"%s"}`, base64.StdEncoding.EncodeToString(envelope(owner, id)), id)
}

func rejectionEnvelope(owner, id, principal string) []byte {
	var body []byte
	for _, field := range []struct {
		number protowire.Number
		value  string
	}{{1, id}, {3, owner}, {4, "root"}, {5, "aggregate-id"}} {
		body = protowire.AppendString(protowire.AppendTag(body, field.number, protowire.BytesType), field.value)
	}
	body = protowire.AppendVarint(protowire.AppendTag(body, 2, protowire.VarintType), 1)
	target := protowire.AppendString(protowire.AppendTag(nil, 2, protowire.BytesType), principal)
	body = protowire.AppendBytes(protowire.AppendTag(body, 7, protowire.BytesType), target)
	return protowire.AppendBytes(protowire.AppendTag(body, 18, protowire.BytesType), []byte{10, 1, 'x'})
}

func TestGatewayRestrictsAuthoritiesAndPreservesBytes(t *testing.T) {
	type received struct {
		path, key, mode string
		body            []byte
	}
	var forwarded []received
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		forwarded = append(forwarded, received{r.URL.Path, r.Header.Get("X-API-Key"), r.Header.Get("X-Centrifugo-Error-Mode"), body})
		_, _ = w.Write([]byte(`{"result":{}}`))
	}))
	defer upstream.Close()
	config := Config{UpstreamURL: upstream.URL, UpstreamKey: "full-key", SessionKey: "sessions", PublisherKeys: map[string]string{}}
	for _, owner := range []string{"menu", "ordering", "preparation", "collection", "loyalty", "communication"} {
		config.PublisherKeys[owner] = owner + "-key"
	}
	handler, err := config.Handler()
	if err != nil {
		t.Fatal(err)
	}
	for _, owner := range []string{"menu", "ordering", "preparation", "collection", "loyalty", "communication"} {
		t.Run(owner, func(t *testing.T) {
			body := publication(owner, "publication-id")
			r := httptest.NewRequest("POST", "/api/publish", strings.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("X-API-Key", owner+"-key")
			check.Request(t, contract.Document("gateway"), r)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			check.Response(t, contract.Document("gateway"), r, w.Result())
			if w.Code != 200 {
				t.Fatalf("own publication status %d: %s", w.Code, w.Body.String())
			}
			last := forwarded[len(forwarded)-1]
			var upstreamBody struct {
				Channel string `json:"channel"`
				Data    string `json:"b64data"`
				ID      string `json:"idempotency_key"`
			}
			if err := json.Unmarshal(last.body, &upstreamBody); err != nil {
				t.Fatal(err)
			}
			if last.path != "/api/publish" || last.key != "full-key" || last.mode != "transport" || upstreamBody.Channel != "cafe:"+owner || upstreamBody.Data != base64.StdEncoding.EncodeToString(envelope(owner, "publication-id")) || upstreamBody.ID != "publication-id" {
				t.Fatal("forwarding changed authority or persisted bytes")
			}
		})
	}
	principal := "00000000-0000-4000-8000-000000000001"
	wire := rejectionEnvelope("ordering", "rejection-id", principal)
	body := fmt.Sprintf(`{"b64data":"%s","idempotency_key":"rejection-id"}`, base64.StdEncoding.EncodeToString(wire))
	r := httptest.NewRequest("POST", "/api/publish", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-API-Key", "ordering-key")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("principal publication status %d: %s", w.Code, w.Body.String())
	}
	var principalPublication struct {
		Channel string `json:"channel"`
	}
	if err := json.Unmarshal(forwarded[len(forwarded)-1].body, &principalPublication); err != nil || principalPublication.Channel != "principal:"+principal {
		t.Fatal("principal target did not derive its channel")
	}
	legacy := strings.TrimSuffix(publication("menu", "legacy-id"), "}") + `,"channel":"cafe:menu"}`
	r = httptest.NewRequest("POST", "/api/publish", strings.NewReader(legacy))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-API-Key", "menu-key")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("matching legacy channel status %d: %s", w.Code, w.Body.String())
	}
	for _, test := range []struct {
		name, method, path, key, body string
		status                        int
	}{
		{"publisher channel selection", "POST", "/api/publish", "menu-key", strings.TrimSuffix(publication("menu", "id"), "}") + `,"channel":"cafe:ordering"}`, 403},
		{"forged envelope context", "POST", "/api/publish", "menu-key", publication("ordering", "id"), 400},
		{"forged envelope identity", "POST", "/api/publish", "menu-key", strings.Replace(publication("menu", "id"), `"idempotency_key":"id"`, `"idempotency_key":"another"`, 1), 400},
		{"publisher disconnect", "POST", "/api/disconnect", "menu-key", `{"user":"operator"}`, 403},
		{"session publish", "POST", "/api/publish", "sessions", publication("menu", "id"), 403},
		{"no identity", "POST", "/api/publish", "", publication("menu", "id"), 403},
		{"master key not accepted", "POST", "/api/publish", "full-key", publication("menu", "id"), 403},
		{"history blocked", "POST", "/api/history", "menu-key", `{}`, 404},
		{"batch blocked", "POST", "/api", "menu-key", `{}`, 404},
		{"invalid method", "GET", "/api/publish", "menu-key", `{}`, 405},
		{"unknown publish option", "POST", "/api/publish", "menu-key", strings.TrimSuffix(publication("menu", "id"), "}") + `,"skip_history":true}`, 400},
		{"multiple bodies", "POST", "/api/publish", "menu-key", publication("menu", "id") + `{}`, 400},
		{"invalid wire", "POST", "/api/publish", "menu-key", `{"b64data":"AA==","idempotency_key":"id"}`, 400},
		{"oversized", "POST", "/api/publish", "menu-key", strings.Repeat("x", MaxRequestBytes+1), 400},
		{"disconnect options", "POST", "/api/disconnect", "sessions", `{"user":"operator","whitelist":["client"]}`, 400},
		{"empty disconnect", "POST", "/api/disconnect", "sessions", `{"user":""}`, 400},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := len(forwarded)
			r := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			r.Header.Set("X-API-Key", test.key)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if test.method == "POST" && (test.path == "/api/publish" || test.path == "/api/disconnect") {
				check.Response(t, contract.Document("gateway"), r, w.Result())
			}
			if w.Code != test.status || len(forwarded) != before {
				t.Fatalf("status=%d forwarded=%d", w.Code, len(forwarded)-before)
			}
		})
	}
	r = httptest.NewRequest("POST", "/api/disconnect", strings.NewReader(`{"user":"operator"}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-API-Key", "sessions")
	check.Request(t, contract.Document("gateway"), r)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	check.Response(t, contract.Document("gateway"), r, w.Result())
	last := forwarded[len(forwarded)-1]
	var disconnect struct {
		User       string `json:"user"`
		Disconnect struct {
			Code   int    `json:"code"`
			Reason string `json:"reason"`
		} `json:"disconnect"`
	}
	if json.Unmarshal(last.body, &disconnect) != nil || w.Code != 200 || last.path != "/api/disconnect" ||
		disconnect.User != "operator" || disconnect.Disconnect.Code != revocationDisconnectCode ||
		disconnect.Disconnect.Reason != revocationDisconnectReason {
		t.Fatal("session disconnect refused")
	}
	r = httptest.NewRequest("GET", "/healthz", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	check.Response(t, contract.Document("gateway"), r, w.Result())
	config.PublisherKeys["menu"] = config.SessionKey
	if _, err := config.Handler(); err == nil {
		t.Fatal("ambiguous identity accepted")
	}
}

func TestEnvelopeRejectsAmbiguousOrTruncatedMetadata(t *testing.T) {
	valid := envelope("menu", "id")
	if channel, ok := validatedChannel(valid, "menu", "id"); !ok || channel != "cafe:menu" {
		t.Fatal("valid transport rejected")
	}
	for _, wire := range [][]byte{
		valid[:len(valid)-1],
		append(bytes.Clone(valid), protowire.AppendString(protowire.AppendTag(nil, 3, protowire.BytesType), "ordering")...),
		append(bytes.Clone(valid), protowire.AppendBytes(protowire.AppendTag(nil, 11, protowire.BytesType), nil)...),
	} {
		if _, ok := validatedChannel(wire, "menu", "id"); ok {
			t.Fatal("ambiguous or truncated transport accepted")
		}
	}
}
