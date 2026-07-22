"""Jobs Example plugin for Orven — the reference HTTP-source plugin.

It observes a made-up jobs API (GET {url}/api/jobs) so the pattern
stays generic: any service with an HTTP endpoint is observed the same
way. The pattern to copy:

  - stdlib HTTP only, GET requests only, hard timeout;
  - the secret arrives in the engine input and goes into a header,
    never a URL;
  - transport failures map honestly onto contract statuses;
  - when the engine passes a fixture path, read it instead of the
    network, so tests and `orven validate` need no real service;
  - completed work is reported as events filtered by window_start
    (only what's new since the last successful run); ongoing
    conditions are states, re-reported until they clear.

Facts only: this plugin states what the jobs API reports, never what
to do about it.
"""
import json
import sys
import urllib.error
import urllib.request
from datetime import datetime

CONTRACT_VERSION = 1


def parse_ts(value):
    """Parse an ISO timestamp; Go's zero time and junk mean 'unknown'."""
    if not value:
        return None
    try:
        ts = datetime.fromisoformat(value.replace("Z", "+00:00"))
    except ValueError:
        return None
    if ts.year <= 1:
        return None
    return ts


def result(status, summary, observations=None):
    out = {"contract_version": CONTRACT_VERSION, "status": status, "summary": summary}
    if observations:
        out["observations"] = observations
    return out


def fetch_jobs(url, token, timeout=10):
    req = urllib.request.Request(
        url + "/api/jobs",
        headers={"Authorization": "Bearer " + token, "Accept": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.load(resp)


def build(payload, window_start):
    jobs = payload.get("jobs", [])
    if not jobs:
        return result("nothing", "The jobs service reported an empty queue.")

    observations = []

    for job in (j for j in jobs if j.get("state") == "completed"):
        finished = parse_ts(job.get("completed_at"))
        # Events are new-since-last-run: skip what a previous run
        # already reported.
        if window_start and finished and finished <= window_start:
            continue
        obs = {
            "title": f"{job.get('name', 'unnamed job')} completed",
            "kind": "fact",
            "scope": "event",
        }
        if finished:
            obs["occurred_at"] = finished.isoformat()
        observations.append(obs)

    failed = [j for j in jobs if j.get("state") == "failed"]
    if failed:
        details = "; ".join(
            f"{j.get('name', 'unnamed')}: {j.get('error', 'no error message')}" for j in failed[:3]
        )
        observations.append({
            "title": f"{len(failed)} job{'s are' if len(failed) != 1 else ' is'} in a failed state",
            "body": details + ".",
            "kind": "notice",
            "scope": "state",  # still true right now; re-reported until it clears
        })

    running = [j for j in jobs if j.get("state") == "running"]
    if running:
        observations.append({
            "title": f"{len(running)} job{'s' if len(running) != 1 else ''} running",
            "body": ", ".join(j.get("name", "unnamed") for j in running[:5]) + ".",
            "kind": "count",
            "scope": "state",
        })

    if not observations:
        return result("nothing", "No job activity since the last briefing.")
    return result("ok", f"{len(jobs)} jobs checked.", observations)


def main():
    inp = json.load(sys.stdin)
    window_start = parse_ts(inp.get("window_start"))
    try:
        if inp.get("fixture"):
            with open(inp["fixture"], encoding="utf-8") as f:
                payload = json.load(f)
        else:
            cfg = inp.get("config", {})
            url = (cfg.get("url") or "").rstrip("/")
            token = (inp.get("secrets") or {}).get("api_token", "")
            if not url or not token:
                json.dump(result("error", "The service URL and API token are not configured."), sys.stdout)
                return
            payload = fetch_jobs(url, token)
    except urllib.error.HTTPError as e:
        if e.code in (401, 403):
            json.dump(result("auth_failed", "The service rejected the configured API token."), sys.stdout)
        else:
            json.dump(result("error", f"The service answered with HTTP {e.code}."), sys.stdout)
        return
    except (urllib.error.URLError, OSError, TimeoutError):
        json.dump(result("unavailable", "The service could not be reached."), sys.stdout)
        return
    except (ValueError, KeyError):
        json.dump(result("error", "The service returned a response that could not be understood."), sys.stdout)
        return
    json.dump(build(payload, window_start), sys.stdout)


if __name__ == "__main__":
    main()
