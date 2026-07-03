# Grunts

A grunt is a connected client session managed by a listener. Grunts appear in the dashboard after a client connects and can be messaged from the terminal or through the API.

## What A Grunt Represents

- A grunt is a single connected client tracked by ID.
- Grunts are associated with the listener that accepted the connection.
- Grunt messages are persisted so you can inspect command and response history later.

## Dashboard Views

- The grunts list shows all tracked grunts.
- The grunt detail page shows message history for a single grunt.
- The terminal view lets you send commands to the selected grunt.

## API Endpoints

- `GET /api/grunts` lists all grunts.
- `GET /api/messages/grunt?grunt_id=...` returns message history for one grunt.
- `POST /api/messages/grunt` saves a message for a grunt.
- `POST /actions/grunts/terminal/command?id=...` sends a command from the web UI.

## Curl Examples

List grunts:

```bash
curl -L http://localhost:8080/api/grunts
```

Get message history for one grunt:

```bash
curl -L http://localhost:8080/api/messages/grunt?grunt_id=GRUNT_ID
```

Save a message for a grunt:

```bash
curl -X POST http://localhost:8080/api/messages/grunt \
	-H 'Content-Type: application/json' \
	-d '{
                "grunt_id": "GRUNT_ID",
                "content": "whoami"
            }'
```

Send a command through the web action route:

```bash
curl -X POST 'http://localhost:8080/actions/grunts/terminal/command?id=GRUNT_ID' \
	-d 'command=whoami'
```

## Sending Commands

Commands are written to the live connection for the target grunt.

- Plain commands are sent as-is.
- Slash commands like `/whoami` are parsed before sending.
- `/cmd <command>` can be used for raw command execution on the client side.

## Grunt Status

- A grunt is marked active when it connects.
- A grunt is marked inactive when the connection closes.
- The listener and the stored grunt record are updated together so the dashboard stays in sync.

## Message History

- Messages are stored in creation order.
- The dashboard and API both read from the same message records.
- Command history includes the operator message and the grunt response when available.
