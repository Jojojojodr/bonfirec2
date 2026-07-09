# BonFire C2

<div align="center">
  <img src="/static/images/logoDark.png" alt="BonfireC2" width="250">
</div>

A local educational network operations dashboard built with Go, featuring a modern web interface for managing listeners and viewing system status.

## Overview

BonFire C2 is a clean, polished dashboard application designed for educational purposes and local development. It provides a web-based interface to create and monitor network listeners, track system status, and view real-time activity logs.

**Disclaimer:** This project is for local educational use only. Ensure you comply with all applicable laws and regulations when running network listeners or similar operations.

## Features

- 🎨 **Modern Dashboard UI** - Built with Tailwind CSS and Templ components
- 🔊 **Listener Management** - Create, start, and monitor TCP/network listeners
- 📊 **Real-time Monitoring** - View active listeners with status and metrics
- ⏰ **Scheduled Tasks** - Queue commands for a future time and repeat them on an interval
- 💾 **Persistent Storage** - SQLite or PostgreSQL database support
- 🌓 **Dark Theme** - Eye-friendly interface with customizable components
- 🔄 **Hot Reload** - Automatic asset compilation with Air for development
- 📱 **Responsive Design** - Works on desktop and mobile devices

## Prerequisites

- **Go** 1.26.3 or higher
- **Node.js** 16+ (for Tailwind CSS and asset building)
- **Python** 3.14.6 of higher (for running the Setup script)
- **Task** (task runner) - Install from [taskfile.dev](https://taskfile.dev)
- **Templ** (Go HTML templating) - Install from [templ.guide](https://templ.guide)
- **Air** (Go Live Reloading) - Install from [github.com](https://github.com/air-verse/air)

## Quick Start

### 1. Clone and Navigate

```bash
git clone https://github.com/Jojojojodr/bonfirec2.git
cd bonfirec2
```

### 2. Run Setup Script

This setup script:

- copies `config.example.yaml` to `config.yaml` if `config.yaml` does not already exist
- installs Go dependencies
- installs npm dependencies

```bash
python3 scripts/setup.py
```

### 3. Configure

Edit `config.yaml`:

```yaml
server:
  port: "8080"
  debug: false

database:
  type: "sqlite"
  dsn: "./data/bonfire.db"
```

### 4. Run

Development mode with hot reload:

```bash
task dev
```

Run the sample client in a separate terminal (optional):

```bash
task client
```

Or build and run:

```bash
task build
task run
```

Visit `http://localhost:8080` in your browser.

## Development

### Available Tasks

```bash
task # Run the code
task client # Run the client code
task dev # Development server with hot reload
task build # Build the application
task run # Run the built application
task run-client # Run the built client
task clean # Clean build artifacts
```

### Code Organization

- **Templates** - Located in `web/` using Templ syntax (`.templ` files)
- **Components** - Reusable UI components in `web/components/`
- **Controllers** - HTTP handlers in `controller/`
- **Models** - Data structures in root package and `models/`
- **Routes** - Endpoint definitions in `router/`

### Styling

The project uses **Tailwind CSS v4.3.1** for styling:

- Tailwind configuration: `tailwind.config.js`
- Input CSS: `static/input.css`
- Compiled CSS: `static/styles.css` (generated)

Build CSS:
```bash
task tasks:build-css
```

Watch CSS during development:
```bash
task tasks:watch-css
```

### Templates

Templates are written with **Templ**, a Go HTML templating DSL that generates type-safe Go code:

```bash
# Generate Go code from .templ files
templ generate
```

## Database

### Supported Databases

- **SQLite** (default) - File-based, no setup required
- **PostgreSQL** - For production deployments

### Configuration

**SQLite:**
```yaml
database:
  type: "sqlite"
  dsn: "./data/bonfire.db"
```

**PostgreSQL:**
```yaml
database:
  type: "postgres"
  dsn: "postgres://user:password@localhost:5432/bonfire"
```

### Data Directory

The application automatically creates a `./data/` directory for SQLite databases if it doesn't exist.

## Configuration Options

### Server

- `port` - HTTP server port (default: 8080)
- `debug` - Enable debug logging (default: false)

### Database

- `type` - Database backend ("sqlite" or "postgres")
- `dsn` - Database connection string or file path

## Troubleshooting

### Port Already in Use

Change the port in `config.yaml`:
```yaml
server:
  port: "9090"
```

### Database Connection Issues

- Ensure `./data/` directory exists and is writable
- For PostgreSQL, verify connection string in `config.yaml`
- Check database logs for detailed error messages

### Template Compilation Errors

Regenerate templates:
```bash
templ generate
```

### Client Cannot Connect

- Ensure at least one listener is started in the UI
- Verify host/port values used by the client match the active listener
- Check server logs for connection errors

### CSS Not Loading

Rebuild Tailwind:
```bash
task tasks:build-css
```

## Contributing

Contributions are welcome! Please ensure:
1. Code follows Go conventions
2. Templates use Templ syntax
3. Styles use Tailwind utilities
4. Changes include appropriate tests

## License

This project is provided as-is for educational purposes. Use at your own risk and in compliance with all applicable laws and regulations.

## Support

For issues, questions, or suggestions, please open an issue on the repository.
