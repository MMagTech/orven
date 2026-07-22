"""Demo Activity plugin for Orven.

Reads engine input from stdin, reports pretend activity from fixture
data, writes one JSON result to stdout. States facts only — never
suggestions or fixes.

This plugin is the reference implementation of the two observation
scopes: events are reported only when they occurred after
``window_start`` (the last successful run), while states are reported
on every run for as long as the condition remains true.
"""
import json
import os
import sys
from datetime import datetime, timedelta, timezone

CONTRACT_VERSION = 1


def parse_ts(value):
    """Parse an engine timestamp; Go's zero time means 'never'."""
    if not value:
        return None
    try:
        ts = datetime.fromisoformat(value.replace("Z", "+00:00"))
    except ValueError:
        return None
    if ts.year <= 1:
        return None
    return ts


def load_fixture(path):
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def build_output(inp):
    cfg = inp.get("config", {})
    scenario = cfg.get("scenario", "activity")
    source = cfg.get("source_name", "Demo media server")
    try:
        max_items = int(cfg.get("max_items", 5))
    except (TypeError, ValueError):
        max_items = 5

    if scenario == "outage":
        return {
            "contract_version": CONTRACT_VERSION,
            "status": "unavailable",
            "summary": f"{source} could not be reached.",
        }
    if scenario == "auth-problem":
        return {
            "contract_version": CONTRACT_VERSION,
            "status": "auth_failed",
            "summary": f"{source} rejected the configured credentials.",
        }
    if scenario == "quiet":
        return {
            "contract_version": CONTRACT_VERSION,
            "status": "nothing",
            "summary": f"No new activity on {source}.",
        }

    now = parse_ts(inp.get("now")) or datetime.now(timezone.utc)
    window_start = parse_ts(inp.get("window_start"))

    fixture = inp.get("fixture") or os.path.join(
        os.path.dirname(os.path.abspath(__file__)), "fixtures", "sample.json"
    )
    entries = load_fixture(fixture)["events"]

    reportable = []
    for e in entries:
        scope = e.get("scope", "event")
        occurred = None
        if "at" in e:
            # Fixture events are anchored to a clock time (today, or
            # yesterday if that time hasn't arrived yet) so that window
            # filtering behaves like a real source: each event is new
            # exactly once, then never re-reported.
            hh, mm = (int(x) for x in e["at"].split(":"))
            occurred = now.replace(hour=hh, minute=mm, second=0, microsecond=0)
            if occurred > now:
                occurred -= timedelta(days=1)
        # Events are new-since-last-run; skip ones the previous run
        # already reported. States are re-reported while still true.
        if scope == "event" and window_start and occurred and occurred <= window_start:
            continue
        obs = {
            "title": e["title"],
            "body": e["detail"],
            "kind": e.get("kind", "fact"),
            "scope": scope,
        }
        if occurred is not None:
            obs["occurred_at"] = occurred.isoformat()
        reportable.append(obs)

    if not reportable:
        return {
            "contract_version": CONTRACT_VERSION,
            "status": "nothing",
            "summary": f"No new activity on {source}.",
        }

    observations = reportable[:max_items]
    extra = len(reportable) - max_items
    if extra > 0:
        observations.append(
            {"title": f"{extra} more items", "body": "Not shown to keep the briefing short.", "kind": "count"}
        )

    return {
        "contract_version": CONTRACT_VERSION,
        "status": "ok",
        "summary": f"{len(reportable)} new items on {source} since your last briefing.",
        "observations": observations,
    }


def main():
    inp = json.load(sys.stdin)
    json.dump(build_output(inp), sys.stdout)


if __name__ == "__main__":
    main()
