# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

factory-reader is a Go service that reads data from industrial protocol devices (MQTT, OPC-UA, Modbus TCP, Siemens S7) and streams results to a web UI via WebSocket. Connections are stored in PostgreSQL and can be tested once or continuously monitored.

## Build & Run

```bash
go build -o factory-reader .    # build binary
go run .                        # run directly (serves on :8080 by default)
go vet ./...                    # lint
go test ./...                   # run all tests
```

No test files exist yet. There are no Makefile, Docker, or CI configs.

## Environment

| Variable       | Default                                                                  |
|----------------|--------------------------------------------------------------------------|
| `SERVER_ADDR`  | `:8080`                                                                  |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/factory_reader?sslmode=disable` |

PostgreSQL must be running before start. The app auto-migrates the `connections` table on boot.

## Architecture

**Startup flow:** `main.go` loads config → connects to Postgres → runs migrations → starts HTTP server.

**Key packages:**

- `reader/` — `Reader` interface (`Connect`, `Read`, `Close`) with four implementations: `MQTTReader`, `OPCUAReader`, `ModbusReader`, `S7Reader`. Each wraps a protocol-specific client library. To add a new protocol, implement `Reader` and add a case in `Server.createReader()`.
- `server/` — HTTP server, REST handlers, WebSocket hub, and monitor loop. `Hub` manages WebSocket clients and broadcasts `reader.Result` as JSON. `Monitor` runs a goroutine per connection that reads every 2 seconds and broadcasts via the hub.
- `db/` — Thin wrapper around `pgxpool`. Migrations are inline SQL in `migrations.go`.
- `config/` — Reads env vars with fallbacks.

**REST API routes** (defined in `server.go:routes()`):

- `GET /` — serves `templates/index.html`
- `GET|POST|DELETE /api/connections[/{id}]` — CRUD for connections
- `POST /api/test/{id}` — one-shot read from a connection
- `POST|DELETE /api/monitor/{id}` — start/stop continuous monitoring
- `GET /api/monitor` — list actively monitored connection IDs
- `GET /ws` — WebSocket endpoint for real-time results

**Data flow for monitoring:** HTTP start request → `Monitor.Start()` spawns goroutine → goroutine creates `Reader`, polls every 2s → results broadcast via `Hub` → WebSocket clients receive JSON.

## Protocol-specific notes

- All readers use 5-second connect/read timeouts.
- MQTT subscribes to a topic and caches the last message; `Read()` returns the cached value.
- Modbus reads holding registers and converts raw bytes to `[]uint16`.
- S7 uses `AGReadDB` to read a data block; returns raw `[]byte`.
- The `Connection` struct in `handlers.go` carries fields for all protocols (topic, node_id, unit_id, rack, slot, etc.) — unused fields default to zero values.