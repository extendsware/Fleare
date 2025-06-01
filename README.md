# Fleare

Fleare is a high-performance, distributed data storage and caching system written in Go. It provides a flexible key-value store with support for complex data types and efficient data manipulation operations.

## Features

- **Multiple Data Type Support**:
  - String values
  - Map (JSON objects)
  - List (Arrays)
- **Configurable Logging Levels**
- **TCP Client Support**
- **Docker Support**
- **Cross-platform Installation Scripts**

## Installation

### Using Installation Scripts

Choose the appropriate installation script for your platform:

```bash
# For macOS ARM64
./install-darwin-arm64.sh

# For Linux AMD64
./install-linux-amd64.sh

# For Docker
./install-docker.sh
```

### Manual Installation

1. Ensure Go 1.23.0 or later is installed
2. Set up your Go environment:
   ```bash
   export PATH="$PATH:$(go env GOPATH)/bin"
   ```
3. Install protocol buffers (if needed):
   ```bash
   protoc --go_out=. --go_opt=paths=source_relative internal/comm/comm.proto
   ```

## Running the Server

You can run the server with different logging levels:

```bash
# Debug level (logs debug, info, warn, error)
go run main.go -loglevel=debug

# Info level (logs info, warn, error)
go run main.go -loglevel=info

# Warning level (logs warn, error)
go run main.go -loglevel=warn

# Error level (logs only errors)
go run main.go -loglevel=error
```

## Configuration

The server can be configured using YAML configuration files or environment variables. Configuration can be provided via:

- Command line flag: `--config`
- Environment variable: `config`
