# Askgod

Askgod is the NorthSec CTF backend that handles flags.

## Installation

### 1. (optional) Install go

https://go.dev/doc/install 

```bash
go --version  # Should output `go version go<version> linux/amd64`
go install github.com/go-delve/delve/cmd/dlv@latest  # Install delve (go debugger)
```

### 2. (optional) Install recommended VS Code extensions

You should have a pop up if you opened the root directory in VS Code.


### 3. Build askgod-server and start a postgresql instance using docker compose:

```bash
docker compose up -d
```

### 4. Add seed data

```bash
./seed_data.sh
```


## --------

### Compile binary

From the root directory: 

```bash
make linux
```

This will create two executables in `./bin/linux`: `askgod` and `askgod-server`.

### Launch the askgod-server

```bash
./bin/linux/askgod-server ./askgod.yaml.example
```

## MCP Server

The askgod server supports an MCP server at `<askgod_server_address>/mcp`.
This MCP server allows users to submit flags.
The MCP Server is disabled by default, but can be enabled by setting `mcp: true` in the config.