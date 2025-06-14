Here is an **Installation User Manual** document for the provided Linux ARM64 script for setting up the `Fleare` application.

---

# 🛠️ Fleare Installation Guide (Linux ARM64)

## 📋 Overview

This guide provides step-by-step instructions to install and run the `Fleare` application on a **Linux ARM64** system using the included installation script. It configures binaries, sets up a configuration file, and creates a `systemd` service to manage the application.

---

## ✅ Prerequisites

1. **Linux ARM64** system
2. **Root access** (or `sudo` privileges)
3. `fleare` **compiled binary** placed in the same directory as the installation script
4. `bash` shell

---

## 📦 Installation Steps

### 1. Place the Binary

Ensure the compiled binary of `Fleare` is named `fleare` and is in the same directory as the install script:

```bash
ls
> install.sh fleare
```

---

### 2. Make Script Executable

```bash
chmod +x install.sh
```

---

### 3. Run the Script

```bash
sudo ./install.sh
```

The script will:

* Create necessary directories:

  * Config: `/etc/fleare`
  * Logs: `/usr/local/fleare/log`
  * Data: `/usr/local/fleare/lib`
  * Backups: `/usr/local/fleare/backups`
* Copy binary to: `/usr/local/bin/fleare`
* Generate config at: `/etc/fleare/config.yaml`
* Create a `systemd` service at: `/etc/systemd/system/fleare.service`
* Enable and start the service

---

## 📂 File Structure After Installation

| File/Directory                       | Description                |
| ------------------------------------ | -------------------------- |
| `/usr/local/bin/fleare`              | Application binary         |
| `/etc/fleare/config.yaml`            | Configuration file         |
| `/usr/local/fleare/log/`             | Log directory              |
| `/usr/local/fleare/lib/`             | Data persistence directory |
| `/usr/local/fleare/backups/`         | Automated backup storage   |
| `/etc/systemd/system/fleare.service` | Systemd service definition |

---

## 🔐 What the Script Does

1. **Validates Linux OS**.
2. **Creates necessary directories** for config, data, logs, and backups.
3. **Copies the binary** to `/usr/local/bin/fleare` and makes it executable.
4. **Generates a default configuration** at `/etc/fleare/config.yaml`.
5. **Creates and registers a `systemd` service** for auto-start on boot.
6. **Starts and enables** the `fleare` service.

---

## 🔧 Configuration

A default config is created at `/etc/fleare/config.yaml` with the following sections:

⚠️ **Change the default password** before going to production.

---

## 🖥️ Managing the Service

| Command                         | Description          |
| ------------------------------- | -------------------- |
| `sudo systemctl start fleare`   | Start the service    |
| `sudo systemctl stop fleare`    | Stop the service     |
| `sudo systemctl restart fleare` | Restart the service  |
| `sudo systemctl status fleare`  | Check service status |
| `sudo journalctl -u fleare -f`  | View live logs       |

---

## 🧼 Uninstallation (Optional)

```bash
sudo systemctl stop fleare
sudo systemctl disable fleare
sudo rm /etc/systemd/system/fleare.service
sudo rm /usr/local/bin/fleare
sudo rm -rf /usr/local/fleare
sudo rm -rf /etc/fleare
sudo systemctl daemon-reload
```

---

## 📞 Support

For issues or enhancements, contact the Fleare development team or visit the official repository (if applicable).
