package support

import "testing"

func TestNonDevelopmentCompositionIsRejectedBeforeConnecting(t *testing.T) {
	for _, environment := range []string{"", "staging", "production"} {
		t.Run(environment, func(t *testing.T) {
			t.Setenv("APP_ENV", environment)
			service, err := Open("menu")
			if err == nil || service != nil {
				t.Fatal("reference accepted a non-development environment")
			}
		})
	}
}
