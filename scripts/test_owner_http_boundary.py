import unittest
from owner_http_boundary import violation


class OwnerHTTPBoundaryTests(unittest.TestCase):
    def test_context_http_adaptors_fail_in_every_language(self):
        for path in ('services/storefront/contexts/menu/adaptors/http/routes.go',
                     'services/operations/src/operations/contexts/preparation/adaptors/http/inputs.py',
                     'services/engagement/src/contexts/loyalty/adaptors/http/inputs.ts'):
            self.assertIsNotNone(violation(path, ''))

    def test_business_routes_cannot_move_to_composition_or_shared_http(self):
        for path, content in (
            ('services/storefront/apps/support/service.go', 'mux.HandleFunc("GET /v1/ordering/orders", handler)'),
            ('services/operations/src/operations/apps/service.py', 'app.get("/v1/preparation/tickets")(handler)'),
            ('services/engagement/src/adaptors/http.ts', "if (path === '/v1/loyalty/rewards') handle();"),
        ):
            # Go patterns include their method before the route.
            self.assertIsNotNone(violation(path, content))

    def test_operational_and_api_routes_and_outbound_provider_calls_are_allowed(self):
        for path, content in (
            ('services/storefront/apps/support/service.go', 'mux.HandleFunc("GET /healthz", handler)'),
            ('services/operations/src/operations/apps/service.py', 'app.get("/diagnostics")(handler)'),
            ('services/api/adaptors/http/backend/routes.go', 'mux.HandleFunc("GET /api/v1/menu/drinks", handler)'),
            ('services/storefront/integrations/delivery/client.go', 'url + "/v1/deliver"'),
        ):
            self.assertIsNone(violation(path, content))
