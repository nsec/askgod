# Askgod

Askgod is the NorthSec CTF backend that handles flags.

## Installation

### 1. Install go 1.24.1

https://go.dev/doc/install 

```bash
go --version  # Should output `go version go1.24.1 linux/amd64`
go install github.com/go-delve/delve/cmd/dlv@latest  # Install delve (go debugger)
```

### 2. Install recommended VS Code extensions

You should have a pop up if you opened the root directory in VS Code.


### 3. Start a postgresql instance

```bash
docker compose up -d
```

### 4. Launch askgod-server in debug using VS Code

Press `F5`.

### 5. Add seed data

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