# Build stage: one static binary per command, no cgo.
FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/seed ./cmd/seed

# Verification stage: runs the full Go suite (database-backed cases included)
# against the compose network. Not part of the runtime image.
FROM build AS test
CMD ["go", "test", "-count=1", "./..."]

# Runtime stage: the API binary, the seed binary and the committed media.
FROM alpine:3.20

RUN apk add --no-cache ca-certificates \
 && adduser -D -H -u 10001 clover

WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /out/seed /app/seed
COPY media /app/media

USER clover
ENV MEDIA_DIR=/app/media \
    HTTP_ADDR=:8080 \
    METRICS_ADDR=:9090

EXPOSE 8080
ENTRYPOINT ["/app/server"]
