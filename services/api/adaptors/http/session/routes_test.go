package session

import (
	"net/http"
	"testing"
)

func TestLoginRequiresOriginAndSetsPrivateCookie(t *testing.T) {
	handler := newHandler(config())
	if w := request(t, handler, "POST", "/auth/login", `{"password":"code"}`, nil); w.Code != 403 {
		t.Fatal("cross-origin login accepted")
	}
	w := request(t, handler, "POST", "/auth/login", `{"password":"code"}`, map[string]string{"Origin": "http://cafe.local"})
	cookies := w.Result().Cookies()
	if w.Code != 200 || len(cookies) != 1 || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatal("session cookie policy missing")
	}
}

func TestLoginRejectsTrailingJSONAndUnknownFields(t *testing.T) {
	handler := newHandler(config())
	for _, body := range []string{`{"password":"code"}{}`, `{"password":"code","extra":true}`} {
		if w := request(t, handler, "POST", "/auth/login", body, map[string]string{"Origin": "http://cafe.local"}); w.Code != 401 {
			t.Fatal("malformed login accepted")
		}
	}
}
