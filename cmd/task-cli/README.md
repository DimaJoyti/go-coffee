# Task CLI - Powerful Command-Line Task Manager

Task CLI is a feature-rich command-line task manager that uses Redis for data storage. It supports creating, updating, searching, and managing tasks with advanced filtering, sorting, and export capabilities.

## 🚀 Features

- ✅ **Complete Task Management**: create, update, delete, assign tasks
- 🔍 **Powerful Search**: search by title, description, tags
- 📊 **Statistics & Analytics**: detailed statistics by status, priority, assignees
- 📤 **Data Export**: JSON, CSV, YAML formats
- 🎨 **Colorful Interface**: intuitive colored output
- ⚡ **High Performance**: Redis backend for speed
- 🔧 **Flexible Configuration**: configuration files and environment variables

## 📦 Installation

### Build from Source

```bash
# Clone repository
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee

# Build CLI
make -f Makefile.task-cli build

# Or manually
go build -o build/task-cli ./cmd/task-cli
```

### System Installation

```bash
# Install to /usr/local/bin
make -f Makefile.task-cli install

# Or manually
sudo cp build/task-cli /usr/local/bin/
```

## 🔧 Setup

### Redis

Task CLI requires Redis to operate. Start Redis locally:

```bash
# Using Docker
docker run -d --name task-cli-redis -p 6379:6379 redis:7-alpine

# Or via Makefile
make -f Makefile.task-cli run-with-redis
```

### Configuration

Create configuration file:

```bash
task-cli config init
```

This creates `~/.config/task-cli/task-cli.yaml` with default settings.

## 📚 Usage

### Basic Commands

```bash
# Create task
task-cli create "Fix login bug" --priority high --assignee john

# List tasks
task-cli list

# Search tasks
task-cli search "bug"

# Update task
task-cli update 123 --status completed

# Change status
task-cli status 123 completed

# Assign task
task-cli assign 123 john

# Delete task
task-cli delete 123

# Statistics
task-cli stats

# Export
task-cli export tasks.json
```

### Filtering and Sorting

```bash
# Filter by status
task-cli list --status pending,in-progress

# Filter by priority
task-cli list --priority high,critical

# Filter by assignee
task-cli list --assignee john

# Sorting
task-cli list --sort-by priority --sort-order desc

# Pagination
task-cli list --limit 10 --offset 20
```

### Search

```bash
# Simple search
task-cli search "bug"

# Search with filters
task-cli search "API" --status pending --priority high

# Search with result limit
task-cli search "urgent" --limit 5
```

### Export

```bash
# Export to JSON
task-cli export tasks.json

# Export to CSV
task-cli export tasks.csv --status completed

# Export to YAML
task-cli export --format yaml --assignee john

# Export with filters
task-cli export backup.json --due-before 2024-01-01
```

### Bulk Operations

```bash
# Bulk assignment
task-cli assign --bulk --status pending --assignee john

# Bulk status change
task-cli status --bulk --status pending --new-status in-progress

# Bulk deletion
task-cli delete --status cancelled --force
```

## 🎨 Output Formats

Task CLI supports various output formats:

```bash
# Table (default)
task-cli list

# JSON
task-cli list --output json

# YAML
task-cli list --output yaml

# CSV
task-cli list --output csv
```

## ⚙️ Configuration

### Configuration File

```yaml
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

cli:
  default_user: "username"
  date_format: "2006-01-02 15:04"
  output_format: "table"
  color_output: true
  page_size: 20
  sort_by: "created_at"
  sort_order: "desc"

defaults:
  priority: "medium"
  status: "pending"
  tags: []
```

### Environment Variables

```bash
export REDIS_URL="redis://localhost:6379"
export TASK_CLI_DEFAULT_USER="john"
export TASK_CLI_OUTPUT_FORMAT="json"
```

## 🔍 Usage Examples

### Daily Workflow

```bash
# View today's tasks
task-cli list --due-today

# Create new task
task-cli create "Review PR #123" --priority medium --due "2024-01-15 14:00"

# Start working on task
task-cli status 456 in-progress

# Complete task
task-cli status 456 completed

# View statistics
task-cli stats
```

### Team Collaboration

```bash
# Assign tasks to team
task-cli assign --bulk --status pending --assignee alice
task-cli assign --bulk --priority high --assignee bob

# View team tasks
task-cli list --assignee alice,bob

# Export report for manager
task-cli export weekly-report.csv --due-after 2024-01-01 --due-before 2024-01-07
```

## 🛠️ Development

### Building

```bash
# Build for current platform
make -f Makefile.task-cli build

# Build for all platforms
make -f Makefile.task-cli build-all

# Run tests
make -f Makefile.task-cli test

# Format code
make -f Makefile.task-cli fmt
```

### Testing

```bash
# Start with Redis in Docker
make -f Makefile.task-cli run-with-redis

# Test CLI
./build/task-cli config init
./build/task-cli create "Test task"
./build/task-cli list
```

## 📝 License

MIT License

## 🤝 Contributing

Contributions are welcome! Please create an issue or pull request.

## 📞 Support

If you have questions or issues, please create an issue in the GitHub repository.
