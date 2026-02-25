# CLOG - High-Visibility Caddy Logs 🪵

## Stop squinting at JSON. Start monitoring at the speed of Go.

**clog** is a high-performance, real-time terminal dashboard designed specifically for Caddy's JSON access logs. It transforms messy, hard-to-read JSON streams into a clean, actionable visual interface.

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
Move to your path (optional)

sudo mv clog /usr/local/bin/

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
