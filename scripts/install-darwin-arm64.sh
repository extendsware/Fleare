#!/bin/bash

APP_NAME="fleare"
APP_BINARY_NAME="fleare"
APP_BIN="/usr/local/bin/$APP_NAME"
CONFIG_DIR="/etc/$APP_NAME"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
LOG_DIR="/usr/local/$APP_NAME/log"
DATA_DIR="/usr/local/$APP_NAME/db"
BACKUP_DIR="/usr/local/$APP_NAME/backups"

# Check if script is running on macOS
if [[ "$(uname)" != "Darwin" ]]; then
  echo "This script is intended for macOS only." >&2
  exit 1
fi

# Create application directory if it doesn't exist
echo "Creating application directory at $APP_BIN..."
# sudo mkdir -p "$APP_BIN"

# Create configuration directory and file
echo "Creating configuration directory at $CONFIG_DIR..."
sudo mkdir -p "$CONFIG_DIR"

# Create directories for logs, data, and backups
echo "Creating necessary directories..."
sudo mkdir -p "$LOG_DIR" "$DATA_DIR" "$BACKUP_DIR"

# Coping the compiled file to the appropriate location
echo "Coping the compiled file to $APP_BIN..."
sudo cp $APP_BINARY_NAME $APP_BIN
sudo chmod +x $APP_BIN

# Write configuration file
sudo tee "$CONFIG_FILE" > /dev/null <<EOL
# Configuration for My Infleare Database

# Server settings
server:
  host: "127.0.0.1" # Listen on all network interfaces
  port: 9219 # Port number for the database server

# Logging settings
logging:
  level: "info" # Logging level: debug, info, warn, error
  file: "$LOG_DIR/fleareDB.log" # Log file path

# Memory settings
memory:
  max_size_mb: 1024 # Maximum memory usage in MB
  eviction_policy: "LRU" # Eviction policy: LRU, LFU, FIFO

# Security settings
security:
  enable_auth: true # Enable authentication
  auth_method: "basic" # Authentication method: basic, token, etc.
  users:
    - username: "root"
      password: "root" # In a real-world scenario, use hashed passwords!
      role: "root"

# Data persistence
persistence:
  enable: true # Enable data persistence
  path: "$DATA_DIR" # Path to store data
  after_write_count: 100 # save data on Path after

shard:
  mode: "local" # Path to store data
  shard_count: 3 # save data on Path after

# Backup settings
backup:
  enable: true # Enable automated backups
  interval_minutes: 60 # Backup interval in minutes
  backup_dir: "$BACKUP_DIR"

# Other settings
misc:
  max_connections: 200 # Maximum number of client connections
  timeout_seconds: 30 # Timeout for client requests
  strict_insert: true
EOL

echo "Setup complete! Configuration file is located at $CONFIG_FILE"
