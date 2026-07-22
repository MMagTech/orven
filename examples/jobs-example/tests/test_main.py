"""Run with: python -m unittest discover -s tests (from the plugin folder)."""
import json
import os
import subprocess
import sys
import tempfile
import unittest

PLUGIN_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
FIXTURE = os.path.join(PLUGIN_DIR, "fixtures", "jobs.json")


def run_plugin(fixture, window_start=None):
    inp = {"contract_version": 1, "plugin_id": "jobs-example",
           "config": {}, "secrets": {}, "fixture": fixture}
    if window_start:
        inp["window_start"] = window_start
    out = subprocess.run(
        [sys.executable, "main.py"], cwd=PLUGIN_DIR,
        input=json.dumps(inp), capture_output=True, text=True, timeout=20,
    )
    assert out.returncode == 0, out.stderr
    return json.loads(out.stdout)


def temp_fixture(content):
    f = tempfile.NamedTemporaryFile("w", suffix=".json", delete=False)
    f.write(content)
    f.close()
    return f.name


class TestJobsExample(unittest.TestCase):
    def test_first_run_reports_everything(self):
        res = run_plugin(FIXTURE)
        self.assertEqual(res["status"], "ok")
        self.assertEqual(res["contract_version"], 1)
        titles = {o["title"]: o for o in res["observations"]}
        self.assertIn("nightly-backup completed", titles)
        self.assertIn("weekly-report completed", titles)
        self.assertIn("1 job is in a failed state", titles)
        self.assertIn("2 jobs running", titles)
        self.assertEqual(titles["nightly-backup completed"]["scope"], "event")
        self.assertIn("occurred_at", titles["nightly-backup completed"])
        self.assertEqual(titles["1 job is in a failed state"]["scope"], "state")
        self.assertEqual(titles["2 jobs running"]["scope"], "state")
        for title in titles:
            self.assertFalse(title.endswith("."))

    def test_events_filtered_by_window_states_kept(self):
        res = run_plugin(FIXTURE, window_start="2026-07-21T00:00:00+00:00")
        titles = [o["title"] for o in res["observations"]]
        self.assertIn("nightly-backup completed", titles)      # 03:12, inside window
        self.assertNotIn("weekly-report completed", titles)    # 22:40 previous day
        self.assertIn("1 job is in a failed state", titles)    # states always re-reported
        self.assertIn("2 jobs running", titles)

    def test_empty_queue_is_nothing(self):
        res = run_plugin(temp_fixture('{"jobs": []}'))
        self.assertEqual(res["status"], "nothing")
        self.assertNotIn("observations", res)

    def test_malformed_response_is_an_error_not_silence(self):
        res = run_plugin(temp_fixture("this is not json"))
        self.assertEqual(res["status"], "error")

    def test_unconfigured_without_fixture_is_an_error(self):
        res = run_plugin(None)
        self.assertEqual(res["status"], "error")
        self.assertIn("not configured", res["summary"])

    def test_no_recommendation_language(self):
        res = run_plugin(FIXTURE)
        text = json.dumps(res).lower()
        for banned in ("you should", "we recommend", "consider ", "to fix", "please "):
            self.assertNotIn(banned, text)


if __name__ == "__main__":
    unittest.main()
