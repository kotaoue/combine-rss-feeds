# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY *.go ./

RUN go build -o combine-rss-feeds .

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/combine-rss-feeds .

ENTRYPOINT ["/app/combine-rss-feeds"]
