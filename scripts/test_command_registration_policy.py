"""Negative fixtures prevent per-command executor wiring returning to feature composition."""

import unittest

from command_registration_policy import violation


class CommandRegistrationTests(unittest.TestCase):
    def test_feature_composition_rejects_executor_and_low_level_registration(self):
        fixtures = {
            "services/operations/src/operations/apps/composition/preparation.py": [
                "from operations.adaptors.command_execution import CommandExecutor as Wrapper",
                'requests.command("startPreparation", parse, executor.execute)',
                "from operations.apps.runtime import command_subscription as subscribe",
            ],
            "services/engagement/src/apps/composition/loyalty.ts": [
                "import { bindCommand as wrap } from '../../adaptors/command-execution.js';",
                "credit: CommandExecutor<CreditCollectionCommand>;",
                "registry.command('redeemReward', parse, run);",
                "commandSubscription(event, codec, definitions, target, reaction, executor)",
            ],
            "services/storefront/apps/storefront/di/menu_requests.go": [
                "return wrapper.Bind(transaction, factory)",
                "return renamed.BindProjection(transaction, directory, key, factory)",
                "CreateDrink execution.Executor[commands.CreateDrinkCommand]",
                'rpc.Command(registry, "createDrink", parse, execute)',
            ],
        }
        for path, examples in fixtures.items():
            for content in examples:
                with self.subTest(path=path, content=content):
                    self.assertIsNotNone(violation(path, content))

    def test_direct_registrations_and_shared_execution_helpers_are_allowed(self):
        for path, content in [
            (
                "services/operations/src/operations/apps/composition/preparation.py",
                "commands.request(registry, name, parse, StartPreparationCommandHandler)",
            ),
            (
                "services/engagement/src/apps/composition/loyalty.ts",
                "commands.subscription(event, codec, definitions, target, reaction, factory)",
            ),
            (
                "services/storefront/apps/storefront/di/menu_requests.go",
                "registration.Register(commands, name, parse, factory)",
            ),
            (
                "services/operations/src/operations/apps/command_registration.py",
                "CommandExecutor(transaction, handler)",
            ),
            ("services/engagement/src/apps/command-registration.ts", "bindCommand(transaction, handler)"),
            (
                "services/storefront/apps/commandregistration/registration.go",
                "execution.Bind(transaction, factory)",
            ),
        ]:
            with self.subTest(path=path):
                self.assertIsNone(violation(path, content))
