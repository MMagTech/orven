"""Radarr Queue plugin for Orven — the reference HTTP-source plugin.

The pattern to copy:
  - stdlib HTTP only (urllib), GET requests only, hard timeout;
  - secrets arrive in the engine input and go into a header, never a URL;
  - transport failures map honestly onto contract statuses;
  - when the engine passes a fixture path, read it instead of the
    network, so tests and `orven validate` never need a real Radarr.

Facts only: this plugin states what the queue contains, never what to
do about it.
"""
import json
import sys
import urllib.error
import urllib.request

CONTRACT_VERSION = 1


def result(status, summary, observations=None):
    out = {"contract_version": CONTRACT_VERSION, "status": status, "summary": summary}
    if observations:
        out["observations"] = observations
    return out


def fetch_queue(url, api_key, timeout=10):
    req = urllib.request.Request(
        url + "/api/v3/queue?pageSize=100",
        headers={"X-Api-Key": api_key, "Accept": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.load(resp)


def build(queue):
    records = queue.get("records", [])
    if not records:
        return result("nothing", "The Radarr download queue is empty.")

    stuck = [r for r in records
             if r.get("errorMessage") or r.get("trackedDownloadState") == "importFailed"]
    observations = []
    if stuck:
        names = ", ".join(r.get("title", "unknown") for r in stuck[:3])
        observations.append({
            "title": f"{len(stuck)} download{'s are' if len(stuck) != 1 else ' is'} stuck in the queue",
            "body": f"Waiting on: {names}.",
            "kind": "notice",
            "scope": "state",  # still true right now; re-reported until it clears
        })
    active = len(records) - len(stuck)
    if active:
        observations.append({
            "title": f"{active} download{'s' if active != 1 else ''} in progress",
            "body": "Radarr reports these as still transferring.",
            "kind": "count",
            "scope": "state",
        })
    return result("ok", f"{len(records)} items in the Radarr queue.", observations)


def main():
    inp = json.load(sys.stdin)
    try:
        if inp.get("fixture"):
            with open(inp["fixture"], encoding="utf-8") as f:
                queue = json.load(f)
        else:
            cfg = inp.get("config", {})
            url = (cfg.get("url") or "").rstrip("/")
            api_key = (inp.get("secrets") or {}).get("api_key", "")
            if not url or not api_key:
                json.dump(result("error", "The Radarr URL and API key are not configured."), sys.stdout)
                return
            queue = fetch_queue(url, api_key)
    except urllib.error.HTTPError as e:
        if e.code in (401, 403):
            json.dump(result("auth_failed", "Radarr rejected the configured API key."), sys.stdout)
        else:
            json.dump(result("error", f"Radarr answered with HTTP {e.code}."), sys.stdout)
        return
    except (urllib.error.URLError, OSError, TimeoutError):
        json.dump(result("unavailable", "Radarr could not be reached."), sys.stdout)
        return
    except (ValueError, KeyError):
        json.dump(result("error", "Radarr returned a response that could not be understood."), sys.stdout)
        return
    json.dump(build(queue), sys.stdout)


if __name__ == "__main__":
    main()
