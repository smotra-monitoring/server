# Smotra Monitoring Server - Quick Start Guide

This guide will help you get the server running quickly.

## Prerequisites

- Go 1.23 or later
- For PostgreSQL: PostgreSQL 13+ with TimescaleDB extension (optional for production)
- For SQLite: No additional dependencies required (default for development)

## Quick Start

### 1. Clone and Setup

```bash
# Clone the repository
git clone https://github.com/smotra-monitoring/server.git
cd server
```

### 2. Create Configuration File

```bash
# Copy the example configuration file
cp config.example.yaml config.yaml

# Edit config.yaml with your settings (optional, defaults work for development)
# nano config.yaml
```

### 3. Run with SQLite (Development)

The default configuration uses SQLite, which requires no additional setup:

```bash
# Run with config file
CONFIG_FILE=config.yaml go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

### 4. Test the Server

```bash
# Health check
curl http://localhost:8080/healthz

# Readiness check
curl http://localhost:8080/healthz/ready

# Liveness check
curl http://localhost:8080/healthz/live

# API info
curl http://localhost:8080/api/v1
```

### 5. Using PostgreSQL (Production)

Edit your `config.yaml` file:

```yaml
server:
  environment: production

database_type: postgres

postgres_config:
  type: postgres
  host: localhost
  port: 5432
  username: smotra
  password: your_password
  database: smotra
  sslmode: require
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
  conn_max_idle_time: 10m
```

Then start the server:

```bash
CONFIG_FILE=config.yaml go run cmd/server/main.go
```

## Configuration

The server is configured using a YAML or JSON configuration file. The path to the configuration file must be specified via the `CONFIG_FILE` environment variable.

### Configuration File Format

Both YAML and JSON formats are supported. See `config.example.yaml` or `config.example.json` for complete examples.

**YAML format (config.yaml):**
```yaml
server:
  host: 0.0.0.0
  port: 8080
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 120s
  shutdown_timeout: 30s
  environment: development

database_type: sqlite

sqlite_config:
  type: sqlite
  filepath: ./data/smotra.db
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
  conn_max_idle_time: 10m

logging:
  level: info
  format: json

auth:
  jwt_secret: ""
  jwt_expiration: 24h
```

**JSON format (config.json):**
```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "read_timeout": "15s",
    "write_timeout": "15s",
    "idle_timeout": "120s",
    "shutdown_timeout": "30s",
    "environment": "development"
  },
  "database_type": "sqlite",
  "sqlite_config": {
    "type": "sqlite",
    "filepath": "./data/smotra.db",
    "max_open_conns": 25,
    "max_idle_conns": 5,
    "conn_max_lifetime": "5m",
    "conn_max_idle_time": "10m"
  },
  "logging": {
    "level": "info",
    "format": "json"
  },
  "auth": {
    "jwt_secret": "",
    "jwt_expiration": "24h"
  }
}
```

### Configuration Fields

#### Server
- `host`: HTTP server host (default: 0.0.0.0)
- `port`: HTTP server port (default: 8080)
- `read_timeout`: Read timeout duration (default: 15s)
- `write_timeout`: Write timeout duration (default: 15s)
- `idle_timeout`: Idle timeout duration (default: 120s)
- `shutdown_timeout`: Graceful shutdown timeout (default: 30s)
- `environment`: development, staging, or production

#### Database
- `database_type`: sqlite or postgres

**For SQLite:**
- `filepath`: Database file path (default: ./data/smotra.db)
- `max_open_conns`: Maximum open connections (default: 25)
- `max_idle_conns`: Maximum idle connections (default: 5)
- `conn_max_lifetime`: Connection max lifetime (default: 5m)
- `conn_max_idle_time`: Connection max idle time (default: 10m)

**For PostgreSQL:**
- `host`: Database host
- `port`: Database port (default: 5432)
- `username`: Database username
- `password`: Database password
- `database`: Database name
- `sslmode`: SSL mode (disable, require, verify-full)
- `max_open_conns`: Maximum open connections (default: 25)
- `max_idle_conns`: Maximum idle connections (default: 5)
- `conn_max_lifetime`: Connection max lifetime (default: 5m)
- `conn_max_idle_time`: Connection max idle time (default: 10m)

#### Logging
- `level`: debug, info, warn, or error
- `format`: json or text

#### Authentication
- `jwt_secret`: JWT signing secret (required in production)
- `jwt_expiration`: JWT token expiration duration (default: 24h)

## Building for Production

```bash
# Build the binary
go build -o bin/server cmd/server/main.go

# Run the binary with config file
CONFIG_FILE=config.yaml ./bin/server
```

## Docker Support (Coming Soon)

Docker support with docker-compose configurations will be added soon.

## Project Structure

```
server/
├── cmd/
│   └── server/          # Main application entry point
│       └── main.go
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database interface and implementations
│   ├── handlers/        # HTTP handlers
│   │   └── health/      # Health check handlers
│   ├── logger/          # Logging setup
│   └── middleware/      # HTTP middleware
├── pkg/
│   └── api/             # Generated API code (from OpenAPI spec)
└── oapi-codegen/        # OpenAPI specification
    ├── config.yaml
    └── spec.yaml
```

## Development

### Adding New Features

1. Define API endpoints in `oapi-codegen/spec.yaml`
2. Regenerate API code: `make generate` (or run oapi-codegen manually)
3. Implement handlers in `internal/handlers/`
4. Register routes in `cmd/server/main.go`

### Running Tests

```bash
go test ./...
```

### Code Formatting

```bash
go fmt ./...
```

## Troubleshooting

### Database Connection Issues

**SQLite:**
- Ensure the directory for the database file exists or the application has write permissions
- Default location: `./data/smotra.db`

**PostgreSQL:**
- Verify PostgreSQL is running: `pg_isready -h localhost -p 5432`
- Check credentials and database exists
- Verify network connectivity and firewall rules

### Port Already in Use

If port 8080 is already in use, change it in your `config.yaml`:

```yaml
server:
  port: 8081
```

### Missing Configuration File

If you see an error about `CONFIG_FILE` not being set:

```bash
CONFIG_FILE environment variable must be set
```

Make sure to specify the config file path:

```bash
CONFIG_FILE=config.yaml ./bin/server
```

## Next Steps

- Review the full API documentation in `oapi-codegen/spec.yaml`
- Set up database migrations (coming soon)
- Configure OAuth2 authentication
- Deploy using Docker/Kubernetes

## Support

For issues and questions:
- GitHub Issues: https://github.com/smotra-monitoring/server/issues
- Documentation: https://docs.smotra.net (coming soon)
