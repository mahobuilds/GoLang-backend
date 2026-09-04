# IoT Device Telemetry Service

A simple in-memory REST API built in Go, using only the standard library (`net/http`), for collecting and querying sensor device readings.

## Purpose

This project was built to test various Go topics, including:

1. Structures
2. Filtering
3. HTTP Handlers
4. Concurrency
5. Error handling

The task was to build a backend service that keeps track of sensor devices and their readings.

## Functionality

The program uses built-in, in-memory storage (no external database). It exposes HTTP handlers responsible for parsing incoming client requests, processing the data, and returning an appropriate response. Each handler contains simple validation logic to ensure incoming data is well-formed before it's stored.

## How to Run

```bash
go run .
```

The server starts and listens on port `8080`.

## Project Structure

For scalability reasons, the project has been seperated into multiple files. Including handlers.go, modles.go, stats.go, and store.go.

### 1. main.go

This is our main file. It is responsible for the starting of the server and simple routing.

### 2. handlers.go

This file is reposnible for the handlers.

### 3. modles.go

This file is responsible for the data structres.

### 4. stats.go

This file contains the function to calculate the status of a device's readings.

### 5. store.go

This file contains the Store structre, which has all current data structures connected under it for connectivity purposes

## Design Decisions

### Empty readings on stats

If a device exists but has no readings yet, `GET /devices/{id}/stats` returns a `404` rather than `{"min":0,"max":0,"avg":0}`. Returning zeros would be confusing, a device whose readings genuinely average to zero would produce an identical response to a device with no data at all, giving the client no way to distinguish the two cases. Failing explicitly with a 404 removes that ambiguity.

### Storage: separate maps for devices and readings

Devices are stored as `map[string]Device`, and readings are stored separately as `map[string][]Reading`, both keyed by device ID. Readings are not embedded directly inside `Device` because a device can accumulate many readings over time, a single embedded field couldn't represent that, and Go doesn't allow assigning directly to a struct field nested inside a map value. Keying both maps by the same device ID keeps the relationship simple while avoiding that limitation.

### Timestamp representation

`Reading.Timestamp` is stored as a `float64` (Unix timestamp), rather than `time.Time` or an RFC3339 string. This was chosen for simplicity, it keeps encoding, decoding, and range comparisons for the `?from=&to=` filter straightforward, at the cost of being less human-readable than a formatted date string.

### Concurrency: RWMutex over Mutex

The service uses a single shared `sync.RWMutex`, declared once and shared across all handlers. `RLock()`/`RUnlock()` are used for read-only handlers (`GET /devices/{id}/readings`, `GET /devices/{id}/stats`), since multiple goroutines can safely read shared data at the same time. `Lock()`/`Unlock()` are used for handlers that modify data (`POST /devices`, `POST /devices/{id}/readings`), since writes must be exclusive to prevent concurrent map access, which can corrupt data or crash the program outright. Since reads are expected to significantly outnumber writes in this service, `RWMutex` allows better concurrent throughput than a plain `Mutex`, which would serialize all access, reads included.

### Status codes

- **400 Bad Request** malformed JSON, or a value that fails to parse (e.g. an invalid `from`/`to` query parameter)
- **404 Not Found** a requested device or its readings don't exist, or a device exists but has no readings for `/stats`
- **409 Conflict** attempting to create a device with an ID that already exists

## Dependencies

Only Go's standard library is used, no third-party frameworks:

- `encoding/json` decoding incoming requests and encoding outgoing responses
- `fmt` writing response/error messages
- `net/http` routing and handling HTTP requests
- `sync` read/write locking for concurrency safety (see above)
- `strconv` converting query parameter strings to `float64`
- `math` providing `math.MaxFloat64` as the default upper bound when no `to` parameter is given
