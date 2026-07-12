# Listeners

Listeners are the network endpoints that accept inbound grunt connections. They are created, started, and stopped from the dashboard or through the API.

## Lifecycle

1. Create a listener with an address, port, and protocol.
2. Start the listener so it begins accepting connections.
3. Stop the listener when you no longer want it accepting traffic.
4. Restart the listener if you want to resume the same configuration.

The server keeps listener state in memory and in the database. On startup it reloads listeners from the database and restarts any that were previously active.

## Dashboard Actions

- `New Listener` creates a listener with a generated port.
- `Start` starts a saved listener.
- `Stop` stops a running listener.

## API Endpoints

- `GET /api/listeners` lists all listeners.
- `POST /api/listeners` creates a listener.
- `GET /api/listeners/l/?listener_id=...` returns one listener by ID.
- `POST /api/listeners/l/start?id=...` starts a listener by ID.
- `POST /api/listeners/l/stop?id=...` stops a listener by ID.
- `POST /actions/new-listener` creates a listener from the web UI.
- `POST /actions/start-listener?id=...` starts a listener.
- `POST /actions/stop-listener?id=...` stops a listener.

## Curl Examples

List listeners:

```bash
curl -L http://localhost:8080/api/listeners
```

Create a listener:

```bash
curl -X POST http://localhost:8080/api/listeners \
	-H 'Content-Type: application/json' \
	-d '{
            "address":"127.0.0.1",
            "port":"7777",
            "protocol":"tcp"
        }'
```

Get a single listener by ID:

```bash
curl -L 'http://localhost:8080/api/listeners/l?listener_id=LISTENER_ID'
```

Start or stop an existing listener through the API routes:

```bash
curl -X POST 'http://localhost:8080/api/listeners/l/start?id=LISTENER_ID'
curl -X POST 'http://localhost:8080/api/listeners/l/stop?id=LISTENER_ID'
```

Start or stop an existing listener from the action routes:

```bash
curl -X POST 'http://localhost:8080/actions/start-listener?id=LISTENER_ID'
curl -X POST 'http://localhost:8080/actions/stop-listener?id=LISTENER_ID'
```

## Listener Fields

- `address` is the bind address.
- `port` is the listening port.
- `protocol` is the network protocol, usually `tcp`.
- `status` tracks whether the listener is active or inactive.
- `grunt_count` reflects the number of tracked grunts.

## Operational Notes

- A listener can only be started once at a time.
- Stopping a listener closes its network socket and marks it inactive.
- Incoming grunt connections are tracked under the listener that accepted them.
- Listener activity and grunt messages are persisted in the database.
