# Godis Architecture and Technical Notes

A learning-first Redis-style key-value server written in Go.

For quick setup and usage, see `README.md`.

The codebase is built in phases with clear package boundaries between protocol, server runtime, storage, and persistence.

## Current Status

- Phase 1: done (project structure + core interfaces)
- Phase 2: done (thread-safe in-memory storage)
- Phase 3: done (TCP server bootstrap + parser + executor + unit tests)
- Phase 4: done (append-only file persistence + replay)
- Phase 5: next (integration tests + hardening)

## Implemented Features

- Command support: `PING`, `SET`, `GET`, `DEL`, `EXISTS`
- Concurrent client handling (goroutine per connection)
- In-memory key/value store with `sync.RWMutex`
- Line-based command parsing and validation
- Protocol response helpers (`+`, `-ERR`, `:`, `$`, `$-1`)
- Append-only file (AOF) persistence for successful `SET`/`DEL`
- Replay on startup before accepting traffic
- Periodic AOF sync using configurable interval
- Graceful shutdown on `SIGINT` / `SIGTERM`
- Unit tests for storage, protocol, executor, and AOF behavior

## Architecture

### 1) Configuration

File: `internal/config/config.go`

`Config` provides runtime defaults:
- `Address`: `:6379`
- `AOFPath`: `appendonly.aof`
- `AOFSyncInterval`: `1s`

### 2) Storage Contract and In-Memory Implementation

Files:
- `internal/storage/storage.go`
- `internal/storage/operation.go`

Storage interface:
- `Set(key, value string) error`
- `Get(key string) (string, error)`
- `Delete(key string) error`
- `Exists(key string) bool`

Errors:
- `ErrKeyNotFound`
- `ErrEmptyKey`

`MemoryStorage` uses `map[string]string` protected by `sync.RWMutex`.

### 3) Protocol Layer

Files:
- `internal/protocol/command.go`
- `internal/protocol/parser.go`
- `internal/protocol/response.go`

Responsibilities:
- Parse one line command into `Command{Name, Args, Raw}`
- Validate command arity
- Format consistent protocol responses

### 4) Server Runtime and Command Execution

Files:
- `internal/server/server.go`
- `internal/server/execute.go`
- `cmd/server/main.go`

Flow:
1. `main` builds config, storage, and AOF.
2. Existing AOF entries are replayed into storage.
3. Server is created with an append hook injected into the executor.
4. `Server.Start()` listens on TCP and accepts clients.
5. Each client connection runs in its own goroutine.
6. Incoming line is parsed, executed, and response is written back.
7. Shutdown closes listener, active connections, waits for goroutines, and closes AOF.

### 5) Persistence (AOF)

Files:
- `internal/persistence/aof.go`
- `internal/persistence/aof_test.go`

Behavior:
- Appends one successful write command per line (`SET`, `DEL`).
- Skips malformed replay lines with warning logs (startup continues).
- Replays only write commands into storage.
- Performs periodic `Sync()` based on `AOFSyncInterval`.
- Final `Sync()` and file close are done during shutdown.

## Command Behavior (v1)

- `PING` -> `+PONG\r\n`
- `SET key value` -> `+OK\r\n`
- `GET key` -> `$<len>\r\n<value>\r\n` or `$-1\r\n`
- `DEL key` -> `:1\r\n` if deleted, `:0\r\n` if not found
- `EXISTS key` -> `:1\r\n` or `:0\r\n`

## Project Layout

```text
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── persistence/
│   │   ├── aof.go
│   │   └── aof_test.go
│   ├── protocol/
│   │   ├── command.go
│   │   ├── parser.go
│   │   ├── parser_test.go
│   │   ├── response.go
│   │   └── response_test.go
│   ├── server/
│   │   ├── execute.go
│   │   ├── execute_test.go
│   │   └── server.go
│   └── storage/
│       ├── operation.go
│       ├── operation_test.go
│       └── storage.go
└── go.mod
```

## Run and Test

From project root:

```bash
go test ./...
go test -race ./...
go run ./cmd/server
```

Quick manual check with netcat (in another terminal while server is running):

```bash
nc localhost 6379
PING
SET lang go
GET lang
EXISTS lang
DEL lang
GET lang
```

After restarting the server, `GET lang` should return `$-1\r\n` because the key was deleted and that delete was persisted.

To verify persistence recovery, restart after a write without deleting:

```bash
nc localhost 6379
SET boot persisted
```

Then restart server and run:

```bash
nc localhost 6379
GET boot
```

Expected response contains `persisted`.

## Notes

- This is intentionally a simple text protocol implementation for learning.
- RESP2 compatibility, snapshots, and replication are future extensions.
- A small naming cleanup (`operation.go` -> `memory.go`) can be done later in a dedicated refactor.
