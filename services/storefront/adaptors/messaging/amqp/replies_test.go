package amqp

import "testing"

func TestReplyAddressesStayWithinTheirOwner(t *testing.T) {
	instance := "11111111-1111-4111-8111-111111111111"
	for _, route := range []string{"", "reply.project", "reply.project." + instance} {
		actual, err := validateReplyRoute("project", route)
		if err != nil || actual == "" || (route != "" && actual != route) {
			t.Fatalf("valid reply route %q: %q, %v", route, actual, err)
		}
	}
	for _, route := range []string{"reply.foreign." + instance, "reply.project.#", "reply.project.*",
		"reply.project." + instance + ".extra", "reply.project.invalid", "amq.rabbitmq.reply-to", "ref.project.queries"} {
		if _, err := validateReplyRoute("project", route); err == nil {
			t.Errorf("accepted unscoped reply route %q", route)
		}
	}
}

func TestLegacyReplyAddressPreservesStoredIdentity(t *testing.T) {
	for _, route := range []string{"", "reply.project"} {
		if actual := ReplyRoute(withReplyRoute(t.Context(), "project", route)); actual != "" {
			t.Fatalf("legacy reply address changed persisted identity: %q", actual)
		}
	}
	route := "reply.project.11111111-1111-4111-8111-111111111111"
	if actual := ReplyRoute(withReplyRoute(t.Context(), "project", route)); actual != route {
		t.Fatalf("instance address changed: %q", actual)
	}
}
