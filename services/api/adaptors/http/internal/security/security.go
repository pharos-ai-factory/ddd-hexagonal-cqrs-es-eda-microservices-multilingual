package security

import (
	"crypto/subtle"
	"net/http"
)

const CookieName = "cafe_session"

func Equal(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }

func Token(r *http.Request) string {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func Origin(r *http.Request, origins []string) bool {
	for _, allowed := range origins {
		if r.Header.Get("Origin") == allowed {
			return true
		}
	}
	return false
}
