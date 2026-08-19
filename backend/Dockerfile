FROM golang:1.26.2-alpine3.22 AS builder

WORKDIR /src

COPY go.mod go.sum ./
# Cache module downloads separately — survives --no-cache on the source layers.
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
# Cache the Go build output so incremental rebuilds are fast.
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/main .

FROM alpine:3.22

WORKDIR /app
RUN adduser -D -g '' appuser
USER appuser

ENV PORT=5001
EXPOSE 5001

COPY --from=builder /out/main /app/main
CMD ["/app/main"]
