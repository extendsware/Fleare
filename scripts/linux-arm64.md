
---
### Fleare Installation Manual 
---

`(Linux - arm64)`

Here is a **user installation manual document** for the provided `fleare` Linux installation script (for `linux-arm64`):

---

### Overview

This document provides step-by-step instructions for installing and configuring the `Fleare` application on a **Linux (arm64)** system using the provided shell script.

---

### System Requirements

* **Operating System:** Linux (64-bit)
* **User Access:** `sudo` privileges
* **Dependencies:** None required beyond standard Linux utilities (e.g., `systemd`)

---

### Installation Steps

#### Download the installation file and extract zip file

 1. Download file 👉 [fleare-1-0-1-linux-arm64.tar.gz](https://www.fleare.com/download/binary/fleare-1-0-1-linux-arm64.tar.gz)
 2. Unzip the Installation zip file (`fleare-1-0-1-linux-arm64.tar.gz`)
 3. Ensure the script has execution permissions.
 4. Read `README.md` for installation guide.

---

### Installation Instructions

#### 1. **Prepare the Binary**

Ensure the compiled Fleare binary is named `fleare` and is in the same directory as the script.

#### 2. **Run the Installer Script permissions**

```bash 
chmod +x install.sh
```

#### 3. **Run the Installer Script**
```bash 
./install.sh
```

or

```bash 
sudo ./install.sh
```

> **Note:** Ensure you're running as a user with `sudo` privileges. You don't need to switch to `root`.

---

### What the Script Does

1. **Validates Linux OS**.
2. **Creates necessary directories** for config, data, logs, and backups.
3. **Copies the binary** to `/usr/local/bin/fleare` and makes it executable.
4. **Generates a default configuration** at `/etc/fleare/config.yaml`.
5. **Creates and registers a `systemd` service** for auto-start on boot.
6. **Starts and enables** the `fleare` service.

---

### Directory Structure

| Path                                 | Description                |
| ------------------------------------ | -------------------------- |
| `/usr/local/bin/fleare`              | Binary executable          |
| `/etc/fleare/config.yaml`            | Main configuration file    |
| `/usr/local/fleare/log/`             | Log directory              |
| `/usr/local/fleare/lib/`             | Data storage directory     |
| `/usr/local/fleare/backups/`         | Backup files directory     |
| `/etc/systemd/system/fleare.service` | Systemd service definition |

---


### Configuration

The default configuration is located at:

```
/etc/fleare/config.yaml
```

#### Example Snippet:

```yaml
server:
  host: "127.0.0.1"
  port: 9219

logging:
  level: "info"
  file: "/usr/local/fleare/log/fleareDB.log"

security:
  enable_auth: true
  users:
    - username: "root"
      password: "root"
```

You can manually edit this file as needed.

---

### Managing the Service

| Command                         | Description          |
| ------------------------------- | -------------------- |
| `sudo systemctl start fleare`   | Start the service    |
| `sudo systemctl stop fleare`    | Stop the service     |
| `sudo systemctl restart fleare` | Restart the service  |
| `sudo systemctl status fleare`  | Check service status |
| `sudo journalctl -u fleare -f`  | View live logs       |

---

### Backups

Automated backups will be created in:

```
/usr/local/fleare/backups/
```

By default, backups run every 60 minutes.

---

### Uninstallation

To remove Fleare:

```bash 
sudo systemctl stop fleare
sudo systemctl disable fleare
sudo rm /usr/local/bin/fleare
sudo rm -rf /usr/local/fleare
sudo rm /etc/fleare/config.yaml
sudo rm /etc/systemd/system/fleare.service
sudo systemctl daemon-reload
```

---

### Post-Installation Checklist

* [ ] Confirm binary is in `/usr/local/bin/`
* [ ] Confirm `fleare` service is running
* [ ] Test connection on port `9219`
* [ ] Secure configuration with hashed passwords

---

### Support

For assistance, please contact the Fleare development team or visit the repository documentation.
