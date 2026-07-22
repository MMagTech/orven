# Deploying Orven

The published image is `ghcr.io/mmagtech/orven` (amd64 and arm64).
All application state lives in `/app/data` — mount it somewhere
persistent and backups become a copy of that folder.

## docker run

```bash
docker run -d \
  --name orven \
  -p 8420:8420 \
  -v /path/to/orven-data:/app/data \
  --restart unless-stopped \
  ghcr.io/mmagtech/orven:latest
```

Open `http://<host>:8420`.

## docker compose

```yaml
services:
  orven:
    image: ghcr.io/mmagtech/orven:latest
    ports:
      - "8420:8420"
    volumes:
      - ./orven-data:/app/data
      # optional: manage plugins from the host
      # - ./orven-plugins:/app/plugins
    restart: unless-stopped
```

## Unraid

Add a container manually (or via a Docker template):

- **Repository:** `ghcr.io/mmagtech/orven:latest`
- **Port:** map `8420` to a host port of your choice
- **Path:** map `/app/data` → `/mnt/user/appdata/orven`

The container runs as its own unprivileged user (UID 1000). If the
appdata folder was created by root, give it to that UID once:

```bash
chown -R 1000:1000 /mnt/user/appdata/orven
```

Map `/app/plugins` → `/mnt/user/appdata/orven/plugins` so installed
plugins survive container updates. Plugins are installed from the
app itself (Plugins → Discover), or by dropping a plugin folder there
and restarting.

The Demo Activity plugin is seeded into a fresh installation exactly
once. Uninstalling it is permanent — it will not come back after a
restart or image update — and Settings offers "Restore the demo
plugin" if you ever want it again.

## Configuration

| env variable    | default        | meaning                       |
|-----------------|----------------|-------------------------------|
| `ORVEN_ADDR`    | `:8420`        | listen address                |
| `ORVEN_DATA`    | `/app/data`    | data directory (persist this) |
| `ORVEN_PLUGINS` | `/app/plugins` | installed plugins             |

## Health and updates

The image ships a `HEALTHCHECK` against `GET /healthz`, so `docker ps`
and Unraid show real container health. Releases are tagged
`vX.Y.Z`; the image tags `latest`, `X.Y`, and `X.Y.Z` track them.

## Backups

Settings → Backups manages everything: download a backup on demand,
schedule automatic daily backups with a retention count, browse and
delete existing ones, and restore (a safety backup of the current
state is written before any restore). Credentials are only included
when you opt in, and then always encrypted under a passphrase —
backups never contain plain secrets. Point the backup folder at a
mapped host path so your existing backup tooling picks the archives
up.

**After a restore** you have: every briefing, observation, and run
record from the backup; all settings (schedule, retention,
repositories, backup settings); and every plugin's configuration,
enabled state, and — when included — credentials. Plugin *files* are
not in backups: the app lists which catalog plugins to reinstall
(each resumes with its restored settings), manually added plugin
folders must be copied back by hand, and the backup passphrase must
be re-entered if automatic backups include credentials — it never
travels inside a backup. Restore means *put me back exactly where I
was*: anything created since the backup is removed, preserved only in
the automatic pre-restore safety backup.

**Security note:** Orven has no sign-in yet. Keep it on a trusted
network or behind an authenticating reverse proxy; do not expose it
directly to the internet.
