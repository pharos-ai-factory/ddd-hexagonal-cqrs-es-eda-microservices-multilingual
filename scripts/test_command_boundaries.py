import unittest
from command_boundaries import violations, aggregate_methods, check


class CommandBoundaryTests(unittest.TestCase):
    def test_real_python_application(self):
        self.assertEqual(check(), [])

    def test_mutations_and_aliases_are_rejected(self):
        for action in ('ticket.start()', 'action = ticket.start; action()', 'ticket.complete()'):
            source = f'''from operations.contexts.preparation.domain import PreparationTicket as Ticket
class PlacedIntegrationEventHandler:
    def handle(self, ticket: Ticket):
        {action}
'''
            self.assertTrue(violations(source, aggregate_methods()))

    def test_creating_an_aggregate_in_an_event_handler_is_rejected(self):
        self.assertTrue(violations('from operations.contexts.preparation.domain import PreparationTicket\nclass Reaction:\n def handle(self):\n  PreparationTicket(None)', aggregate_methods()))

    def test_store_capability_and_misplaced_handler_method_are_rejected(self):
        self.assertTrue(violations('from operations.contexts.preparation.domain import PreparationTicket\nclass Reaction:\n def __init__(self, store: AggregateCommandPort): self.store = store', aggregate_methods()))
        self.assertTrue(violations('from operations.contexts.preparation.domain import PreparationTicket\nclass StartCommandHandler:\n def handle(self, ticket: PreparationTicket): ticket.start()', aggregate_methods()))

    def test_command_execution_and_rehydration_are_allowed(self):
        self.assertEqual(violations('from operations.contexts.preparation.domain import PreparationTicket\nclass StartCommandHandler:\n def execute(self, ticket: PreparationTicket): ticket.start()', aggregate_methods()), [])
        self.assertEqual(violations('from operations.contexts.preparation.domain import PreparationTicket\nclass QueryHandler:\n def handle(self, ticket: PreparationTicket): return ticket.snapshot()', aggregate_methods()), [])

    def test_free_functions_and_qualified_constructors_cannot_bypass_commands(self):
        self.assertTrue(violations('import operations.contexts.preparation.domain as d\ndef change(): return d.PreparationTicket(None)', aggregate_methods()))
        self.assertTrue(violations('def change(port: AggregateCommandPort): port.execute(None, None)', aggregate_methods()))

    def test_write_port_alias_in_wrong_method_is_rejected(self):
        self.assertTrue(violations('class StartCommandHandler:\n def __init__(self, port: AggregateCommandPort): self.port = port\n def handle(self): self.port.execute(None, None)', aggregate_methods()))

    def test_private_state_is_owned_by_the_aggregate_even_inside_a_command_handler(self):
        for owner in ('EventHandler', 'StartCommandHandler'):
            self.assertTrue(violations(f'from operations.contexts.preparation.domain import PreparationTicket\nclass {owner}:\n def execute(self, ticket: PreparationTicket): ticket._state = None', aggregate_methods()))
