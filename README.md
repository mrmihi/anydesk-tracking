# AnyDesk Tracker

A Windows service that monitors AnyDesk remote desktop sessions and sends real-time notifications to Slack when users connect or disconnect. Additionally tracks changes to specified configuration files.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## 🚀 Features

- **Real-time Login/Logout Tracking**: Monitors AnyDesk trace files for incoming session requests and session closures
- **Slack Notifications**: Sends formatted notifications to Slack with user details, timestamps, and session status
- **File Change Monitoring**: Watches external configuration files and reports changes with detailed diffs
- **User Attribution**: Associates file changes with the most recent AnyDesk user
- **Windows Service**: Runs as a background service with automatic startup
- **Configurable**: YAML-based configuration for easy customization
- **Persistent State**: Maintains user session state across service restarts

## 📋 Prerequisites

- **Windows OS**: Designed for Windows environments
- **Go 1.21+**: Required for building from source
- **AnyDesk**: Must be installed and running
- **Slack Webhook**: Incoming webhook URL for your Slack workspace

## 🔧 Installation

### 1. Clone the Repository

```bash
git clone https://github.com/mrmihi/any-desk-tracking.git
cd any-desk-tracking
```

### 2. Configure the Application

Copy the example configuration file and edit it with your settings:

```bash
copy config.example.yaml config.yaml
```

Edit `config.yaml` and set:
- `webhook_url`: Your Slack incoming webhook URL
- `vm_name`: A friendly name for this machine (appears in notifications)
- Adjust file paths if your AnyDesk installation differs from defaults

### 3. Build the Application

```bash
go build -o anydesk-tracker.exe
```

### 4. Install as Windows Service

Run as Administrator:

```cmd
anydesk-tracker.exe install
```

### 5. Start the Service

```cmd
anydesk-tracker.exe start
```

## ⚙️ Configuration

All configuration is done via `config.yaml`. See [`config.example.yaml`](config.example.yaml) for a fully documented example.

### Key Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `webhook_url` | Slack incoming webhook URL | **(Required)** |
| `vm_name` | Display name for this machine | **(Required)** |
| `user_trace_file` | Path to AnyDesk user trace file | `C:\Users\...\ad.trace` |
| `service_trace_file` | Path to AnyDesk service trace file | `C:\ProgramData\AnyDesk\ad_svc.trace` |
| `app_log_file` | Application log file path | `AnyDeskGoTracker.log` |
| `external_file` | Optional file to monitor for changes | `""` (disabled) |
| `allowed_recent_duration` | Time window for "recent" events | `5s` |
| `log_time_layout` | Timestamp format in AnyDesk logs | `2006-01-02 15:04:05.000` |

### Getting a Slack Webhook URL

1. Go to [https://api.slack.com/messaging/webhooks](https://api.slack.com/messaging/webhooks)
2. Create a new Incoming Webhook for your workspace
3. Select the channel where notifications should be posted
4. Copy the webhook URL to your `config.yaml`

## 🎯 Usage

### Service Commands

```cmd
# Install the service (run as Administrator)
anydesk-tracker.exe install

# Start the service
anydesk-tracker.exe start

# Stop the service
anydesk-tracker.exe stop

# Uninstall the service
anydesk-tracker.exe uninstall

# Run in foreground (for testing)
anydesk-tracker.exe run

# Show version information
anydesk-tracker.exe version
```

### Viewing Logs

Logs are written to the file specified in `app_log_file` (default: `AnyDeskGoTracker.log` in the same directory as the executable).

```cmd
type AnyDeskGoTracker.log
```

## 🔒 Security Best Practices

> [!IMPORTANT]
> **Protect Your Webhook URL**: The Slack webhook URL in `config.yaml` is sensitive. Anyone with this URL can send messages to your Slack channel.

- **Never commit `config.yaml`** to version control (it's in `.gitignore` by default)
- **Restrict file permissions**: Ensure `config.yaml` is only readable by administrators
- **Rotate webhook URLs**: If a webhook URL is exposed, regenerate it in Slack
- **Monitor logs**: Regularly check logs for unexpected activity
- **Use HTTPS**: Webhook URLs use HTTPS by default - never modify them to use HTTP


## 🏗️ Architecture

The application consists of several key components:

- **Monitor** (`monitor.go`): Tails AnyDesk trace files and parses log entries for login/logout events
- **File Watcher** (`filewatch.go`): Monitors external files for changes using filesystem notifications
- **User Tracker** (`usertracker.go`): Maintains state of the most recent AnyDesk user
- **Slack Integration** (`slack.go`): Sends formatted messages to Slack webhooks
- **Service Manager** (`service.go`): Manages Windows service lifecycle
- **Configuration** (`config.go`): Loads and validates YAML configuration

### How It Works

1. **Login Detection**: Monitors `ad.trace` for "Incoming session request" patterns
2. **Logout Detection**: Monitors `ad_svc.trace` for "Session closed by" patterns
3. **File Monitoring**: Watches configured files for write/create events
4. **Change Attribution**: Associates file changes with the last logged-in AnyDesk user
5. **Notifications**: Sends formatted Slack messages with timestamps and user details

## 🐛 Troubleshooting

### Service won't start

- **Check permissions**: Ensure you're running as Administrator
- **Verify AnyDesk paths**: Confirm trace file paths in `config.yaml` match your installation
- **Check webhook URL**: Ensure the Slack webhook URL is valid and accessible

### Not receiving Slack notifications

- **Test webhook**: Use `curl` or Postman to send a test message to your webhook URL
- **Check firewall**: Ensure outbound HTTPS (port 443) is allowed
- **Review logs**: Check `AnyDeskGoTracker.log` for error messages
- **Verify time window**: Events older than `allowed_recent_duration` are logged but not sent to Slack

### File monitoring not working

- **Check file path**: Ensure `external_file` path in `config.yaml` is correct
- **File permissions**: Verify the service has read access to the monitored file
- **File format**: File monitoring expects YAML format for diff generation

### Old events being reported

The service only sends notifications for "recent" events (within `allowed_recent_duration` of the current time). This prevents notification spam when the service starts or when reading old log entries.

### Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/any-desk-tracking.git

# Install dependencies
go mod download

# Run tests (if available)
go test ./...

# Build
go build
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [kardianos/service](https://github.com/kardianos/service) - Cross-platform service management
- [fsnotify/fsnotify](https://github.com/fsnotify/fsnotify) - File system notifications
- [nxadm/tail](https://github.com/nxadm/tail) - Log file tailing
- [goccy/go-yaml](https://github.com/goccy/go-yaml) - YAML parsing

**Version**: 2.1.0 | **Built**: 2025-12-02
