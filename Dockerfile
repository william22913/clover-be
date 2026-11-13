# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/clover ./cmd/clover

FROM alpine:3.20
RUN addgroup -S clover && adduser -S -G clover clover
WORKDIR /app
COPY --from=build /out/clover /app/clover
COPY web /app/web
COPY media /app/media
USER clover
EXPOSE 8080
# The metrics listener on :9101 stays inside the container network.
HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=6 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/app/clover"]
