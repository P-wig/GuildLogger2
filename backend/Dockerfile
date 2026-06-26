FROM golang:1.26.2-alpine3.22 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/main .

FROM alpine:3.22

WORKDIR /app
RUN adduser -D -g '' appuser
USER appuser

ENV PORT=5001
EXPOSE 5001

COPY --from=builder /out/main /app/main
CMD ["/app/main"]
