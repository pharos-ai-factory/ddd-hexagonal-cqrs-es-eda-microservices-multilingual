import unittest
from readiness import report, expected_queues


class ReadinessTests(unittest.TestCase):
    def test_health_without_consumers_is_unready(self):
        state = report({'OPERATIONS_API_KEY': 'unused', 'OPERATIONS_PORT': '1'}, ('OPERATIONS',), (),
                       fetch=lambda *args: {}, queues_fetch=lambda *args: [])
        self.assertFalse(state['ready'])
        self.assertEqual(state['checks']['operations'], 'ready')

    def test_registered_consumers_and_database_are_ready(self):
        values = {'OPERATIONS_API_KEY': 'unused', 'OPERATIONS_PORT': '1'}
        queues = [{'name': name, 'consumers': 1} for name in expected_queues(('OPERATIONS',))]
        self.assertTrue(report(values, ('OPERATIONS',), (), lambda *a: {}, lambda *a: queues)['ready'])
        def unavailable(*args):
            raise ValueError('private connection details')
        result = report(values, ('OPERATIONS',), (), unavailable, lambda *a: queues)
        self.assertFalse(result['ready'])
        self.assertNotIn('private connection', str(result))

    def test_paused_subscription_skips_both_consumers_only(self):
        queues = expected_queues(('ENGAGEMENT',), ('loyalty.issue-reward',))
        self.assertNotIn('ref.loyalty.issue-reward.command', queues)
        self.assertNotIn('ref.loyalty.issue-reward', queues)
        self.assertIn('ref.loyalty.commands', queues)

    def test_browser_readiness_requires_the_realtime_endpoint(self):
        values = {'WEB_PORT': '1', 'REALTIME_ADMIN_PORT': '2'}
        def fetch(url, headers=None):
            if url.endswith('/health'):
                raise OSError('offline')
            return {}
        state = report(values, ('WEB',), (), fetch, lambda *a: [])
        self.assertFalse(state['ready'])
        self.assertEqual(state['checks']['realtime'], 'realtime endpoint unavailable')
