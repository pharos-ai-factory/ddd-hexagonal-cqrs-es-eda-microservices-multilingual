package realtime

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/application"
)

func TestProxyChoosesChannelsAndRejectsUntrustedCallers(t *testing.T) {
	handler := newHandler(config())
	if w := request(t, handler, "POST", "/api/realtime/connect", `{}`, nil); w.Code != 403 {
		t.Fatalf("untrusted proxy: %d", w.Code)
	}
	for _, cookie := range []string{"", "cafe_session=invalid"} {
		w := request(t, handler, "POST", "/api/realtime/connect", `{}`, map[string]string{
			"X-Cafe-Realtime-Proxy": "proxy", "Origin": "http://cafe.local", "Cookie": cookie})
		if !strings.Contains(w.Body.String(), `"disconnect"`) {
			t.Fatal("unauthenticated subscription accepted")
		}
	}
	w := request(t, handler, "POST", "/api/realtime/connect", `{"channels":["secret:foreign"]}`,
		map[string]string{"X-Cafe-Realtime-Proxy": "proxy", "Origin": "http://cafe.local", "Cookie": "cafe_session=valid"})
	var result struct {
		Result struct {
			User     string
			Subs     map[string]any
			ExpireAt int64 `json:"expire_at"`
		}
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Result.User != a.OperatorID || len(result.Result.Subs) != 6 {
		t.Fatal("missing server authority")
	}
	if result.Result.ExpireAt < time.Now().Unix() || result.Result.ExpireAt > time.Now().Add(a.ConnectLifetime).Unix() {
		t.Fatal("connect authority has no bounded validity")
	}
	if _, ok := result.Result.Subs["secret:foreign"]; ok {
		t.Fatal("browser-selected subscription accepted")
	}
}

func TestRefreshRevalidatesTheOriginalSessionAndOrigin(t *testing.T) {
	handler := newHandler(config())
	if w := request(t, handler, "POST", "/api/realtime/refresh", `{}`, nil); w.Code != 403 {
		t.Fatal("untrusted refresh accepted")
	}
	for _, scenario := range []struct {
		cookie, origin string
		valid          bool
	}{
		{"cafe_session=valid", "http://cafe.local", true},
		{"cafe_session=invalid", "http://cafe.local", false},
		{"", "http://cafe.local", false},
		{"cafe_session=valid", "http://attacker.local", false},
	} {
		w := request(t, handler, "POST", "/api/realtime/refresh", `{}`,
			map[string]string{"X-Cafe-Realtime-Proxy": "proxy", "Cookie": scenario.cookie, "Origin": scenario.origin})
		var reply struct {
			Result struct {
				Expired  bool  `json:"expired"`
				ExpireAt int64 `json:"expire_at"`
			}
		}
		if err := json.Unmarshal(w.Body.Bytes(), &reply); err != nil {
			t.Fatal(err)
		}
		if scenario.valid {
			if reply.Result.Expired || reply.Result.ExpireAt <= time.Now().Unix() {
				t.Fatal("valid session was not refreshed")
			}
		} else if !reply.Result.Expired || reply.Result.ExpireAt != 0 {
			t.Fatal("invalid authority was refreshed")
		}
	}
}
