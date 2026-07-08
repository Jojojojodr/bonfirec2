# Tasks

Tasks let you schedule commands to run against a grunt at a specific time, with optional repetition.

## Purpose

Use tasks when you want a command to run later instead of immediately. The scheduler polls for due tasks and dispatches them to connected grunts.

## Creating A Task

Create a task with:

- `grunt_id` for the target grunt.
- `command` for the command to run.
- `scheduled_for` for the time the task should first run.
- `repeat` to enable repetition.
- `repeat_every_seconds` for the repeat interval.
- `repeat_count` for how many times the task should run.
- `timeout` for the command timeout value stored with the task.

`scheduled_for` accepts either RFC3339 or `2006-01-02 15:04:05`.

## API Endpoints

- `GET /api/tasks` lists all tasks.
- `POST /api/tasks` creates a new task.

## Curl Examples

List tasks:

```bash
curl -L http://localhost:8080/api/tasks
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
            "timeout": 0
        }'
```

## Task States

- `scheduled` means the task is waiting for its run time.
- `waiting` means the scheduler will retry when the grunt is not available.
- `completed` means the task finished its configured run count.
- `cancelled` means the task was cancelled before completion.

## Repeat Behavior

- If `repeat` is false, the task runs once.
- If `repeat` is true, `repeat_every_seconds` controls how often the task is re-queued.
- If `repeat_count` is greater than zero, the task stops after that many runs.
- If `repeat_count` is zero, the task can keep repeating.

## Runtime Notes

- The scheduler runs in the background while the server is running.
- Due tasks are loaded from the database and dispatched to the matching grunt.
- Dispatch activity is recorded in the message history.
