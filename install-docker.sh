#!/bin/bash

APP_NAME="fleare"
APP_BINARY_NAME="fleare-1-0-1-linux-amd64"
APP_BIN="/usr/local/bin/$APP_NAME"
CONFIG_DIR="/etc/$APP_NAME"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
LOG_DIR="/usr/local/$APP_NAME/log"
DATA_DIR="/usr/local/$APP_NAME/lib"
BACKUP_DIR="/usr/local/$APP_NAME/backups"

# Create configuration directory and file
echo "Creating configuration directory at $CONFIG_DIR..."
mkdir -p "$CONFIG_DIR"

# Create directories for logs, data, and backups
echo "Creating necessary directories..."
mkdir -p "$LOG_DIR" "$DATA_DIR" "$BACKUP_DIR"

# Coping the compiled file to the appropriate location
echo "Coping the compiled file to $APP_BIN..."
cp $APP_BINARY_NAME $APP_BIN
chmod +x $APP_BIN

tee "$CONFIG_FILE" > /dev/null <<EOL
# Configuration for My Fleare Database

# Server settings
server:
  host: "0.0.0.0" # Listen on all network interfaces
  port: 4775 # Port number for the database server

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
    - username: "admin"
      password: "admin123" # In a real-world scenario, use hashed passwords!
      role: "root"

# Data persistence
persistence:
  enable: true # Enable data persistence
  path: "$DATA_DIR" # Path to store data
  after_write_count: 100 # save data on Path after

# Backup settings
backup:
  enable: true # Enable automated backups
  interval_minutes: 60 # Backup interval in minutes
  backup_dir: "$BACKUP_DIR"

# Other settings
misc:
  max_connections: 100 # Maximum number of client connections
  timeout_seconds: 30 # Timeout for client requests
  strict_insert: true
EOL

# Set permissions for the config file
echo "Setting permissions for $CONFIG_FILE..."
# chown root:root $CONFIG_FILE
chmod 644 $CONFIG_FILE

echo "Setup complete!"
