package backend

import (
	"context"
	"encoding/base64"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/menu"
	contract "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/openapi"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPBoundaryRejectsMalformedCommandsBeforeDispatch(t *testing.T) {
	for _, body := range []string{`{}`, `{"code":"coffee","minor":null}`, `{"code":"coffee","minor":"0"}`, `{"code":"coffee","minor":0,"unknown":true}`, `[]`, `{"code":"coffee","minor":0} {}`, strings.Repeat("x", 65537)} {
		t.Run(body[:min(len(body), 40)], func(t *testing.T) {
			c := config()
			c.Owners = []string{"menu"}
			called := false
			c.Requests = caller(func(context.Context, pb.Request) (pb.Reply, error) { called = true; return nil, nil })
			w := request(t, backendHandler(c), "POST", "/api/v1/menu/editions/"+id+"/prices", body, map[string]string{"Authorization": "Bearer cli", "Idempotency-Key": id, "If-Match": "1"})
			if w.Code != 400 || called {
				t.Fatalf("malformed command dispatched: %d", w.Code)
			}
		})
	}
}
func TestPriceZeroRetainsPresenceAndCommandIdentityAcrossAttempts(t *testing.T) {
	operation := contract.Document("api").Paths.Value("/api/v1/menu/editions/{id}/prices").Post
	var first pb.Request
	for i := 0; i < 2; i++ {
		r := httptest.NewRequest("POST", "/api/v1/menu/editions/"+id+"/prices", strings.NewReader(`{"code":"coffee","minor":0}`))
		r.SetPathValue("id", id)
		r.Header.Set("Idempotency-Key", id)
		r.Header.Set("If-Match", "0")
		incoming, err := translate(r, "menu", "/api/v1/menu/editions/{id}/prices", operation)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := rpc.Validate(incoming, "menu"); err != nil {
			t.Fatal(err)
		}
		if incoming.(*menu.Request).GetCommand().GetChangePrice().Minor == nil || incoming.(*menu.Request).GetCommand().Metadata.ExpectedVersion == nil {
			t.Fatal("explicit zero lost presence")
		}
		if first != nil && (incoming.GetRequestId() == first.GetRequestId() || incoming.(*menu.Request).GetCommand().Metadata.CommandId != first.(*menu.Request).GetCommand().Metadata.CommandId) {
			t.Fatal("transport retry changed command identity")
		}
		first = incoming
	}
}
func TestPaginationIsTranslatedAndResourceBound(t *testing.T) {
	path := "/api/v1/menu/drinks"
	cursor := base64.RawURLEncoding.EncodeToString([]byte("1|/v1/menu/drinks|" + id))
	operation := contract.Document("api").Paths.Value(path).Get
	r := httptest.NewRequest("GET", path+"?limit=1&cursor="+cursor, nil)
	incoming, err := translate(r, "menu", path, operation)
	if err != nil {
		t.Fatal(err)
	}
	page := incoming.(*menu.Request).GetQuery().GetListDrinks().Page
	if page.Limit != 1 || page.GetAfter() != id {
		t.Fatal("page translation failed")
	}
	reply := &menu.Reply{ContractVersion: 1, RequestId: incoming.GetRequestId(), Context: "menu", Payload: &menu.Reply_Drinks{Drinks: &menu.Drinks{Paged: true, NextId: &idCopy}}}
	status, body, err := translateReply(incoming, reply, path)
	if err != nil || status != 200 || body.(map[string]any)["nextCursor"] != cursor {
		t.Fatal("reply cursor changed", err)
	}
	r = httptest.NewRequest("GET", path+"?limit=1&cursor="+base64.RawURLEncoding.EncodeToString([]byte("1|/v1/menu/editions|"+id)), nil)
	if _, err = translate(r, "menu", path, operation); err == nil {
		t.Fatal("foreign resource cursor accepted")
	}
}

var idCopy = id

func TestMismatchedAndMalformedRepliesBecomeUnavailable(t *testing.T) {
	c := config()
	c.Owners = []string{"menu"}
	c.Requests = caller(func(_ context.Context, r pb.Request) (pb.Reply, error) {
		return &menu.Reply{ContractVersion: 1, RequestId: r.GetRequestId(), Context: "menu", Payload: &menu.Reply_Drink{Drink: &menu.LoadedDrink{Exists: true}}}, nil
	})
	w := request(t, backendHandler(c), http.MethodGet, "/api/v1/menu/drinks/"+id, "", map[string]string{"Authorization": "Bearer cli"})
	if w.Code != 503 {
		t.Fatal("malformed owner reply accepted")
	}
}
