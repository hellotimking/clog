# CLOG - High-Visibility Caddy Logs 🪵
## Stop squinting at JSON. Start monitoring at the speed of Go.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://github.com/hellotimking/clog)
[![License](https://img.shields.io/github/license/hellotimking/clog)](LICENSE)

**CLOG** is a specialized log processor and visualizer built in Go. It solves user JSON-squinting by transforming Caddy's structured logs into an interactive, human-centric dashboard. Designed for systems administrators and developers who need instant situational awareness without the overhead of heavy logging stacks.

**Transforming messy, hard-to-read JSON streams into a clean, actionable visual interface.**

---
- [Features](#-features)
- [Installation](#-installation)
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

🏎 Performance

CLOG is designed to be lightweight. It runs as a single static binary with:
* **Low CPU Overhead:** Log parsing happens in parallel worker pools using Go routines.
* **Predictable Memory:** Uses a fixed-size ring buffer for history to prevent memory leaks.
* **Efficiency:** Capable of processing thousands of lines per second with negligible latency.

---

## 🚀 Installation

### From Source
**Requires Go 1.21 or higher.**

#### Clone the repository
```bash
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
---

## 🚀 Command Line Interface

| Switch | Long Flag | Description |
| :---- | :---- | :---- |
| \-l | \--lines | Number of previous lines to show from the log file. |
| \-h | \--host | Only show logs for a specific domain or IP address. |
| \-f | \--find | Only show lines containing a specific string. |
| \-e | \--errors | Only show requests with status code \>= 400\. |
| \-ha | \--hide-assets | Hides common asset types (.js, .css, images, etc). |
| \-a | \--all | Show entire history and ignore asset filters. |
| \-s | \--status | Show system resource bar at the bottom of the terminal. |
| \-d | \--dashboard | Enable 1-second dashboard mode for real-time metrics. |
| \-c | \--clear-screen | Clear terminal before starting and on exit. |
| \ | \--help | Show the help menu and usage instructions. |



**Basic Tail**
Simply point **clog** at your Caddy access log to see a cleaned-up, human-readable stream:
