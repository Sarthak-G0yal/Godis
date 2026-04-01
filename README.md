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
Future roadmap and Phase 5 plan: [ignore.md](ignore.md)

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
├── web-ui/
│   ├── app.py
│   ├── pyproject.toml
│   ├── uv.lock
│   └── modules/
├── ARCHITECTURE.md
├── ignore.md
└── README.md
```

## Prerequisites

- Go 1.26+
- Python 3.11+
- `uv`

## Run Locally

```bash
go run ./cmd/server
```

Server defaults:
- address: `:6379`
- AOF file: `appendonly.aof`
- sync interval: `1s`

## Run Frontend Locally (uv)

In another terminal:

```bash
cd web-ui
uv sync
GODIS_HOST=localhost GODIS_PORT=6379 uv run streamlit run app.py
```

Open `http://localhost:8501`.

The frontend auto-configures the server target from environment variables:
- `GODIS_HOST` (default: `godis`)
- `GODIS_PORT` (default: `6379`)

Important:
- Start the UI with `uv run streamlit run app.py`.
- Do not run `uv run python app.py` directly.

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

Run full stack with Docker Compose (backend + frontend):

```bash
docker compose up --build
```

Open:
- Godis UI: `http://localhost:8501`

Compose networking behavior:
- `godis-ui` is publicly published on `8501`.
- `godis` is internal-only (`expose: 6379`) on the Docker network.
- `godis-ui` reaches `godis` using service DNS name `godis`.

Optional Compose environment overrides:
- `UI_BIND_HOST` (default: `0.0.0.0`)
- `UI_PORT` (default: `8501`)
- `GODIS_CPUS`, `GODIS_MEM`, `GODIS_UI_CPUS`, `GODIS_UI_MEM`

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

For local backend run (`go run ./cmd/server`) in a second terminal:

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

Godis now includes a Streamlit frontend in `web-ui` for local and Docker-based interaction.

Live frontend link:
- https://godis.sarthakgoyal.tech