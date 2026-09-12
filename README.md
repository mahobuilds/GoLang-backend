# IoT Device Telemetry Service

A simple in-memory REST API built in Go, using various relative libraries, for collecting and querying sensor device readings.

## Purpose

This project was built to test various Go topics, including:

1. Structures
2. Filtering
3. HTTP Handlers
4. Concurrency
5. Error handling
6. Context and request lifetime
7. Graceful shutdown

The task was to build a backend service that keeps track of sensor devices and their readings.

## Functionality

The program uses built-in, in-memory storage (no external database). It exposes HTTP handlers responsible for parsing incoming client requests, processing the data, and returning an appropriate response. Each handler contains simple validation logic to ensure incoming data is well-formed before it's stored.

## How to Run

```bash
go run .
```

The server starts and listens on port `8080`. Press `Ctrl+C` (or send `SIGTERM`) to stop it. The server finishes any in-flight requests before exiting (see Graceful shutdown below).

## Project Structure

For scalability reasons, the project has been seperated into multiple files. Including handlers.go, modles.go, stats.go, and store.go.

### 1. main.go

This is our main file. It is responsible for the starting of the server and simple routing. Routes are registered on a dedicated `http.NewServeMux()` instead of the default global mux, and the mux is passed to an `http.Server` so the server can be shut down gracefully.

### 2. handlers.go

This file is reposnible for the handlers.

### 3. models.go

This file is responsible for the data structres.

### 4. stats.go

This file contains the function to calculate the status of a device's readings.

### 5. store.go

This file contains the Store structre, which has all current data structures connected under it for connectivity purposes. `GetDevice` now accepts a `context.Context` and checks whether it was cancelled before taking the lock.

### 6. errors.go

Contains sentinel error values for better error checking. It also contains the `ErrorResponse` struct and the `writeError` helper, which every handler uses to return errors in the same JSON shape.

## Design Decisions

### Empty readings on stats

If a device exists but has no readings yet, `GET /devices/{id}/stats` returns a `404` rather than `{"min":0,"max":0,"avg":0}`. Returning zeros would be confusing, a device whose readings genuinely average to zero would produce an identical response to a device with no data at all, giving the client no way to distinguish the two cases. Failing explicitly with a 404 removes that ambiguity.

### Storage: separate maps for devices and readings

Devices are stored as `map[string]Device`, and readings are stored separately as `map[string][]Reading`, both keyed by device ID. Readings are not embedded directly inside `Device` because a device can accumulate many readings over time, a single embedded field couldn't represent that, and Go doesn't allow assigning directly to a struct field nested inside a map value. Keying both maps by the same device ID keeps the relationship simple while avoiding that limitation.

### Timestamp representation

`Reading.Timestamp` is stored as a `float64` (Unix timestamp), rather than `time.Time` or an RFC3339 string. This was chosen for simplicity, it keeps encoding, decoding, and range comparisons for the `?from=&to=` filter straightforward, at the cost of being less human-readable than a formatted date string.

### Concurrency: RWMutex over Mutex

The service uses a single shared `sync.RWMutex`, declared once and shared across all handlers. `RLock()`/`RUnlock()` are used for read-only handlers (`GET /devices/{id}/readings`, `GET /devices/{id}/stats`), since multiple goroutines can safely read shared data at the same time. `Lock()`/`Unlock()` are used for handlers that modify data (`POST /devices`, `POST /devices/{id}/readings`), since writes must be exclusive to prevent concurrent map access, which can corrupt data or crash the program outright. Since reads are expected to significantly outnumber writes in this service, `RWMutex` allows better concurrent throughput than a plain `Mutex`, which would serialize all access, reads included.

### Consistent error responses

Previously each handler wrote its own plain-text error message with `fmt.Fprintln`, so the format differed from one endpoint to the next and some messages leaked the raw decoding error to the client. All handlers now call a single `writeError(w, status, message)` helper in `errors.go`. It sets `Content-Type: application/json`, writes the status code, and encodes an `ErrorResponse`:

```json
{ "error": "device with ID: abc does not exist" }
```

This gives clients one predictable shape to parse regardless of which endpoint failed. Internal error details (such as JSON decode errors) are no longer sent back, the client only gets a short generic message like `bad request`. As part of this cleanup, `GET /devices/{id}/stats` now returns `404` instead of `409` when the device does not exist, and `getDeviceData` was renamed to `getDevice`.

### Context and request lifetime

`Store.GetDevice` now takes a `context.Context` as its first argument, and the `GET /devices/{id}` handler passes `r.Context()` into it. Before acquiring the read lock, the store does a non-blocking `select` on `ctx.Done()` and returns `ctx.Err()` if the context is already cancelled. This means that if the client disconnects or the request is cancelled, the store can bail out early instead of doing work whose result nobody will read. The `deviceGetter` interface was updated to match the new signature. Only `GetDevice` has been converted so far, the other store methods still use their original signatures.

### Graceful shutdown

The server no longer calls `http.ListenAndServe(":8080", nil)` directly. Instead, `main.go`:

1. Builds an `http.Server` with the mux as its handler.
2. Creates a context with `signal.NotifyContext` that is cancelled on `os.Interrupt` (`Ctrl+C`) or `syscall.SIGTERM`.
3. Starts `server.ListenAndServe()` in a goroutine. `http.ErrServerClosed` is ignored because it is the expected error after a shutdown.
4. Blocks on `<-signalCtx.Done()` until a signal arrives.
5. Calls `server.Shutdown` with a 10 second timeout context. `Shutdown` stops accepting new connections and waits for in-flight requests to complete. If they do not finish within 10 seconds, the process exits with `log.Fatal`.

Without this, stopping the process would cut off any request that was still being handled.

### Status codes

- **400 Bad Request** malformed JSON, a mismatched `Content-Type`, or a value that fails to parse (e.g. an invalid `from`/`to` query parameter)
- **404 Not Found** a requested device or its readings don't exist, or a device exists but has no readings for `/stats`
- **409 Conflict** attempting to create a device with an ID that already exists
- **500 Server Error** an unknown error that is yet to be defined by the system

All error responses use the JSON `{"error": "..."}` format described above.

## Dependencies

- `encoding/json` decoding incoming requests and encoding outgoing responses
- `fmt` writing response/error messages
- `net/http` routing and handling HTTP requests
- `sync` read/write locking for concurrency safety (see above)
- `strconv` converting query parameter strings to `float64`
- `math` providing `math.MaxFloat64` as the default upper bound when no `to` parameter is given
- `errors` used to classify errors and give them more context
- `context` carrying request cancellation into the store and setting the shutdown timeout
- `os`, `os/signal`, `syscall` listening for `Ctrl+C` and `SIGTERM` to trigger shutdown
- `time` defining the 10 second shutdown deadline
- `log` reporting fatal server errors on startup or forced shutdown
