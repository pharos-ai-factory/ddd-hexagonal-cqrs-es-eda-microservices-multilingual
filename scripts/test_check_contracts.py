"""The regeneration gate rejects untracked additions and removed or changed outputs."""
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import check_contracts


class ContractDriftTests(unittest.TestCase):
    def test_every_kind_of_generated_drift_fails(self):
        for kind, action in (
            ('added', "Path('services/generated/new.ts').write_text('generated')"),
            ('removed', "Path('services/generated/existing.ts').unlink()"),
            ('changed', "Path('services/generated/existing.ts').write_text('changed')"),
        ):
            with self.subTest(kind=kind), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                for name in ('contracts', 'scripts', 'devops', 'node_modules', 'services/generated'):
                    (root/name).mkdir(parents=True)
                (root/'services/generated/existing.ts').write_text('original')
                (root/'scripts/generate.py').write_text('from pathlib import Path\n'+action+'\n')
                with patch.object(check_contracts, 'ROOT', root):
                    with self.assertRaisesRegex(SystemExit, 'services/generated/'):
                        check_contracts.check()
                self.assertEqual((root/'services/generated/existing.ts').read_text(), 'original')
                self.assertFalse((root/'services/generated/new.ts').exists())


if __name__ == '__main__':
    unittest.main()
