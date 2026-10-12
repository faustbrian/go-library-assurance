"""Focused bootstrap/selection checks; no Go commands or scenario execution."""
import copy
import hashlib
import importlib.util
import io
from pathlib import Path
import time
import unittest


ROOT = Path(__file__).resolve().parents[2]


def adapter():
    spec = importlib.util.spec_from_file_location(
        "assurance_selection", ROOT / "tools/compatibility/check-selection.py"
    )
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class FetchTests(unittest.TestCase):
    def test_digest_mismatch_refuses_helper_before_execution(self):
        module = adapter()
        payload = b"raise AssertionError('unverified code executed')"
        with self.assertRaisesRegex(ValueError, "digest mismatch"):
            module.fetch_verified(
                module.HELPER_URL, "0" * 64, 1024,
                open_url=lambda *_args, **_kwargs: io.BytesIO(payload),
            )

    def test_fetch_returns_only_verified_bounded_bytes(self):
        module = adapter()
        payload = b"immutable input\n"
        actual = module.fetch_verified(
            module.HELPER_URL, hashlib.sha256(payload).hexdigest(), len(payload),
            open_url=lambda *_args, **_kwargs: io.BytesIO(payload),
        )
        self.assertEqual(actual, payload)

    def test_oversize_is_fatal_even_with_a_matching_digest(self):
        module = adapter()
        payload = b"x" * 65
        with self.assertRaisesRegex(ValueError, "byte budget"):
            module.fetch_verified(
                module.HELPER_URL, hashlib.sha256(payload).hexdigest(), 64,
                open_url=lambda *_args, **_kwargs: io.BytesIO(payload),
            )

    def test_transport_failure_is_not_a_cached_or_latest_fallback(self):
        module = adapter()

        def unavailable(*_args, **_kwargs):
            raise OSError("unavailable")

        with self.assertRaisesRegex(OSError, "unavailable"):
            module.fetch_verified(module.HELPER_URL, "0" * 64, 64, open_url=unavailable)

    def test_duplicate_selection_keys_refuse(self):
        with self.assertRaisesRegex(ValueError, "duplicate JSON key"):
            adapter().decode_json(b'{"format":"first","format":"second"}')

    def test_expired_deadline_refuses_before_network_access(self):
        def unexpected(*_args, **_kwargs):
            self.fail("expired bootstrap deadline attempted network access")

        with self.assertRaisesRegex(TimeoutError, "time budget"):
            adapter().fetch_verified(
                adapter().HELPER_URL, "0" * 64, 64,
                open_url=unexpected, deadline=time.monotonic() - 1,
            )


class SelectionTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.module = adapter()
        # These are real immutable published inputs, verified before execution.
        # Fetch once; all refusal cases then use the same in-memory inputs.
        cls.inputs = cls.module.load_inputs()
        directory = ROOT / "scenarios/compatibility-consumer"
        cls.selection = cls.module.decode_json((directory / "selection.json").read_bytes())
        cls.go_mod = (directory / "go.mod").read_text()
        cls.imports = (directory / "consumer_test.go").read_text()

    def check(self, *, selection=None, go_mod=None, imports=None):
        return self.module.validate_selection(
            self.inputs,
            self.selection if selection is None else selection,
            self.go_mod if go_mod is None else go_mod,
            self.imports if imports is None else imports,
        )

    def test_actual_relocated_scenario_matches_the_published_generic_renderer(self):
        self.assertEqual(self.check(), 167)

    def test_old_tools_module_identity_refuses(self):
        changed = self.go_mod.replace(self.module.ASSURANCE_HEADER, self.module.TOOLS_HEADER, 1)
        with self.assertRaisesRegex(ValueError, "go.mod"):
            self.check(go_mod=changed)

    def test_changed_direct_dependency_refuses(self):
        changed = self.go_mod.replace("go-adaptive-throttle v1.0.1", "go-adaptive-throttle v1.0.2", 1)
        self.assertNotEqual(changed, self.go_mod)
        with self.assertRaisesRegex(ValueError, "go.mod"):
            self.check(go_mod=changed)

    def test_changed_import_inventory_refuses(self):
        with self.assertRaisesRegex(ValueError, "consumer_test.go"):
            self.check(imports=self.imports + "// stale import inventory\n")

    def test_unknown_replacement_family_refuses_in_actual_generic_policy(self):
        changed = copy.deepcopy(self.selection)
        changed["replacements"][0]["module_path"] = "github.com/faustbrian/go-unrelated/v2"
        with self.assertRaisesRegex(ValueError, "module family"):
            self.check(selection=changed)

    def test_duplicate_replacement_refuses_in_actual_generic_policy(self):
        changed = copy.deepcopy(self.selection)
        changed["replacements"].append(copy.deepcopy(changed["replacements"][0]))
        with self.assertRaisesRegex(ValueError, "duplicated"):
            self.check(selection=changed)


if __name__ == "__main__":
    unittest.main()
