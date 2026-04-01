# Mini-Redis (Go)

A learning-first Redis-style key-value server in Go.

This project is being built in phases, with clean architecture boundaries between:
- protocol parsing/formatting
- command execution/server runtime
- storage engine
- persistence

## Current Status

- Phase 1: done (project structure + contracts)
- Phase 2: done (thread-safe in-memory storage)
- Phase 3: in progress (server + protocol wiring)
- Phase 4: pending (append-only file persistence)
- Phase 5: pending (hardening + integration tests)

## Architecture

## 1) Storage Interface (contract-first)

File: `internal/storage/storage.go`

Defined behavior:
- `Set(key, value string) error`
- `Get(key string) (string, error)`
- `Delete(key string) error`
- `Exists(key string) bool`

Shared storage errors:
- `ErrKeyNotFound`
- `ErrEmptyKey`

Why this matters:
- Server code depends on `Storage` interface, not implementation details.
- You can plug in additional backends later (disk-backed store, distributed store, etc.).

## 2) In-Memory Engine + Concurrency

File: `internal/storage/operation.go`

Implemented type:
- `MemoryStorage`

Data model:
- `map[string]string`

Concurrency model:
- guarded by `sync.RWMutex`
- read operations (`Get`, `Exists`) use `RLock`
- write operations (`Set`, `Delete`) use `Lock`

Behavior:
- empty keys return `ErrEmptyKey`
- missing keys on `Get`/`Delete` return `ErrKeyNotFound`

## 3) Protocol Command Model

File: `internal/protocol/command.go`

Supported command names (v1 scope):
- `PING`
- `SET`
- `GET`
- `DEL`
- `EXISTS`

`Command` fields:
- `Name`
- `Args`
- `Raw`

Helper:
- `IsWrite()` returns true for `SET` and `DEL`.

## 4) Runtime Config

File: `internal/config/config.go`

Default values:
- Address: `:6379`
- AOF path: `appendonly.aof`
- AOF sync interval: `1s`

## 5) Process Entry Point

File: `cmd/server/main.go`

Current behavior:
- loads default config
- logs boot configuration

Note:
- full TCP server wiring is part of Phase 3 (next steps below).

## Directory Layout

```text
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── persistence/        # planned for Phase 4
│   ├── protocol/
│   │   ├── command.go
│   │   ├── parser.go       # placeholder, Phase 3 target
│   │   ├── response.go     # placeholder, Phase 3 target
│   │   └── tests/
│   │       ├── parser_test.go
│   │       └── response_test.go
│   ├── server/             # Phase 3 target
│   └── storage/
│       ├── operation.go
│       ├── storage.go
│       └── tests/
│           └── operation_test.go
└── go.mod
```

## Roadmap

## Phase 3 (Now)

Goal: add a working TCP command server.

Planned additions:
- parser implementation in `internal/protocol/parser.go`
- response formatting helpers in `internal/protocol/response.go`
- server runtime in `internal/server/server.go`
- command executor in `internal/server/execute.go`
- wiring + graceful shutdown in `cmd/server/main.go`

Execution flow:
1. accept TCP connection
2. read command line
3. parse into `protocol.Command`
4. execute via `storage.Storage`
5. write protocol response

## Phase 4

Goal: durability with append-only file (AOF).

Planned additions:
- `internal/persistence/aof.go`
- append mutating commands (`SET`, `DEL`)
- periodic `Sync()` using config interval
- startup replay before accepting clients

## Phase 5

Goal: hardening and confidence.

Planned work:
- parser tests (valid + invalid command/arity)
- executor tests for all v1 commands
- persistence replay tests
- integration tests with concurrent clients
- race gate + full test gate

## How to Run (current)

From project root:

```bash
go test ./...
go test -race ./...
go run ./cmd/server
```

Expected right now:
- tests should pass for implemented packages
- server binary currently logs startup config only

## Immediate TODO (high priority)

1. Implement `internal/protocol/parser.go`.
2. Implement `internal/protocol/response.go`.
3. Add protocol tests under `internal/protocol/tests/`.
4. Build server runtime and execution path.

## Notes

- This repo currently uses a simple text protocol plan for v1, not full RESP2.
- Keep interfaces stable while features evolve; avoid coupling transport and storage layers.
- Type/file naming can be cleaned later (`operation.go` -> `memory.go`) in a dedicated refactor step.
