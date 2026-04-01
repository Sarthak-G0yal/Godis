# Godis

A Redis-inspired key-value server built in Go for learning systems design, protocol handling, and persistence.

## Why This Project

Godis is designed to be simple enough to understand end-to-end, while still including practical backend concerns:
- concurrent TCP clients
- command parsing and response protocol
- thread-safe in-memory storage
- append-only persistence with startup replay
- graceful shutdown behavior

## Current Status

- Phase 1: complete
- Phase 2: complete
- Phase 3: complete
- Phase 4: complete (AOF persistence)
- Phase 5: planned (integration tests + hardening)

Detailed technical notes: [ARCHITECTURE.md](ARCHITECTURE.md)
Future roadmap and Phase 5 plan: [FutureScope.md](FutureScope.md)

## Features

- Commands: `PING`, `SET`, `GET`, `DEL`, `EXISTS`
- Thread-safe in-memory store using `sync.RWMutex`
- AOF append for successful write commands (`SET`, `DEL`)
- AOF replay before accepting traffic
- Periodic AOF sync using configurable interval
- Graceful shutdown with connection cleanup

## Project Structure

```text
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   ├── persistence/
│   ├── protocol/
│   ├── server/
│   └── storage/
├── ARCHITECTURE.md
├── FutureScope.md
└── README.md
```

## Prerequisites

- Go 1.26+

## Run Locally

```bash
go run ./cmd/server
```

Server defaults:
- address: `:6379`
- AOF file: `appendonly.aof`
- sync interval: `1s`

## Run With Docker

Build image:

```bash
docker build -t godis:latest .
```

Run container:

```bash
docker run --name godis \
	-p 6379:6379 \
	-v godis-data:/data \
	godis:latest
```

Notes:
- The container runs with working directory `/data`, so `appendonly.aof` is written there by default.
- The named volume `godis-data` keeps data across container restarts.

Run with Docker Compose:

```bash
docker compose up --build
```

Stop compose:

```bash
docker compose down
```

## Test

```bash
go test ./...
go test -race ./...
```

## Quick Usage

In a second terminal:

```bash
nc localhost 6379
PING
SET lang go
GET lang
EXISTS lang
DEL lang
GET lang
```

## Frontend (Online Interaction)

A simple web frontend will be added so people can interact with Godis online.

Live frontend link:
- https://your-frontend-link.example