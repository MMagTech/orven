# Build the Orven binary
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /orven ./cmd/orven

# Runtime: python included so script-based plugins work out of the box
FROM python:3.13-alpine

LABEL org.opencontainers.image.title="Orven" \
      org.opencontainers.image.description="A read-only daily briefing for your self-hosted world. Orven observes; it never changes anything." \
      org.opencontainers.image.source="https://github.com/MMagTech/orven" \
      org.opencontainers.image.licenses="Apache-2.0"

RUN adduser -D -h /app orven
WORKDIR /app
COPY --from=build /orven /usr/local/bin/orven
COPY plugins/ /app/plugins/
RUN mkdir -p /app/data && chown -R orven:orven /app
USER orven
ENV ORVEN_DATA=/app/data ORVEN_PLUGINS=/app/plugins ORVEN_ADDR=:8420
VOLUME /app/data
EXPOSE 8420
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
  CMD wget -qO- http://127.0.0.1:8420/healthz || exit 1
CMD ["orven"]
