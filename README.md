# Factory Reader

A Go service that reads data from industrial protocol devices and streams results to a web dashboard via WebSocket.

## Supported Protocols

- **MQTT** — subscribes to a topic, returns the latest message
- **OPC-UA** — reads a node value by node ID
- **Modbus TCP** — reads holding registers, returns `[]uint16`
- **Siemens S7** — reads a data block via `AGReadDB`, returns raw bytes

## Prerequisites

- Go 1.22+
- PostgreSQL

## Getting Started

1. Start a PostgreSQL instance (the `connections` table is auto-created on boot):

```bash
# example with Docker
docker run -d --name factory-pg -p 5432:5432 \
  -e POSTGRES_DB=factory_reader \
  -e POSTGRES_PASSWORD=postgres \
  postgres:16
```

2. Run the service:

```bash
go run .
```

3. Open http://localhost:8080 in your browser.

## Configuration

| Variable       | Default                                                                      | Description          |
|----------------|------------------------------------------------------------------------------|----------------------|
| `SERVER_ADDR`  | `:8080`                                                                      | HTTP listen address  |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/factory_reader?sslmode=disable`  | PostgreSQL connection string |

## API

| Method   | Path                      | Description                        |
|----------|---------------------------|------------------------------------|
| `GET`    | `/`                       | Web dashboard                      |
| `GET`    | `/api/connections`        | List all connections               |
| `POST`   | `/api/connections`        | Create a connection                |
| `DELETE` | `/api/connections/{id}`   | Delete a connection                |
| `POST`   | `/api/test/{id}`          | One-shot read from a connection    |
| `POST`   | `/api/monitor/{id}`       | Start continuous monitoring (2s interval) |
| `DELETE` | `/api/monitor/{id}`       | Stop monitoring                    |
| `GET`    | `/api/monitor`            | List actively monitored IDs        |
| `GET`    | `/ws`                     | WebSocket for real-time results    |

### Create a connection

```bash
curl -X POST http://localhost:8080/api/connections \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Temperature Sensor",
    "protocol": "mqtt",
    "host": "broker.local",
    "port": 1883,
    "topic": "factory/temp"
  }'
```

Protocol-specific fields:

| Field          | Used by       |
|----------------|---------------|
| `topic`        | mqtt          |
| `node_id`      | opcua         |
| `unit_id`, `start_address`, `quantity` | modbus |
| `rack`, `slot`, `db_number`, `start_address`, `quantity` | s7 |

### WebSocket

Connect to `ws://localhost:8080/ws` to receive JSON messages as connections are tested or monitored:

```json
{
  "connection_id": 1,
  "name": "Temperature Sensor",
  "protocol": "mqtt",
  "host": "broker.local",
  "port": 1883,
  "connected": true,
  "data": "23.5",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

## License

See [LICENSE](LICENSE).