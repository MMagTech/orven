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

To install plugins beyond the bundled demo, also map
`/app/plugins` → `/mnt/user/appdata/orven-plugins` and drop plugin
folders there (one folder per plugin; restart the container after
adding one).

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

**Security note:** Orven has no sign-in yet. Keep it on a trusted
network or behind an authenticating reverse proxy; do not expose it
directly to the internet. The `data/secrets/` subfolder holds plugin
credentials — encrypt any backup that includes it.
