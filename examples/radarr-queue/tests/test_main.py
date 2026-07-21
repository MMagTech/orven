"""Run with: python -m unittest discover -s tests (from the plugin folder)."""
import json
import os
import subprocess
import sys
import tempfile
import unittest

PLUGIN_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
FIXTURE = os.path.join(PLUGIN_DIR, "fixtures", "queue.json")


def run_plugin(fixture, config=None, secrets=None):
    inp = {"contract_version": 1, "plugin_id": "radarr-queue",
           "config": config or {}, "secrets": secrets or {}, "fixture": fixture}
    out = subprocess.run(
        [sys.executable, "main.py"], cwd=PLUGIN_DIR,
        input=json.dumps(inp), capture_output=True, text=True, timeout=20,
    )
    assert out.returncode == 0, out.stderr
    return json.loads(out.stdout)


def temp_fixture(payload):
    f = tempfile.NamedTemporaryFile("w", suffix=".json", delete=False)
    json.dump(payload, f)
    f.close()
    return f.name


class TestRadarrQueue(unittest.TestCase):
    def test_stuck_and_active_reported_as_states(self):
        res = run_plugin(FIXTURE)
        self.assertEqual(res["status"], "ok")
        self.assertEqual(res["contract_version"], 1)
        titles = {o["title"]: o for o in res["observations"]}
        self.assertIn("1 download is stuck in the queue", titles)
        self.assertIn("2 downloads in progress", titles)
        for o in res["observations"]:
            self.assertEqual(o["scope"], "state")
            self.assertFalse(o["title"].endswith("."))

    def test_empty_queue_is_nothing(self):
        res = run_plugin(temp_fixture({"records": []}))
        self.assertEqual(res["status"], "nothing")
        self.assertNotIn("observations", res)

    def test_malformed_response_is_an_error_not_silence(self):
        path = temp_fixture({})
        with open(path, "w") as f:
            f.write("this is not json")
        res = run_plugin(path)
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
