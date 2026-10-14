package support

import "fmt"

// validateSubscriptions rejects incomplete graphs before any worker starts.
func (s *StorefrontRuntime) validateSubscriptions() error {
	seen := map[string]bool{}
	for _, subscription := range s.subscriptions {
		binding := subscription.Binding
		expected, ok := expectedSubscriptions[binding.Consumer]
		if !ok || expected != binding || seen[binding.Consumer] || s.Databases[binding.Context] == nil {
			return fmt.Errorf("undeclared, duplicate or mismatched subscription: %s", binding.Consumer)
		}
		seen[binding.Consumer] = true
	}
	for consumer, binding := range expectedSubscriptions {
		if s.Databases[binding.Context] != nil && !seen[consumer] {
			return fmt.Errorf("missing subscription: %s", consumer)
		}
	}
	return nil
}
