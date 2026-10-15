"""Exercise actual Required aggregation, including failed and missing work."""
import os
from pathlib import Path
import subprocess
import unittest


SCRIPT = Path(__file__).with_name("check-required.sh")


class RequiredTests(unittest.TestCase):
    def check(self, updates, success):
        environment = dict(os.environ, EVENT_NAME="push", QUALITY_RESULT="success",
                           SECRETS_RESULT="success", SOURCE_DOCS_RESULT="success",
                           SCENARIOS_REQUIRED="false",
                           SCENARIOS_RESULT="skipped", DEPENDENCY_REVIEW_RESULT="skipped",
                           MINIMUM_COMPILER_REQUIRED="false", MINIMUM_COMPILER_RESULT="skipped")
        for key, value in updates.items():
            if value is None:
                environment.pop(key, None)
            else:
                environment[key] = value
        result = subprocess.run(["bash", str(SCRIPT)], env=environment,
                                capture_output=True, timeout=5, check=False)
        self.assertEqual(result.returncode == 0, success, result.stderr.decode())

    def test_inapplicable_scenarios_are_honestly_skipped(self):
        self.check({}, True)

    def test_selected_scenarios_must_pass(self):
        self.check({"SCENARIOS_REQUIRED": "true", "SCENARIOS_RESULT": "success"}, True)
        for result in ("failure", "cancelled", "skipped", "", "unknown"):
            with self.subTest(result=result):
                self.check({"SCENARIOS_REQUIRED": "true", "SCENARIOS_RESULT": result}, False)

    def test_failed_or_missing_cheap_checks_cannot_pass(self):
        for key in ("QUALITY_RESULT", "SECRETS_RESULT"):
            for result in ("failure", "cancelled", "skipped", ""):
                with self.subTest(key=key, result=result):
                    self.check({key: result}, False)

    def test_source_documentation_is_always_mandatory(self):
        self.check({"SOURCE_DOCS_RESULT": None}, False)
        self.check({"SOURCE_DOCS_RESULT": "success"}, True)
        for result in ("failure", "cancelled", "skipped", "", "unknown"):
            with self.subTest(result=result):
                self.check({"SOURCE_DOCS_RESULT": result}, False)

    def test_invalid_classification_cannot_hide_scenarios(self):
        for selected in ("", "unknown", "TRUE"):
            self.check({"SCENARIOS_REQUIRED": selected}, False)
        self.check({"SCENARIOS_RESULT": "success"}, False)

    def test_pr_dependency_review_is_mandatory(self):
        self.check({"EVENT_NAME": "pull_request", "DEPENDENCY_REVIEW_RESULT": "success"}, True)
        for result in ("skipped", "cancelled", "failure", ""):
            self.check({"EVENT_NAME": "pull_request", "DEPENDENCY_REVIEW_RESULT": result}, False)

    def test_manual_execution_is_supported_and_unknown_events_refuse(self):
        self.check({"EVENT_NAME": "workflow_dispatch"}, True)
        self.check({"EVENT_NAME": "unknown"}, False)

    def test_requested_minimum_compiler_matrix_must_pass(self):
        selected = {"EVENT_NAME": "workflow_dispatch", "MINIMUM_COMPILER_REQUIRED": "true"}
        self.check(dict(selected, MINIMUM_COMPILER_RESULT="success"), True)
        for result in ("failure", "cancelled", "skipped", "", "unknown", None):
            with self.subTest(result=result):
                self.check(dict(selected, MINIMUM_COMPILER_RESULT=result), False)

    def test_unrequested_minimum_compiler_is_honestly_skipped(self):
        for event in ("push", "pull_request", "workflow_dispatch"):
            updates = {"EVENT_NAME": event}
            if event == "pull_request":
                updates["DEPENDENCY_REVIEW_RESULT"] = "success"
            self.check(updates, True)
            for result in ("success", "failure", "cancelled", "", "unknown", None):
                with self.subTest(event=event, result=result):
                    self.check(dict(updates, MINIMUM_COMPILER_RESULT=result), False)

    def test_minimum_compiler_classification_fails_closed(self):
        for selected in ("", "unknown", "TRUE", None):
            with self.subTest(selected=selected):
                self.check({"MINIMUM_COMPILER_REQUIRED": selected}, False)
        for event in ("push", "pull_request"):
            updates = {"EVENT_NAME": event, "MINIMUM_COMPILER_REQUIRED": "true",
                       "MINIMUM_COMPILER_RESULT": "success"}
            if event == "pull_request":
                updates["DEPENDENCY_REVIEW_RESULT"] = "success"
            self.check(updates, False)


if __name__ == "__main__":
    unittest.main()
