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

## Q&A

### How long did you spend working on the problem?

Around 2 hours.

### What did you find to be the most difficult part?

The most difficult part was working with Go itself, since I don't have
prior experience with it. That included researching syntax, package
layout, idiomatic best practices, and how core mechanisms
(goroutines/locks, maps, structs) work in Go.

### Discuss your solution's runtime complexity

The runtime complexity is O(n) per request. I chose a map for the
in-memory store so device lookups stay O(1), and I avoided nested loops
in the aggregation paths. I don't currently see meaningful algorithmic
bottlenecks; the next realistic problem is unbounded in-memory growth,
not running time.

### How would you modify your data model or code to account for more kinds of metrics?

Right now we have a simple data structure since I only have 2 metrics and
didn't want to complicate things. This solution works if we have a small
set of metrics and aren't foreseeing any additions. However, if the number
of metrics is large, or if we'll be updating them often (for both POST and
GET), I would want to make it more scalable. Adding a new metric currently
requires modifying 5 files, which isn't the most extensible solution.

A better approach would be to have a registry of metrics with a validation
schema per metric, unify the POST endpoint and the DB method, and use a
schemaless database so the storage layer doesn't need to know each
metric's shape. When a request comes in, we'd look up the
metric name in the registry, validate the body shape against the
registered schema, and push the document to the DB. Adding a new metric
would become a one-file change. The only real cost is maintaining the
schemas as the metric count grows — but those are the API contract, so
we'd want them documented either way.

A similar idea applies to the GET endpoint: each registered metric could
also carry an aggregation function, and the existing `/stats` endpoint
would walk the registry to build the response. We may also want to expose
a per-metric endpoint, or accept a query parameter on `/stats` to specify
which metrics to include.


