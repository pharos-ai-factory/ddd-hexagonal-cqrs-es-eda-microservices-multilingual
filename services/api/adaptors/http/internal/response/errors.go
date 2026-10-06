package response

import "net/http"

// ErrorResponse is the technical error body declared by the OpenAPI Error schema.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

// Error pairs a stable public code with its HTTP status.
type Error struct {
	status int
	code   string
}

var (
	Unauthorised           = Error{http.StatusUnauthorized, "unauthorised"}
	OriginDenied           = Error{http.StatusForbidden, "origin_denied"}
	InvalidAccessCode      = Error{http.StatusUnauthorized, "invalid_access_code"}
	SessionUnavailable     = Error{http.StatusServiceUnavailable, "session_unavailable"}
	DiagnosticsUnavailable = Error{http.StatusServiceUnavailable, "diagnostics_unavailable"}
	ProxyDenied            = Error{http.StatusForbidden, "proxy_denied"}
	TemporarilyUnavailable = Error{http.StatusServiceUnavailable, "temporarily_unavailable"}
)

func (e Error) Write(w http.ResponseWriter) {
	JSON(w, e.status, ErrorResponse{Code: e.code})
}
