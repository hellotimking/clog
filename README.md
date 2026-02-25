# CLOG v0.6.3 - High-Visibility Caddy Logs 🪵
## Stop squinting at JSON. Start monitoring at the speed of Go.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://github.com/hellotimking/clog)
[![License](https://img.shields.io/github/license/hellotimking/clog)](LICENSE)

**clog** is a specialized log processor and visualizer built in Go. It solves user JSON-squinting by transforming Caddy's structured logs into an interactive, human-centric dashboard. Designed for systems administrators and developers who need instant situational awareness without the overhead of heavy logging stacks.

Transforming messy, hard-to-read JSON streams into a clean, actionable visual interface.

---

## 📋 Table of Contents
- [Features](#-features)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Command Line Interface](#-command-line-interface)
- [Advanced Usage](#-advanced-usage)
- [Performance](#-performance)
- [License](#-license)

---

## ✨ Features

* **⚡ Zero-Latency Streaming:** Uses non-blocking I/O and optimized Go channels to handle high-traffic environments without dropping frames.
* **📊 Real-time Analytics:** Instant status code distribution (2xx, 3xx, 4xx, 5xx) visualized in the TUI.
* **🔍 Power Filtering:** Regex-based or field-specific filtering to isolate problematic endpoints or specific status codes.
* **🧠 Schema Aware:** Deep understanding of Caddy's default JSON log structure—no configuration required.
* **🎨 Responsive TUI:** Built with a terminal UI that scales from small side-panes to full-screen NOC displays.

---

## 🚀 Installation

### From Source
Requires Go 1.21 or higher.

```bash
# Clone the repository
git clone https://github.com/hellotimking/clog.git
cd clog
```

#### Build optimized binary
```
go build -ldflags="-s -w" -o clog
```
#### Global install
```
sudo mv clog /usr/local/bin/
```

=========
# CLOG - High-Visibility Caddy Logs 🪵

## Stop squinting at JSON. Start monitoring at the speed of Go.

**clog** is a high-performance, real-time terminal dashboard designed specifically for Caddy's JSON access logs. It transforms messy, hard-to-read JSON streams into a clean, actionable visual interface.

## 📋 Core Features

- **Native JSON Awareness:** Built to understand the Caddy log schema natively. No external dependencies required.
- **Live Metrics:** The dashboard tracks status code distribution (2xx/3xx/4xx/5xx) so you can spot spikes in errors instantly.
- **Zero-Lag Stream:** Optimized Go routines ensure that even under high-traffic loads (10k+ requests/sec), your monitoring tool doesn't become the bottleneck.
- **Human-Centric Design:** Focused on the four golden signals: Method, Path

## 🚀 Installation

### From Source
#### Clone the repository
```
git clone https://github.com/hellotimking/clog.git
cd clog
```

#### Build the binary
```
go build -ldflags="-s -w" -o clog
```
#### Move to your path (optional)
```
sudo mv clog /usr/local/bin/
```

## 🛠 Usage

### Command Line Switches

| Switch | Long Flag | Description | 
 | ----- | ----- | ----- | 
| `-d` | `--dashboard` | Enables the TUI Dashboard mode (Status distribution & metrics). | 
| `-s` | `--strip` | Strips the module prefix (e.g., `http.log.access`) from the output. | 
| `-f` | `--filter` | Filter logs by a specific field or value (e.g., `-f "status:500"`). | 
| `-m` | `--max` | Limit the number of lines displayed in the dashboard view. | 
| `-c` | `--color` | Toggle color output (default: on). | 
| `-v` | `--version` | Displays the current version of clog. | 
| `-h` | `--help` | Shows the help menu and usage instructions. | 

### Examples

**Basic Tail**
Simply point **clog** at your Caddy access log to see a cleaned-up, human-readable stream:
