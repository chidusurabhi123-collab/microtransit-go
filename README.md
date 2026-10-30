# MicroTransit Go

A real-time bus-tracking prototype built with **Go, PostgreSQL, Python (FastAPI), Leaflet, OpenStreetMap, and OSRM**. The Go backend receives and stores GPS updates, manages route rooms and stops, and serves data to a browser map. A separate Python service exposes an ETA prediction endpoint.

> **Status:** Active prototype. The Python ETA service and model-training experiment are present in the repository. The currently published Go ETA handler does not yet call the Python prediction client, so end-to-end Python ETA integration is still a work item.

## Features

### Go backend

- HTTP API for bus locations, buses, rooms, stops, and route-distance information.
- PostgreSQL persistence using `pgx`.
- Buffered GPS ingestion channel and worker goroutines.
- Latest bus location upserted by bus ID.
- Route rooms with an admin token returned at creation.
- Admin-token check when adding stops.
- Ordered route stops and stored segment-distance fields.
- Haversine distance calculations.

### Browser map

- Leaflet map with OpenStreetMap tiles.
- Bus and stop markers, including a distinct final-stop marker.
- OSRM road geometry request with straight-line fallback.
- Remaining-route overlay.
- Browser geolocation control ("Track Me").

### Python ETA service

- FastAPI application with `/health` and `/predict` endpoints.
- ETA input fields: `distance_km` and `speed_kmh`.
- Linear Regression model loaded from `python-eta/model.pkl` when available.
- Training script creates a model artifact using generated synthetic data.
- Mathematical fallback when no model artifact is available.

**Important ML limitation:** `train.py` generates synthetic rows using a known distance/speed/peak-hour formula and trains a Linear Regression model on those generated labels. This is a useful integration prototype, but it is not a model trained or validated on real bus GPS histories, traffic observations, or actual arrival-time labels. Do not claim real-world ETA accuracy until that data and evaluation exist.

## Architecture

```text
Bus GPS client / load generator
             |
             v
        Go HTTP API
             |
       +-----+---------------------+
       |                           |
       v                           v
GPS ingestion queue           Room / stop handlers
       |                           |
       v                           v
Worker goroutines          Route and distance logic
       |                           |
       +-------------+-------------+
                     |
                     v
                 PostgreSQL

Browser map <---- Go API
     |
     +---- OpenStreetMap tiles
     |
     +---- OSRM road geometry

Python ETA service (FastAPI)
     |
     +---- /health
     +---- /predict
     |
     +---- Linear Regression model artifact
           trained by train.py on synthetic data

Current integration gap:
Go constructs an ETA client, but the published Go
GET /api/rooms/{room_id}/eta handler does not yet call it.
```

## Repository structure

```text
microtransit-go/
├── cmd/
│   └── server/
│       └── main.go                 # Go application wiring and routes
├── internal/
│   ├── eta/
│   │   └── client.go               # Go HTTP client for Python ETA service
│   ├── geo/                         # Geospatial calculations
│   ├── handler/
│   │   ├── bus.go                   # Bus location and list handlers
│   │   ├── room_handler.go          # Room endpoints
│   │   └── stop_handler.go          # Stop and route-distance endpoints
│   ├── ingestor/
│   │   └── worker.go                # Buffered queue and worker goroutines
│   ├── models/                      # Domain models
│   └── store/
│       ├── postgres.go              # PostgreSQL connection and bus storage
│       ├── room_store.go            # Room persistence
│       └── stop_store.go            # Stop persistence
├── migrations/
│   └── 001_init.sql                 # Initial database schema
├── python-eta/
│   ├── app.py                       # FastAPI application
│   ├── models.py                    # Model loading and ETA inference/fallback
│   ├── train.py                     # Synthetic-data model training
│   ├── model.pkl                    # Serialized model artifact (if committed)
│   └── requirements.txt             # Python dependencies
├── static/
│   └── index.html                   # Leaflet frontend
├── .env.example                     # Safe placeholder configuration (recommended)
├── .gitignore                       # Exclude secrets and local artifacts (recommended)
├── LICENSE                          # MIT License (add/commit this file)
├── go.mod
├── go.sum
└── README.md
```

The tree above documents the paths inspected in the public repository and the expected local configuration files. Keep it synchronized with the actual files in your branch.

## Technology stack

| Component | Technology | Purpose |
|---|---|---|
| Backend | Go | HTTP API, ingestion, workers, application logic |
| Database | PostgreSQL | Latest bus locations, rooms, and stops |
| Database driver | pgx | PostgreSQL connection pooling and queries |
| Frontend | HTML, CSS, JavaScript | Browser interface |
| Map | Leaflet | Map rendering and interaction |
| Map tiles | OpenStreetMap | Base map tiles |
| Road geometry | OSRM | Road-following route geometry |
| ETA API | Python + FastAPI | Prediction endpoint |
| ML | NumPy + scikit-learn Linear Regression | Experimental ETA model |
| Model serialization | joblib | Save/load model artifact |
| Load testing | PowerShell | Local HTTP load tests |

## Run locally

### Prerequisites

- Go version compatible with `go.mod`.
- PostgreSQL installed and running.
- Python 3.10+ recommended for the Python dependencies.
- Git.
- A browser with geolocation support for Track Me.

### 1. Clone the repository

```bash
git clone https://github.com/chidusurabhi123-collab/microtransit-go.git
cd microtransit-go
```

### 2. Configure PostgreSQL

Create the database:

```bash
createdb microtransit
```

Apply the initial schema:

```bash
psql -d microtransit -f migrations/001_init.sql
```

The Go entry point reads `DATABASE_URL`. Its current local fallback is:

```text
postgres://postgres:postgres@localhost:5432/microtransit?sslmode=disable
```

Use your own local PostgreSQL credentials. Do not use the fallback credentials in a deployed environment.

Create a local `.env` file in the repository root:

```dotenv
DATABASE_URL=postgres://YOUR_USER:YOUR_PASSWORD@localhost:5432/microtransit?sslmode=disable
PYTHON_ETA_URL=http://127.0.0.1:8000
```

The Go server calls `godotenv.Load()` from the repository root. Keep the real `.env` local and out of Git.

### 3. Start the Python ETA service

Open a terminal:

```bash
cd python-eta

python -m venv .venv
```

Activate the virtual environment.

Windows PowerShell:

```powershell
.\.venv\Scripts\Activate.ps1
```

macOS/Linux:

```bash
source .venv/bin/activate
```

Install dependencies:

```bash
python -m pip install -r requirements.txt
```

Train/generate the local model artifact:

```bash
python train.py
```

Start FastAPI:

```bash
python -m uvicorn app:app --host 127.0.0.1 --port 8000
```

Check health:

```text
http://127.0.0.1:8000/health
```

Test prediction:

```bash
curl -X POST http://127.0.0.1:8000/predict \
  -H "Content-Type: application/json" \
  -d '{"distance_km": 2.5, "speed_kmh": 30}'
```

Expected response shape:

```json
{
  "eta_minutes": 5.0,
  "hour": 14
}
```

The numeric value depends on the current hour, model artifact, and generated training data. The example is illustrative.

### 4. Start the Go backend

From the repository root, in a second terminal:

```bash
go mod download
go run ./cmd/server
```

The server listens on port `8080` and serves the frontend from `static/`.

Open:

```text
http://localhost:8080/
```

**Current integration note:** Start the Python service so it is available for future Go integration and independent API testing. In the published source, the Go `/api/rooms/{room_id}/eta` handler still returns distance/progress fields without invoking `Eta.Predict`. Wire and test that call before claiming the UI/API displays Python-generated ETA.

## API

The following routes are registered by the published Go server:

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/api/bus/location` | Submit a GPS update |
| `GET` | `/api/buses` | Get latest bus locations |
| `POST` | `/api/rooms` | Create a route room |
| `GET` | `/api/rooms/{room_id}` | Get room details |
| `POST` | `/api/rooms/{room_id}/stops` | Add a stop (admin token required) |
| `GET` | `/api/rooms/{room_id}/stops` | List room stops |
| `GET` | `/api/rooms/{room_id}/eta` | Return route-distance/progress data |

### Submit bus location

```http
POST /api/bus/location
Content-Type: application/json
```

```json
{
  "bus_id": "MYS-101",
  "lat": 12.2958,
  "lng": 76.6394,
  "speed": 30
}
```

The handler submits the decoded location to the in-memory ingestion queue and responds with a queued status. A queued response is not proof that PostgreSQL has committed the write.

### Create a room

```http
POST /api/rooms
Content-Type: application/json
```

```json
{
  "name": "MYS-101"
}
```

The response contains a room ID and an admin token. Treat the token as a secret.

### Add a stop

```http
POST /api/rooms/MYS-101/stops
Content-Type: application/json
X-Admin-Token: YOUR_ADMIN_TOKEN
```

```json
{
  "name": "Kuvempunagar",
  "lat": 12.2958,
  "lng": 76.6394
}
```

### List stops

```http
GET /api/rooms/MYS-101/stops
```

### Get route distance/progress data

```http
GET /api/rooms/MYS-101/eta
```

The current handler chooses the geographically nearest stop using Haversine distance and returns route-distance/progress fields. This is a prototype heuristic; loops, crossings, nearby parallel roads, and GPS noise can cause the nearest stop not to be the next stop along the route.

### Python prediction API

Health check:

```http
GET http://127.0.0.1:8000/health
```

Prediction:

```http
POST http://127.0.0.1:8000/predict
Content-Type: application/json
```

```json
{
  "distance_km": 2.5,
  "speed_kmh": 30
}
```

Response shape:

```json
{
  "eta_minutes": 5.0,
  "hour": 14
}
```

## ETA model: what it currently does

`python-eta/train.py` creates 2,000 synthetic training rows. Its target is generated from a formula using distance, speed, a peak-hour traffic multiplier, and random noise. It fits a scikit-learn `LinearRegression` model and saves it as `model.pkl` using joblib.

`python-eta/models.py` loads that artifact and predicts using four input features: distance, speed, hour, and traffic multiplier. If the model is unavailable, it falls back to the distance/speed formula.

This is useful for proving the service boundary and model-serving workflow. It does **not** demonstrate that the model has learned real bus traffic patterns. For a credible transit ETA model, collect timestamped bus GPS histories and actual stop-arrival times, engineer route/time/traffic features, use a time-based train/validation split, and report MAE (mean absolute error) against a simple baseline.

Also note that the current `/health` implementation returns `model_loaded: true` unconditionally. Improve it to report the actual model-loading state before relying on that field operationally.

## Local load-test results

The following local PowerShell results were reported during development:

| Run | Attempts | Successful | Failed | Elapsed time | Reported RPS |
|---|---:|---:|---:|---:|---:|
| 1 | 30,000 | 27,612 | 2,388 | 202.05 s | 148 |
| 2 | 30,000 | 30,000 | 0 | 259.56 s | 88 |

Recalculated successful-response throughput:

- Run 1: 27,612 / 202.05 ≈ **136.7 successful responses/second**.
- Run 2: 30,000 / 259.56 ≈ **115.6 successful responses/second**.

The reported RPS values do not match successful responses divided by elapsed time. For future benchmarks, define the numerator clearly, record the worker throttle limit, capture full error messages, and measure p50/p95/p99 latency. `ForEach-Object -Parallel` uses a bounded worker pool; 30,000 submitted requests does not mean 30,000 simultaneous open connections.

A PostgreSQL total row count of 30,001 is not by itself proof that exactly 30,000 new updates were persisted. Verify test-run IDs or unique bus identifiers.

These are local development measurements, not production capacity claims.

## Known issues and next engineering tasks

1. **Connect Go to Python:** the Go ETA client is created in `main.go`, but `StopHandler.GetETA` does not yet call `Eta.Predict`.
2. **Check Python HTTP responses:** the Go client should check non-2xx status codes, JSON decode errors, request marshal errors, and invalid response values.
3. **Handle zero speed:** avoid division by zero in the fallback ETA calculation.
4. **Fix model health reporting:** `/health` should report the true model-loading state, not always `true`.
5. **Replace synthetic training data:** use real timestamped journey data and actual arrival labels before making accuracy claims.
6. **Fix room creation conflicts:** avoid returning a newly generated admin token when an existing room caused the insert to do nothing.
7. **Use transactions for stop updates:** updating `dist_to_final` and inserting a stop must succeed or fail atomically.
8. **Prevent concurrent stop-order races:** serialize route edits or enforce an appropriate database constraint.
9. **Check all database errors:** do not silently discard query, scan, or row-iteration errors.
10. **Improve route matching:** nearest-stop selection is not robust route-progress tracking.
11. **Make ingestion durable if required:** the in-memory queue can lose acknowledged updates on process failure, and worker write failures are currently logged rather than retried.
12. **Protect secrets:** remove `.env` from Git and rotate any real credentials that were exposed.

## Roadmap

### Core tracking and map

- [x] Go HTTP server and bus-location endpoint.
- [x] PostgreSQL latest-location storage.
- [x] Buffered ingestion queue and worker goroutines.
- [x] Leaflet map and bus/stop markers.
- [x] Room and stop API scaffolding.
- [x] OSRM road-geometry attempt with fallback.
- [x] Browser geolocation control.

### ETA service

- [x] Python FastAPI `/health` and `/predict` endpoints.
- [x] Synthetic-data Linear Regression training script.
- [x] Model serialization/loading with joblib.
- [x] Mathematical fallback when no model is loaded.
- [ ] Call the Python service from the Go route ETA handler.
- [ ] Validate service errors, timeouts, and fallback behavior.
- [ ] Collect real bus journey and stop-arrival data.
- [ ] Evaluate model against a baseline with time-separated validation.
- [ ] Report MAE and monitor prediction quality.

### Reliability and deployment

- [ ] Automated Go unit and integration tests.
- [ ] Python tests for model loading, inputs, and predictions.
- [ ] Go-to-Python integration tests.
- [ ] Transaction-safe stop updates and route-order constraints.
- [ ] Graceful shutdown and worker lifecycle management.
- [ ] Retry/dead-letter handling for failed GPS writes.
- [ ] Structured logs and request metrics.
- [ ] Repeatable load tests with latency percentiles.
- [ ] Docker Compose for Go, PostgreSQL, and Python.
- [ ] CI checks, HTTPS, rate limits, and stronger authentication.

## Security and configuration

- Do not commit `.env`. Commit `.env.example` with placeholder values only.
- If a real password, token, or connection string has been committed to a public repository, rotate it. Removing the file in a later commit does not remove it from Git history.
- Treat `X-Admin-Token` as a bearer secret.
- Use HTTPS outside localhost.
- Validate request bodies, coordinate ranges, IDs, and speed values.
- Use parameterized SQL, request timeouts, body-size limits, and rate limiting.
- Restrict PostgreSQL access and use a least-privilege database role.
- The public OSRM demo service is an external dependency, not a production SLA.

## License

This project is intended to be distributed under the MIT License. See [`LICENSE`](LICENSE). Ensure the `LICENSE` file is committed at the repository root.

## Contributing

1. Create a focused branch.
2. Add or update tests for behavior changes.
3. Format and test Go:

   ```bash
   gofmt -w .
   go test ./...
   ```

4. Test the Python service and Go-to-Python integration.
5. Record benchmark configuration and results.
6. Update this README when routes, environment variables, or behavior change.
