# Client

The sample client connects back to the server and executes commands it receives from a listener.

## Running The Client

Start the client with the task runner:

```bash
task client
```

You can also run it directly:

```bash
go run ./cmd/client
```

## Flags

- `-port` sets the server port to connect to. The default is `7777`.
- `-address` sets the server address. The default is `localhost`.
- `-local-port` sets the local source port used by the client. The default is `10007`.

Example:

```bash
go run ./cmd/client -address localhost -port 7777 -local-port 10007
```

## Connection Behavior

- The client reconnects automatically when the connection closes.
- The client logs the local socket address after connecting.
- If the local source port is unavailable, the connection may fail until the port becomes free.

## Commands

The client can execute commands sent by the server.

- Slash commands such as `/whoami` are parsed before execution.
- `/cmd <command>` allows raw command execution on the client.
- Built-in commands are resolved through the command helper package before execution.

## Notes

- The client is intended for local educational use.
- Keep the client and server running on the same machine if you want to use the default localhost settings.
