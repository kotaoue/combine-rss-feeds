# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY *.go ./
COPY internal/ ./internal/

RUN go build -o combine-rss-feeds .

# Runtime stage
FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/combine-rss-feeds .

ENTRYPOINT ["/app/combine-rss-feeds"]
