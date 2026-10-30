# MicroTransit Go

A Go-based bus location tracking backend built around HTTP APIs,
PostgreSQL persistence, and a browser-based map interface. The project
is being developed in versions: V1 establishes the core tracking
pipeline; future versions are planned to add performance-focused C++
components and Python-based bus arrival-time prediction.

> **Project status:** V1 prototype / active development. The benchmark
> figures below are local test results, not production-capacity claims.

## Table of contents

-   [Overview](#overview)
-   [Current capabilities](#current-capabilities)
-   [Architecture](#architecture)
-   [Repository structure](#repository-structure)
-   [Technology stack](#technology-stack)
-   [Run locally](#run-locally)
-   [API](#api)
-   [Load-test notes](#load-test-notes)
-   [Security and configuration](#security-and-configuration)
-   [Roadmap](#roadmap)
-   [Known limitations and next
    steps](#known-limitations-and-next-steps)
-   [Contributing](#contributing)

## Overview

MicroTransit Go is a bus-location tracking project. A client submits a
bus identifier and location to the Go HTTP server. The backend processes
the request and stores location data in PostgreSQL. A static web
interface is included for the map/front-end portion of the project.

The main V1 goal is to establish a working end-to-end path:

1.  Receive bus location updates over HTTP.
2.  Handle incoming requests in Go.
3.  Process updates through the ingestion/worker component.
4.  Persist data in PostgreSQL.
5.  Provide a web interface for visualizing bus information.

The exact behavior of each endpoint and database operation should be
confirmed against the current source code before treating this
description as an API contract.

## Current capabilities

-   Go server entry point under `cmd/server`.
-   Bus HTTP handler under `internal/handler`.
-   Ingestion worker under `internal/ingestor`.
-   PostgreSQL storage code under `internal/store`.
-   SQL initialization migration under `migrations`.
-   Static frontend under `static`.
-   Environment-based configuration file (`.env`, local only).
-   Local load testing of the location endpoint.

## Architecture

``` text
Client / Load Test
        |
        | POST /api/bus/location
        v
Go HTTP Server
(cmd/server/main.go)
        |
        v
Bus Handler
(internal/handler/bus.go)
        |
        v
Ingestion Worker
(internal/ingestor/worker.go)
        |
        v
PostgreSQL Store
(internal/store/postgres.go)
        |
        v
PostgreSQL Database
(migrations/001_init.sql)

Browser
   |
   v
Static UI (static/index.html)
```

This diagram represents the intended component flow based on the
repository structure. Verify the actual call path and concurrency
behavior in the implementation.

## Repository structure

``` text
microtransit-go/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── handler/
│   │   └── bus.go               # Bus HTTP handlers / request handling
│   ├── ingestor/
│   │   └── worker.go            # Location ingestion / worker processing
│   ├── models/                  # Domain and data models
│   └── store/
│       └── postgres.go          # PostgreSQL persistence layer
├── migrations/
│   └── 001_init.sql             # Initial database schema
├── static/
│   └── index.html               # Static browser interface
├── .env                         # Local environment configuration; do not commit secrets
├── go.mod                       # Go module and dependency declarations
└── go.sum                       # Dependency checksums
```

### What each directory is for

-   **`cmd/server/`** --- executable entry point. Server setup and
    application wiring normally live here.
-   **`internal/handler/`** --- HTTP-facing code: decoding requests,
    validating input, and returning responses.
-   **`internal/ingestor/`** --- ingestion/worker logic for processing
    incoming location updates.
-   **`internal/models/`** --- shared data structures representing buses
    or location records.
-   **`internal/store/`** --- database access code, kept separate from
    HTTP handling.
-   **`migrations/`** --- SQL schema changes used to create or evolve
    the database.
-   **`static/`** --- frontend assets served to the browser.
-   **`.env`** --- local configuration. Keep credentials private and out
    of version control.
-   **`go.mod` / `go.sum`** --- Go module metadata and dependency
    integrity information.

## Technology stack

-   **Go** --- HTTP backend and request-processing components.
-   **PostgreSQL** --- persistent storage.
-   **SQL migrations** --- database schema initialization.
-   **HTML / JavaScript** --- static frontend; the repository's UI is
    intended to use a map.
-   **PowerShell** --- used for local endpoint load tests.

## Run locally

These steps are a general starting point. Adjust database name,
credentials, and environment variable names to match
`internal/store/postgres.go` and `cmd/server/main.go`.

### Prerequisites

-   Go installed.
-   PostgreSQL installed and running.
-   Git installed.
-   PowerShell is optional and only needed for the example test script.

### 1. Clone the repository

``` bash
git clone https://github.com/chidusurabhi123-collab/microtransit-go.git
cd microtransit-go
```

### 2. Configure environment variables

Create a local `.env` file using the variables expected by the
application. For example, a typical PostgreSQL configuration may look
like:

``` dotenv
DATABASE_URL=postgres://USERNAME:PASSWORD@localhost:5432/DATABASE_NAME?sslmode=disable
```

**Important:** `DATABASE_URL` is an example name, not a guarantee that
the current code reads that exact variable. Check the source and use its
actual configuration keys. Do not commit real passwords, API keys, or
production credentials.

### 3. Create the database and apply the migration

Create a PostgreSQL database, then run the initial migration against it:

``` bash
createdb YOUR_DATABASE_NAME
psql -d YOUR_DATABASE_NAME -f migrations/001_init.sql
```

If your local PostgreSQL setup requires a username or password, use your
configured local authentication method. Avoid putting database passwords
directly into shell history or source control.

### 4. Download Go dependencies

``` bash
go mod download
```

### 5. Start the server

From the repository root:

``` bash
go run ./cmd/server
```

If the application expects a different working directory or
configuration, follow the requirements in `cmd/server/main.go`. The
observed local endpoint during testing was:

``` text
http://localhost:8080/api/bus/location
```

### 6. Open the frontend

If the Go server serves the `static` directory, open the route
configured in the server. Otherwise, serve `static/index.html` using the
project's configured frontend workflow. The exact frontend URL depends
on the server setup.

## API

### Submit a bus location

The endpoint used during local tests is:

``` http
POST http://localhost:8080/api/bus/location
Content-Type: application/json
```

Example request body:

``` json
{
  "bus_id": "MYS-101",
  "lat": 12.2958,
  "lng": 76.6323,
  "speed": 45
}
```

Fields used in the test payload:

  Field      Meaning                                   Example
  ---------- ----------------------------------------- -----------
  `bus_id`   Bus identifier                            `MYS-101`
  `lat`      Latitude                                  `12.2958`
  `lng`      Longitude                                 `76.6323`
  `speed`    Reported speed; confirm units in source   `45`

The accepted fields, validation rules, response body, status codes, and
persistence semantics must be verified in `internal/handler/bus.go` and
related code. Do not assume the example documents every validation rule.

### Example PowerShell request

``` powershell
$body = @{
    bus_id = "MYS-101"
    lat    = 12.2958
    lng    = 76.6323
    speed  = 45
} | ConvertTo-Json -Compress

Invoke-RestMethod `
    -Method Post `
    -Uri "http://localhost:8080/api/bus/location" `
    -ContentType "application/json" `
    -Body $body
```

## Load-test notes

A local PowerShell test submitted 30,000 location requests to the
endpoint using `ForEach-Object -Parallel`. The user reported the
following two runs:

  --------------------------------------------------------------------------
  Run           Attempts   Successful       Failed Elapsed time Reported RPS
                            responses    responses              
  --------- ------------ ------------ ------------ ------------ ------------
  1               30,000       27,612        2,388     202.05 s          148

<<<<<<< HEAD
  2               30,000       30,000            0     259.56 s           88
=======
  2               30,000       30,000            0     289.56 s           88
>>>>>>> 21dd052 (feat: road-snapped route, shrinking blue line, Track Me fix, Palace final)
  --------------------------------------------------------------------------

Run 1 had a reported failure rate of 7.96%. Run 2 completed with zero
failed responses according to the test output.

### Interpret these numbers carefully

-   These are local development-machine results, not a production
    benchmark.
-   `ForEach-Object -Parallel` uses a bounded worker pool; this does
    **not** mean 30,000 connections were open simultaneously.
-   The reported RPS values do not match `attempts / elapsed time` for
    the displayed durations. Based on the supplied figures, the
    approximate aggregate rates are:
    -   Run 1: 30,000 / 202.05 ≈ 148.5 attempts/second.
    -   Run 2: 30,000 / 259.56 ≈ 115.6 attempts/second.
-   The source of the second run's reported `88 RPS` should be checked:
    the timer interval, numerator, or measurement scope may differ.
-   A PostgreSQL count of 30,001 was displayed after Run 2. That is not
    automatically proof of exactly 30,000 new rows: an existing row, a
    different table/count scope, or a duplicate can affect the count.
-   A successful HTTP response is not sufficient by itself to prove
    durable storage. Verify the relevant table, unique identifiers, and
    expected row count.
-   The first run's failures and second run's success show that
    reliability should be retested and failure messages recorded.

### Recommended benchmark procedure

1.  Record the exact commit, machine, Go version, PostgreSQL version,
    worker limit, and test script.
2.  Run a warm-up test, then at least three measured runs.
3.  Store each request's result and full error message.
4.  Calculate throughput using `completed requests / elapsed seconds`,
    and state whether failed requests are included.
5.  Measure p50, p95, and p99 response latency.
6.  Check CPU, memory, database connections, database write latency, and
    server logs.
7.  Verify unique IDs and persisted records after each run.
8.  Test increasing concurrency gradually, for example 10, 25, 50, and
    100 workers.
9.  Save benchmark output so later versions can be compared fairly.

Do not describe this as "30,000 simultaneous users" or
"production-ready" based on the current test alone.

## Security and configuration

-   Keep `.env` out of Git. Commit a `.env.example` with placeholder
    values instead.
-   If a real credential has ever been committed or shared, rotate it;
    deleting the file in a later commit does not remove it from Git
    history.
-   Validate latitude and longitude ranges and reject malformed or
    unreasonable payloads.
-   Use parameterized SQL queries; never build SQL by concatenating
    untrusted request data.
-   Add request timeouts, graceful shutdown, and bounded queues if the
    ingestion path uses workers/channels.
-   Limit request body size and return consistent HTTP status codes.
-   Add authentication and rate limiting before exposing write endpoints
    publicly.
-   Configure TLS and PostgreSQL access controls for deployed
    environments.
-   Avoid logging passwords, tokens, or unnecessary personal/location
    data.
-   Add database migrations that can be applied repeatably and tracked
    in source control.

## Roadmap

### V1 --- Core tracking backend

-   [x] Go server and bus location endpoint exist.
-   [x] PostgreSQL persistence is integrated.
-   [x] Static frontend exists.
-   [x] Local tests have been run with 10,000 sequential requests and
    30,000 requests through a parallel PowerShell test.
-   [ ] Make the benchmark repeatable and reconcile RPS calculations.
-   [ ] Verify data integrity with unique IDs and database assertions.
-   [ ] Add automated unit and integration tests.

### V2 --- Reliability and observability

-   [ ] Structured logging and request IDs.
-   [ ] Metrics for request count, error count, latency, queue depth,
    and database operations.
-   [ ] Graceful shutdown and correct cancellation of workers.
-   [ ] Input validation and consistent API error responses.
-   [ ] Database migration workflow and automated integration tests.
-   [ ] Repeatable load test with p50/p95/p99 latency and resource
    measurements.
-   [ ] Docker-based local development and CI checks.

### V3 --- Targeted C++ optimization

C++ is planned for components where profiling shows a real CPU-bound
bottleneck. Do not add a second language purely because it may be faster
in isolation. Define a narrow interface and benchmark the complete
Go-to-C++ path, including data conversion and inter-process or FFI
overhead.

Potential future use cases could include expensive route calculations or
high-volume geospatial computations, if those are proven to be
bottlenecks.

### V4 --- Python bus arrival-time prediction

-   [ ] Collect historical GPS observations with timestamps.
-   [ ] Obtain or derive trustworthy actual arrival-time labels.
-   [ ] Build features such as route, stop, time of day, day of week,
    speed, and historical travel duration.
-   [ ] Establish a simple baseline model before trying complex ML.
-   [ ] Evaluate prediction error on time-separated validation data.
-   [ ] Serve predictions through a separate Python API or model
    service.
-   [ ] Monitor prediction accuracy and data drift.

A model needs relevant historical data and reliable labels; GPS update
volume alone does not guarantee useful predictions.

### V5 --- Integrated system

-   Go: HTTP APIs, ingestion, and orchestration.
-   PostgreSQL: operational state and historical location records.
-   C++: only proven performance-critical components.
-   Python: model training and prediction service.
-   Monitoring and load testing: validate the complete system under
    realistic traffic.

## Known limitations and next steps

This README documents the repository layout and the test results
reported during development. It does not independently establish
production readiness, a maximum supported request rate, zero data loss
under every failure mode, or the exact behavior of every code path.

The highest-value next steps are to inspect the implementation, add
automated tests, reconcile the benchmark's RPS calculation, capture
errors, and verify persisted records by unique request ID.

## Contributing

For changes:

1.  Create a focused branch.

2.  Add or update tests.

3.  Run formatting and tests:

    ``` bash
    gofmt -w .
    go test ./...
    ```

4.  Run the benchmark with a recorded configuration.

5.  Document behavior changes and benchmark results.
