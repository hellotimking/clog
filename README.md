# CLOG - High-Visibility Caddy Logs 🪵

## Stop squinting at JSON. Start monitoring at the speed of Go.

**clog** is a high-performance, real-time terminal dashboard designed specifically for Caddy's JSON access logs. It transforms messy, hard-to-read JSON streams into a clean, actionable visual interface.

## 🚀 Installation

### From Source
Clone the repository

git clone https://github.com/hellotimking/clog.git
cd clog

Build the binary

go build -ldflags="-s -w" -o clog

Move to your path (optional)

sudo mv clog /usr/local/bin/
