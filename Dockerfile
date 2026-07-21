# Build the Orven binary
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /orven ./cmd/orven

# Runtime: python included so script-based plugins work out of the box
FROM python:3.13-alpine
RUN adduser -D -h /app orven
WORKDIR /app
COPY --from=build /orven /usr/local/bin/orven
COPY plugins/ /app/plugins/
USER orven
ENV ORVEN_DATA=/app/data ORVEN_PLUGINS=/app/plugins ORVEN_ADDR=:8420
VOLUME /app/data
EXPOSE 8420
CMD ["orven"]
