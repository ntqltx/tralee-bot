FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/tralee-bot .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app && \
    mkdir -p /data && chown app:app /data

WORKDIR /app
COPY --from=builder /out/tralee-bot /app/tralee-bot

USER app
ENV SEEN_PATH=/data/seen.json

ENTRYPOINT ["/app/tralee-bot"]
