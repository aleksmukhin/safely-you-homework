# Fleet Management Simple Metrics Server

A small Go HTTP service implementing the Fleet Management Metrics coding
assessment. Devices post heartbeats and per-upload stats; the server returns
per-device uptime and average upload time on demand.

The full HTTP contract lives in [openapi.json](openapi.json).

## Getting started

These steps assume you've never touched this project (or Go) before.

### 1. Install Go

This project targets **Go 1.26+** (declared in [go.mod](go.mod)). On macOS:

```sh
brew install go
```

Verify the install:

```sh
go version
# go version go1.26.x ...
```

### 2. Install dependencies

From the repository root:

```sh
go mod download
```

This pulls every module listed in `go.mod` (Gin, validators, etc.) into your
local module cache. You only need to run this once per checkout.

### 3. Run the server

```sh
go run main.go
```

The server listens on `127.0.0.1:6733` with all routes mounted under
`/api/v1`. Devices are bootstrapped from [devices.csv](devices.csv) at
startup. Press `Ctrl-C` to stop.

Quick smoke test (in another terminal):

```sh
curl -i http://127.0.0.1:6733/api/v1/devices/60-6b-44-84-dc-64/stats
# HTTP/1.1 200 OK
# {"uptime":0,"avg_upload_time":"0s"}
```

### 4. Run the device simulator

The simulator drives traffic against the running server, then compares
the returned stats to expected values. From the repo root, while the
server is running:

```sh
./device-simulator-mac-arm64
```

Output streams to your terminal; sample output is preserved in
[results.txt](results.txt).

## Architecture

The project is split into three layers, each in its own package:

```
main.go               # entry point: gin engine, CSV bootstrap, route wiring
handlers/             # HTTP layer (request/response, OpenAPI conformance)
  device_handler.go
adapters/             # storage layer (in-memory DB)
  db.go
services/             # placeholder
```

The HTTP contract is defined in [openapi.json](openapi.json); responses
in `handlers/` follow it.
