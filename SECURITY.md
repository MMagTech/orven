# Security Policy

Orven handles credentials for the systems it observes, so security
reports get priority attention.

## Reporting a vulnerability

Please report vulnerabilities **privately** via GitHub: *Security →
Report a vulnerability* on this repository (GitHub private
vulnerability reporting). Do not open a public issue for anything you
believe is exploitable.

You can expect an acknowledgment within a few days. There is no bug
bounty — this is a community project — but reports are credited in
release notes unless you prefer otherwise.

## Supported versions

| Version | Supported |
|---------|-----------|
| 0.1.x   | ✅        |

Pre-1.0, only the latest release line receives fixes.

## What counts

Reports we especially want:

- Escapes of the **credential publication boundary** — any way a
  plugin's assigned secret can reach a briefing, log, run history,
  export, or page in full (see `docs/CONSTRAINTS.md` §14).
- A plugin obtaining **another plugin's credentials through the
  application** (the input channel, not the shared filesystem — see
  below).
- Injection into rendered briefings (HTML/script through observation
  content).
- Authentication or authorization flaws in whatever auth lands in
  future versions.

## Known, documented limitations

Some boundaries are documented as not yet enforced — see **"Known
enforcement gaps"** in
[docs/CONSTRAINTS.md](docs/CONSTRAINTS.md#known-enforcement-gaps-recorded-planned):
plugin network egress is not restricted, and plugins currently share
the application's OS user (same-filesystem access). These are planned
work, and reports that only restate them will be closed as known —
but reports that show them to be *worse than documented* are very
welcome.

## Deployment guidance

Orven currently ships without authentication. Run it on a trusted
network or behind an authenticating reverse proxy, and treat plugin
installation as the trust decision it is. See
[docs/DEPLOY.md](docs/DEPLOY.md).
