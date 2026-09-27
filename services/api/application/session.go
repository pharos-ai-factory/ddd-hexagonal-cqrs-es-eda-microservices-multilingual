package application

import "context"

// The demonstration operator is provisioned by the server, never selected by a browser.
const OperatorID = "00000000-0000-4000-8000-000000000001"

type Principal struct {
	Subject string `json:"subject"`
	Name    string `json:"name"`
	Role    string `json:"role"`
}

type Sessions interface {
	Create(context.Context) (string, error)
	Authenticate(context.Context, string) (Principal, bool, error)
	Revoke(context.Context, string) error
}

func AuthorisedChannels(principal Principal) []string {
	if principal.Subject != OperatorID || principal.Role != "operator" {
		return nil
	}
	return []string{"cafe:menu", "cafe:ordering", "cafe:preparation", "cafe:collection", "cafe:loyalty", "cafe:communication"}
}
