## BonFire C2 Docs

BonFire C2 is a local educational dashboard for managing listeners, connected grunts, and scheduled tasks.

### Quick Start

1. Copy the example config into place.
	```bash
	cp config.example.yaml config.yaml
	```
2. Start the server.
	```bash
	task dev
	```
3. Start the client in another terminal.
	```bash
	task client
	```
4. Open the dashboard in your browser.

### What To Use

- [Listeners](listeners.md) for creating, starting, and stopping network listeners.
- [Grunts](grunts.md) for viewing connected clients and sending commands.
- [Tasks](tasks.md) for scheduling commands to run later or repeat.
- [Client](client.md) for running the sample client and understanding its flags.

### API Summary

- `GET /api/health` returns the health status.
- `GET /api/listeners` and `POST /api/listeners` manage listeners.
- `GET /api/grunts` lists connected grunts.
- `GET /api/messages` and `GET /api/messages/grunt?grunt_id=...` return message history.
- `GET /api/tasks` and `POST /api/tasks` manage scheduled tasks.

### Curl Examples

Use `http://localhost:8080` as the base URL if you are running the server with the default config.

```bash
curl -L http://localhost:8080/api/health
curl -L http://localhost:8080/api/listeners
curl -L http://localhost:8080/api/grunts
curl -L http://localhost:8080/api/tasks
```

Create a listener:

```bash
curl -X POST http://localhost:8080/api/listeners \
	-H 'Content-Type: application/json' \
	-d '{
                "address": "127.0.0.1",
                "port": "7777",
                "protocol": "tcp",
            }'
```

Create a task:

```bash
curl -X POST http://localhost:8080/api/tasks \
	-H 'Content-Type: application/json' \
	-d '{
                "grunt_id": "GRUNT_ID",
                "command": "whoami",
                "scheduled_for": "2026-07-04T12:00:00Z",
                "repeat": false,
                "repeat_every_seconds": 0,
                "repeat_count": 1,
                "timeout": 0,
             }'
```

### Notes

- The app is intended for local use only.
- The sample client defaults to connecting to `localhost` on port `7777`.
