# Task CLI API Documentation

This document describes the command-line interface and available commands for the Task CLI application.

## Global Flags

These flags are available for all commands:

- `--config string`: Configuration file path (default: `$HOME/.config/task-cli/task-cli.yaml`)
- `--redis-url string`: Redis connection URL (overrides config)
- `--output string`: Output format (table, json, yaml, csv)
- `--no-color`: Disable colored output
- `--verbose`: Enable verbose output

## Commands

### `task-cli create`

Create a new task.

**Usage:**
```bash
task-cli create [title] [flags]
```

**Flags:**
- `-d, --description string`: Task description
- `-p, --priority string`: Task priority (low, medium, high, critical)
- `-a, --assignee string`: Task assignee
- `--due string`: Due date (YYYY-MM-DD or YYYY-MM-DD HH:MM)
- `-t, --tags stringSlice`: Task tags (comma-separated)
- `-i, --interactive`: Interactive mode for task creation

**Examples:**
```bash
task-cli create "Fix login bug"
task-cli create "Fix login bug" --priority high --assignee john --due "2024-01-15 14:00"
task-cli create --interactive
```

### `task-cli list`

List tasks with filtering and sorting options.

**Usage:**
```bash
task-cli list [flags]
```

**Flags:**
- `-s, --status stringSlice`: Filter by status
- `-p, --priority stringSlice`: Filter by priority
- `-a, --assignee string`: Filter by assignee
- `-c, --creator string`: Filter by creator
- `-t, --tags stringSlice`: Filter by tags
- `--due-before string`: Filter tasks due before date
- `--due-after string`: Filter tasks due after date
- `--sort-by string`: Sort by field (created_at, updated_at, due_date, priority, status)
- `--sort-order string`: Sort order (asc, desc)
- `-l, --limit int`: Maximum number of results (default 20)
- `-o, --offset int`: Number of results to skip
- `--my-tasks`: Show only tasks assigned to current user

**Examples:**
```bash
task-cli list
task-cli list --status pending,in-progress
task-cli list --priority high --assignee john
task-cli list --sort-by priority --sort-order desc
```

### `task-cli search`

Search tasks by title, description, or tags.

**Usage:**
```bash
task-cli search [query] [flags]
```

**Flags:**
- `-a, --assignee string`: Filter by assignee
- `-s, --status stringSlice`: Filter by status
- `-p, --priority stringSlice`: Filter by priority
- `-t, --tags stringSlice`: Filter by tags
- `-l, --limit int`: Maximum number of results (default 20)
- `-o, --offset int`: Number of results to skip
- `--sort-by string`: Sort by field
- `--sort-order string`: Sort order
- `--show-completed`: Include completed tasks in search

**Examples:**
```bash
task-cli search "bug"
task-cli search "API" --status pending --priority high
task-cli search "urgent" --limit 5
```

### `task-cli update`

Update an existing task.

**Usage:**
```bash
task-cli update [task-id] [flags]
```

**Flags:**
- `-t, --title string`: New task title
- `-d, --description string`: New task description
- `-s, --status string`: New task status
- `-p, --priority string`: New task priority
- `-a, --assignee string`: New task assignee
- `--tags stringSlice`: New task tags
- `--due string`: New due date
- `-i, --interactive`: Interactive mode for task update

**Examples:**
```bash
task-cli update 123 --status completed
task-cli update 123 --priority high --due "2024-01-20"
task-cli update 123 --interactive
```

### `task-cli status`

Change the status of a task.

**Usage:**
```bash
task-cli status [task-id] [new-status] [flags]
```

**Flags:**
- `-i, --interactive`: Interactive mode for status change
- `--bulk`: Bulk status change mode
- `--status string`: Current status filter for bulk operations
- `--new-status string`: New status for bulk operations
- `-a, --assignee string`: Filter by assignee for bulk operations
- `-p, --priority stringSlice`: Filter by priority for bulk operations
- `-t, --tags stringSlice`: Filter by tags for bulk operations
- `--confirm`: Skip confirmation for bulk operations

**Valid Statuses:**
- `pending`
- `in-progress`
- `completed`
- `cancelled`
- `on-hold`

**Examples:**
```bash
task-cli status 123 completed
task-cli status --interactive
task-cli status --bulk --status pending --new-status in-progress
```

### `task-cli assign`

Assign a task to a user.

**Usage:**
```bash
task-cli assign [task-id] [assignee] [flags]
```

**Flags:**
- `--bulk`: Bulk assign tasks based on filters
- `--status string`: Filter tasks by status for bulk assignment
- `--priority string`: Filter tasks by priority for bulk assignment
- `--creator string`: Filter tasks by creator for bulk assignment
- `--tags stringSlice`: Filter tasks by tags for bulk assignment
- `--assignee string`: New assignee for bulk assignment
- `--unassign`: Remove assignee from task(s)

**Examples:**
```bash
task-cli assign 123 john
task-cli assign --bulk --status pending --assignee alice
task-cli assign 123 --unassign
```

### `task-cli delete`

Delete one or more tasks.

**Usage:**
```bash
task-cli delete [task-id...] [flags]
```

**Flags:**
- `-f, --force`: Skip confirmation prompt
- `--all`: Delete all tasks (requires --force)
- `--status string`: Delete all tasks with specific status
- `--assignee string`: Delete all tasks assigned to specific user

**Examples:**
```bash
task-cli delete 123
task-cli delete 123 456 789
task-cli delete --status cancelled --force
task-cli delete --all --force
```

### `task-cli stats`

Display task statistics and analytics.

**Usage:**
```bash
task-cli stats [flags]
```

**Flags:**
- `--detailed`: Show detailed statistics
- `--chart`: Display ASCII charts

**Examples:**
```bash
task-cli stats
task-cli stats --detailed --chart
task-cli stats --output json
```

### `task-cli export`

Export tasks to various formats.

**Usage:**
```bash
task-cli export [filename] [flags]
```

**Flags:**
- `-f, --format string`: Export format (json, csv, yaml)
- `-s, --status stringSlice`: Filter by status
- `-p, --priority stringSlice`: Filter by priority
- `-a, --assignee string`: Filter by assignee
- `-t, --tags stringSlice`: Filter by tags
- `--due-before string`: Filter tasks due before date
- `--due-after string`: Filter tasks due after date
- `--all`: Export all tasks (ignore other filters)
- `--pretty`: Pretty print JSON output

**Examples:**
```bash
task-cli export tasks.json
task-cli export tasks.csv --status completed
task-cli export --format yaml --assignee john
task-cli export backup.json --all --pretty
```

### `task-cli config`

Manage configuration settings.

**Subcommands:**
- `init`: Create default configuration file
- `show`: Show current configuration
- `path`: Show configuration file path

**Examples:**
```bash
task-cli config init
task-cli config show
task-cli config path
```

### `task-cli version`

Print version information.

**Usage:**
```bash
task-cli version
```

## Output Formats

### Table Format (Default)

Displays tasks in a formatted table with columns for ID, Title, Status, Priority, Assignee, and Due Date.

### JSON Format

Returns tasks as JSON array:
```json
[
  {
    "id": "123",
    "title": "Fix login bug",
    "description": "Users can't login with email",
    "status": "pending",
    "priority": "high",
    "assignee": "john",
    "creator": "admin",
    "tags": ["bug", "urgent"],
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z",
    "due_date": "2024-01-15T14:00:00Z"
  }
]
```

### YAML Format

Returns tasks in YAML format:
```yaml
- id: "123"
  title: "Fix login bug"
  description: "Users can't login with email"
  status: "pending"
  priority: "high"
  assignee: "john"
  creator: "admin"
  tags:
    - "bug"
    - "urgent"
  created_at: "2024-01-01T10:00:00Z"
  updated_at: "2024-01-01T10:00:00Z"
  due_date: "2024-01-15T14:00:00Z"
```

### CSV Format

Returns tasks as comma-separated values with headers:
```csv
ID,Title,Description,Status,Priority,Assignee,Creator,Tags,Created At,Updated At,Due Date,Completed At
123,Fix login bug,Users can't login with email,pending,high,john,admin,bug;urgent,2024-01-01T10:00:00Z,2024-01-01T10:00:00Z,2024-01-15T14:00:00Z,
```

## Exit Codes

- `0`: Success
- `1`: General error
- `2`: Configuration error
- `3`: Redis connection error
- `4`: Invalid arguments
- `5`: Task not found
