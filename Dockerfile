# syntax=docker/dockerfile:1.7

FROM golang:1.26.1-alpine AS builder
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/godis ./cmd/server

FROM alpine:3.20
RUN addgroup -S godis && adduser -S -G godis godis

WORKDIR /data
COPY --from=builder /out/godis /usr/local/bin/godis

RUN chown -R godis:godis /data
USER godis

EXPOSE 6379
VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/godis"]
