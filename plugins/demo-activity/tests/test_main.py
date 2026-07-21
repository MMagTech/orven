"""Run with: python -m unittest discover -s tests (from the plugin folder)."""
import json
import os
import subprocess
import sys
import unittest

PLUGIN_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
FIXTURE = os.path.join(PLUGIN_DIR, "fixtures", "sample.json")


def run_plugin(config, **extra_input):
    inp = {"contract_version": 1, "plugin_id": "demo-activity",
           "config": config, "fixture": FIXTURE, **extra_input}
    out = subprocess.run(
        [sys.executable, "main.py"], cwd=PLUGIN_DIR,
        input=json.dumps(inp), capture_output=True, text=True, timeout=20,
    )
    assert out.returncode == 0, out.stderr
    return json.loads(out.stdout)


class TestDemoActivity(unittest.TestCase):
    def test_activity_reports_ok(self):
        res = run_plugin({"scenario": "activity", "max_items": 5})
        self.assertEqual(res["status"], "ok")
        self.assertEqual(res["contract_version"], 1)
        # 5 items + 1 "more items" rollup
        self.assertEqual(len(res["observations"]), 6)
        self.assertIn("since your last briefing", res["summary"])

    def test_quiet_reports_nothing(self):
        res = run_plugin({"scenario": "quiet"})
        self.assertEqual(res["status"], "nothing")
        self.assertNotIn("observations", res)

    def test_outage_is_not_silent(self):
        res = run_plugin({"scenario": "outage"})
        self.assertEqual(res["status"], "unavailable")

    def test_auth_failure_reported(self):
        res = run_plugin({"scenario": "auth-problem"})
        self.assertEqual(res["status"], "auth_failed")

    def test_events_filtered_by_window_states_always_reported(self):
        # A window starting 60 minutes ago: only events that occurred
        # within the hour are new; states are reported regardless.
        from datetime import datetime, timedelta, timezone
        window = (datetime.now(timezone.utc) - timedelta(minutes=60)).isoformat()
        res = run_plugin({"scenario": "activity", "max_items": 10},
                         window_start=window)
        titles = [o["title"] for o in res["observations"]]
        self.assertIn("3 movies finished downloading", titles)   # 25 min ago
        self.assertIn("Certificate renewed", titles)             # 55 min ago
        self.assertNotIn("Overnight backup completed", titles)   # 340 min ago
        self.assertNotIn("Library growth", titles)               # 130 min ago
        for state_title in ("1 episode is stuck in the queue",
                            "2 new requests are awaiting approval",
                            "Container update available"):
            self.assertIn(state_title, titles)
        for o in res["observations"]:
            self.assertIn(o["scope"], ("event", "state"))

    def test_all_events_old_but_states_remain(self):
        # Nothing happened since the last run, but conditions persist:
        # the plugin must still report the states, not "nothing".
        from datetime import datetime, timezone
        res = run_plugin({"scenario": "activity", "max_items": 10},
                         window_start=datetime.now(timezone.utc).isoformat())
        self.assertEqual(res["status"], "ok")
        scopes = {o["scope"] for o in res["observations"]}
        self.assertEqual(scopes, {"state"})

    def test_no_recommendation_language(self):
        res = run_plugin({"scenario": "activity"})
        text = json.dumps(res).lower()
        for banned in ("you should", "we recommend", "try restarting", "please fix"):
            self.assertNotIn(banned, text)


if __name__ == "__main__":
    unittest.main()
