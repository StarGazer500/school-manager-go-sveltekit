# School Manager: Go (Gin) + SvelteKit

A small school management app for learning Gin and SvelteKit. It manages
**students**, **teachers**, **courses** and **enrollments**. Both halves use a
feature-based layout: code is grouped by business feature, not by technical layer.

```
gosvelte/
├── backend/        Go + Gin + pgrepo + PostgreSQL → http://localhost:8080
├── frontend/       SvelteKit 3 + Svelte 5         → http://localhost:5173
└── observability/  Grafana + Tempo + Prometheus + Loki → http://localhost:3000
```

You need **Go** (any version from 1.21 up: it downloads Go 1.27.3, which `backend/go.mod`
asks for, automatically), **Node 22.17+** and **Docker** (for PostgreSQL and Redis).

## Running it

With `make`, from the project root:

```bash
make obs-up     # optional: Grafana + traces/metrics/logs storage on :3000 (see Observability)
make dev        # Postgres + Redis (Docker) + backend on :8080 + frontend on :5173 (Ctrl+C stops the servers)
make check      # go vet/test + svelte-check
```

Each folder has its own Makefile too. Run `make` in `backend/` or `frontend/` to list
the targets (`run`, `db-up`, `db-down`, `db-reset`, `db-shell`, `cache-shell`, `cache-flush`,
`migrate-*`, `dev`, `check`, …). The frontend Makefile switches to the Node version in `.nvmrc`
automatically. Postgres and Redis keep running after Ctrl+C; stop them with `make -C backend db-down`.

> **pgrepo and rediscache are private modules.** Go can only download them with
> `GOPRIVATE=github.com/StarGazer500` set (and SSH access to the repos). `backend/Makefile`
> sets this for you. If you run `go` commands directly, export it first, or run
> `go env -w GOPRIVATE=github.com/StarGazer500` once.

Without `make`:

**Backend.** The API applies the database migrations itself when it starts:

```bash
cd backend
docker compose up -d --wait db redis        # PostgreSQL on localhost:5433, Redis on 6380
export GOPRIVATE=github.com/StarGazer500
go run ./cmd/api
```

**Frontend** (SvelteKit 3 needs Node ≥ 22.17; `.nvmrc` pins Node 22):

```bash
cd frontend
nvm use          # switches to the Node 22 version from .nvmrc
npm install
npm run dev
```

Open http://localhost:5173. Try creating a teacher, then a course, then a student, then enrolling the student.

Settings (environment variables):

| Where    | Variable       | Default                                                    |
| -------- | -------------- | ---------------------------------------------------------- |
| backend  | `PORT`         | `8080`                                                     |
| backend  | `DATABASE_URL` | `postgres://school:school@localhost:5433/school?sslmode=disable` |
| backend  | `REDIS_URL`    | `redis://localhost:6380/0`                                 |
| backend  | `LOG_FORMAT`   | `console` (readable, colored) or `json` (one JSON object per line), for stdout |
| backend  | `LOG_LEVEL`    | `info` (also `debug`, `warn`, `error`)                     |
| backend  | `LOG_FILE`     | `logs/api.log` via the Makefile (empty = stdout only); JSON lines the collector ships to Loki |
| backend  | `METRICS_PORT` | `9464` (`/metrics` for Prometheus, kept off the public API port) |
| backend  | `OTEL_SERVICE_NAME` | `school-api` (the service's name in Grafana)          |
| backend  | `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4318` (where traces are sent) |
| backend  | `OTEL_SDK_DISABLED` | `false` (`true` turns tracing and metrics off)         |
| backend  | `APP_ENV`      | `development` (shown on all telemetry)                     |
| frontend | `API_TARGET`   | `http://localhost:8080` (where the dev proxy sends `/api`) |
| frontend | `FARO_TARGET`  | `http://localhost:8027` (where the dev proxy sends `/collect`, browser telemetry) |

Postgres and Redis run from [backend/docker-compose.yml](backend/docker-compose.yml) on ports
**5433** and **6380**, so they don't clash with another Postgres or Redis on the defaults
(5432, 6379). Postgres keeps its data in a Docker volume that survives restarts
(`make -C backend db-reset` deletes it). Redis is only a cache, keeps nothing on disk, and starts empty.

## Backend layout

```
backend/
├── docker-compose.yml           PostgreSQL 18 + Redis 7.4 for local development
├── migrations/                  numbered SQL files (goose), embedded into the binary
├── cmd/
│   ├── api/main.go              entrypoint: config → telemetry → logger → DB → migrate → server → shutdown
│   └── migrate/main.go          `make migrate-*` commands: up, down, status, new
└── internal/
    ├── core/                    one package (`core`): infrastructure set up at startup
    │   ├── config.go            env vars (LoadConfig)
    │   ├── telemetry.go         OpenTelemetry: traces (OTLP) and metrics (/metrics) (SetupTelemetry)
    │   ├── database.go          pgx pool (OpenDatabase), Postgres errors → app errors (MapError)
    │   ├── migrate.go           goose migrations (Migrate, Migrator)
    │   └── cache.go             Redis (OpenRedis), caches for services (NewCache[T]), not-found helpers
    ├── logger/                  builds the zerolog logger (stdout + JSON log file, trace-ID hook)
    ├── server/                  the HTTP layer (package `server`):
    │   ├── server.go            composition root: middleware + wires every feature together
    │   ├── middleware.go        tracing, request logging, panic recovery (used only by server.go)
    │   └── health.go            /healthz and /readyz
    ├── shared/                  one package (`shared`) used by every feature:
    │   └── commonutils.go       error kinds (NotFound / Conflict / Invalid, type AppError) +
    │                            Gin helpers: errors → HTTP status, ID parsing, validation messages
    └── features/
        ├── students/
        │   ├── model.go         DB entity with `db` tags + TableName() (+ list filter)
        │   ├── dto.go           JSON request/response shapes + mapping
        │   ├── repository.go    Postgres access with pgrepo (a concrete struct, no interface)
        │   ├── service.go       business rules (no HTTP, no SQL); repository + cache injected
        │   ├── handler.go       HTTP ↔ service
        │   └── routes.go        mounts endpoints
        ├── teachers/            same shape
        ├── courses/             same shape (+ optional teacher FK)
        └── enrollments/         same shape, no cache (+ depends on courses),
                                 + telemetry.go: the Enroll span and enrollment counters
```

A request flows **router → handler → service → repository → DB**, and each layer has one job:

- **Handler**: parse/validate input (`ShouldBindJSON` + `binding:` tags), call the service, write JSON.
- **Service**: business rules, e.g. "a course can't take more students than its capacity"
  ([enrollments/service.go](backend/internal/features/enrollments/service.go)).
- **Repository**: database access only, through [pgrepo](https://github.com/StarGazer500/pgrepo),
  which writes the SQL from each model's `db` tags. It returns `shared` error kinds, so services
  never see SQL or driver errors.
- **Errors**: services return `shared.NotFound(...)` and similar, and handlers turn them into 404/409/400:
  `status, msg := shared.ErrorResponse(c, err)` then `c.AbortWithStatusJSON(status, gin.H{"error": msg})`.
  Unknown errors become 500 and are logged.

Things worth reading in the code:

- **Cross-feature dependency.** `enrollments` needs to look up courses. It defines a tiny
  `CourseReader` interface on *its* side, and `courses.Service` happens to satisfy it.
  The two features stay loosely coupled.
- **The database enforces integrity.** Deleting a teacher sets their courses' `teacher_id`
  to NULL, and deleting a student or course cascades to its enrollments. Postgres doesn't
  index foreign keys by itself, so the migrations add those indexes.
- **Read models.** `enrollments.Detail` (an enrollment plus the student's name and course
  code) is not a table. The repository loads the enrollments, then all their students in one
  query and all their courses in another with `pgrepo.ManyToOne`, instead of one query per row.
- **pgrepo basics.** `Get`/`Update` return `nil` when no row matches, which the repositories
  turn into a 404. `Create`/`Update` take a `map[string]any` of columns and return the
  saved row, so database defaults like `id` and `created_at` come back filled in.

### Caching (rediscache)

Looking up one student, teacher or course by id goes through a Redis cache, built with
[rediscache](https://github.com/StarGazer500/rediscache). The cache is **injected into each
service** next to its repository, in [server.go](backend/internal/server/server.go):

```go
studentSvc := students.NewService(
	students.NewRepository(db),                    // Postgres
	core.NewCache[students.Student](rdb, "student:"),  // Redis
)
```

The service holds the concrete types, with no interfaces in between:

```go
type Service struct {
	repo  *Repository                 // Postgres
	cache *rediscache.Cache[Student]  // Redis
}
```

> **Interfaces in this project** follow the Go habit *"discover interfaces, don't design
> them"*: add one when a second implementation is needed at the same time, e.g. a fake
> repository or cache for tests. Until then, concrete types are simpler. The one interface
> left is `CourseReader` in enrollments, which has a different job: it limits enrollments
> to the single courses method it needs.

- **Reads** (`Service.Get`) use `GetOrLoad`: the cached value if there is one, otherwise load
  it from the repository and store it. `core.ToCache`/`core.FromCache` translate between
  the app's not-found error and rediscache's, so misses are remembered and handlers still
  answer 404. Entries last 10 minutes. Concurrent misses for the same id
  share one query, and unknown ids are remembered as "not found" for 1 minute.
- **Writes** (`Create`, `Update`, `Delete`) go to Postgres first, then remove the cached entry,
  so the next read loads fresh data. `Update` reads the row from Postgres, not the cache, so
  it starts from current data. Lists (`List`) aren't cached: every search would be its own entry.
- **Tags across features.** Deleting a teacher makes Postgres set `teacher_id` to NULL on
  their courses, which the courses cache can't see. So every cached course carries the
  `courses` tag (`courses.CacheTag`), and the teachers service invalidates it in `Delete`.
- **Redis is optional.** If it's down, the API still starts, reads go straight to Postgres
  (a couple of milliseconds, since the client gives up quickly), and warnings are logged.
- **Look inside** with `make -C backend cache-shell`, then `KEYS *` or `TTL student:1`.
  `make -C backend cache-flush` empties it.

Changed data directly in the database (e.g. with `db-shell`)? The cache won't notice until the
entry expires; run `cache-flush`.

### Logging (zerolog)

Logs are **structured**: a message plus named fields, written with
[zerolog](https://github.com/rs/zerolog). In development they print as readable lines:

```
12:00:01 INF request method=GET route=/api/v1/students status=200 duration_ms=3.36 trace_id=8a7e408c… span_id=5602ea22…
12:00:01 WRN request method=GET route=/api/v1/students/:id status=404 duration_ms=2.56 trace_id=dbb9bd3f… span_id=…
```

**The app only writes logs; it doesn't ship them.** Each line goes to stdout (readable, or
JSON with `LOG_FORMAT=json`) and, with `LOG_FILE` set (the Makefile sets `logs/api.log`), to
that file **as JSON**. The OpenTelemetry Collector tails the file and sends the lines to Loki.
That's how logs are usually handled in production: a slow or broken log backend can never
slow down the app, and the same setup works for any service in any language.

- **One line per request** from the request-logging middleware (`server/middleware.go`): `INF` for success, `WRN` for 4xx,
  `ERR` for 5xx, with the route pattern (`/api/v1/students/:id`) and `duration_ms`.
  Health checks (`/healthz`, `/readyz`) aren't logged: they'd bury real traffic.
- **The trace ID is the request ID.** A zerolog *hook* (internal/logger) adds `trace_id` and
  `span_id` to every line logged with a request's context. It's in the `X-Trace-ID` response
  header and in Grafana too: one ID connects all three.
- **How code gets the logger.** There's no global logger. The rule is *injected at startup,
  carried in `ctx` during a request*:
  - **Outside a request** (startup, migrations, Redis, the caches' error handler), `main.go`
    passes the logger as a parameter: `core.Migrate(ctx, db, log)`,
    `core.OpenRedis(ctx, url, log)`, `core.NewCache[T](rdb, prefix, log)`, `server.New(db, rdb, log, …)`.
  - **Inside a request**, the middleware stores a logger carrying the request's context. Get
    it with `zerolog.Ctx(ctx)`, and the line gets the request's `trace_id`:

    ```go
    zerolog.Ctx(ctx).Info().Int64("course_id", id).Msg("course is full")
    ```

    Inside a span you started yourself, add `.Ctx(ctx)` to the event for that span's `span_id`.
  Outside a request `zerolog.Ctx(ctx)` returns a logger that writes **nothing**, so pass a
  logger in there instead.
- **go-redis's own messages** go through zerolog at debug level (`source=go-redis`).
- **Unexpected errors** (anything that isn't a `shared.AppError`) are logged by
  `shared.ErrorResponse` with the real error and recorded on the request's trace, while
  the client only sees `internal server error` plus the `X-Trace-ID` to report. **Panics**
  are caught by the recovery middleware, logged and recorded on the trace with their stack, and
  answered with a 500.
- **Rotating the file** (so it doesn't grow forever) is the platform's job: `logrotate` on
  a server, or the container runtime when the app runs in Docker/Kubernetes and logs to
  stdout only. The collector follows rotated files.

> **Why not `otelzerolog`?** The `vincentfree/otelzerolog` package only adds `trace_id`/
> `span_id` to a line, and you have to pass the span on every call
> (`.Func(otelzerolog.AddTracingContext(span))`). The ~10-line hook in internal/logger does the
> same automatically for every line logged with a request's context.

### Database migrations (goose)

The schema lives in [backend/migrations/](backend/migrations/) as numbered SQL files. Each
file has an **Up** part (apply the change) and a **Down** part (undo it):

```sql
-- +goose Up
ALTER TABLE students ADD COLUMN phone TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE students DROP COLUMN phone;
```

goose keeps a `goose_db_version` table recording which files have run, so each one runs
exactly once, in order. The API applies new migrations every time it starts.

To change the schema (run these in `backend/`):

```bash
make migrate-new name=add_phone_to_students   # creates migrations/00005_add_phone_to_students.sql
# edit the file: write the Up and Down SQL
make migrate-up                               # apply it (or just restart the API)
make migrate-status                           # see which migrations have run
make migrate-down                             # undo the most recent one
```

Rules of thumb:

- **Never edit a migration that has already run.** goose won't run it again, so the change
  would never reach the database. Write a new migration instead.
- **Always write the Down part**, and test it with `migrate-down` then `migrate-up`.
- After a schema change, update the matching `model.go` (and the DTOs, if the API should
  expose the new column). pgrepo needs a struct field for every column, so add the field in
  the same change as the migration.
- The first four migrations use `CREATE TABLE IF NOT EXISTS` only so that databases created
  before goose was added could switch over. New migrations don't need it.

### API

All routes are under `/api/v1`:

| Method                  | Path                                     | Notes                        |
| ----------------------- | ---------------------------------------- | ---------------------------- |
| GET / POST              | `/students`                              | `?q=` search, `?grade=` filter |
| GET / PUT / DELETE      | `/students/:id`                          |                              |
| GET / POST              | `/teachers`                              | `?q=`                        |
| GET / PUT / DELETE      | `/teachers/:id`                          |                              |
| GET / POST              | `/courses`                               | `?q=`, `?teacherId=`         |
| GET / PUT / DELETE      | `/courses/:id`                           |                              |
| GET / POST              | `/enrollments`                           | `?studentId=`, `?courseId=`  |
| DELETE                  | `/enrollments/:id`                       | withdraw                     |
| GET                     | `/healthz`                               | liveness: the process is up  |
| GET                     | `/readyz`                                | readiness: Postgres (and Redis) reachable |

```bash
curl -X POST localhost:8080/api/v1/students \
  -H 'content-type: application/json' \
  -d '{"firstName":"Ada","lastName":"Lovelace","email":"ada@school.test","gradeLevel":10}'
```

## Observability

**Observability** answers "what is the app doing, and why?" from the outside, without adding
`fmt.Println` and redeploying. It rests on three signals, covering both the Go API and the
browser:

| Signal | Answers | Stored in | Sent by |
| --- | --- | --- | --- |
| **Traces** | "Where did *this* request spend its time, and where did it fail?" | Tempo | API (OpenTelemetry), browser (Faro) |
| **Metrics** | "How is the system doing overall?" (rate, errors, latency, pool, cache…) | Prometheus | API's `/metrics`, scraped |
| **Logs** | "What exactly happened?" Each line carries its request's `trace_id` | Loki | API log file, browser (Faro) |

### Running it

```bash
make obs-up      # start the stack; first run creates observability/.env with a random Grafana password
make dev         # the API and the browser send telemetry automatically
make obs-open    # Grafana: http://localhost:3000, user admin
make -C observability password   # shows the Grafana password
```

Then open **Dashboards → School → School API › Overview**. `make obs-down` stops the stack
(data is kept), `make obs-reset` deletes the data too. The app works the same without the
stack: it logs one warning a minute that it can't send traces.

### How it fits together

One container per job, as in production:

```
          ┌── traces (OTLP :4318) ───────────┐
Go API ───┤                                  ▼
          ├── logs/api.log (JSON) ──▶  OpenTelemetry Collector ──▶ Tempo (traces, 3 days)
          └── :9464/metrics ◀─┐        (memory limit, batching,  ──▶ Loki  (logs, 7 days)
                              │         disk retry queue,
Browser ── Faro (/collect) ──────────▶  tail sampling)
                              │
                       Prometheus (metrics, 15 days) ◀── Tempo: service graph
                              │
                           Grafana  ◀── dashboards + alert rules from files
```

```
observability/
├── docker-compose.yml        collector, tempo, loki, prometheus, grafana (pinned versions)
├── .env.example              copied to .env (not committed): Grafana password, sampling %
├── Makefile                  up, down, reset, restart, logs, ps, open, password
├── collector/config.yaml     receivers → processors → exporters (the pipeline above)
├── tempo/tempo.yaml
├── loki/loki.yaml
├── prometheus/prometheus.yml what to scrape (the API, and the stack itself)
└── grafana/
    ├── provisioning/
    │   ├── datasources/      Prometheus, Tempo, Loki + the links between them
    │   ├── dashboards/       "load dashboards/ into a School folder"
    │   └── alerting/         alert rules
    └── dashboards/school-api-overview.json
```

Dashboards, alerts and every service's config are **files in git**, reviewed and versioned
like the app. If you change a dashboard in the UI, export it (Share → Export → Save to file)
over the JSON here, or the change is lost when the container is recreated.

**Why each piece:**
- **The collector in the middle**: apps only send to one place; buffering, retries, sampling
  and routing are configured once, here, for every service.
  - `memory_limiter` refuses data rather than crash under a traffic spike.
  - `batch` sends in bulk: far fewer requests to Tempo and Loki.
  - **Disk-backed retry queues**: if Tempo or Loki restarts, telemetry waits on disk and is
    sent when they're back, instead of being lost.
  - **Tail sampling**: the collector waits until a trace is complete, then keeps **every trace
    with an error or slower than 500 ms**, and 25% of the rest (`TRACE_SAMPLING_PERCENT` in
    `.env`). The interesting traces are never lost, and storage stays small. (A side effect:
    some "exemplar" dots on charts point to traces that weren't kept.)
- **Prometheus pulls** (scrapes `/metrics`) instead of the app pushing. Because Prometheus
  asks, it knows when the app stops answering: `up{job="school-api"}` is 1 or 0, which is
  what the "API down" alert uses. `/metrics` is on its own port (9464), never exposed with
  the public API. Prometheus also scrapes the collector, Tempo, Loki and Grafana: the stack
  watches itself (e.g. `otelcol_exporter_send_failed_spans_total`).
- **Retention**: traces 3 days, logs 7 days, metrics 15 days.
- **Grafana requires a login**, with the password from `.env` (never committed). Only
  Grafana, the collector's inputs and Prometheus are published, and only on 127.0.0.1.

### Buffering under high traffic

Every hop is built to **drop or queue rather than slow the app down**:

| Where | Mechanism | When full |
| --- | --- | --- |
| API: traces | in-memory batch queue (2,048 spans, `OTEL_BSP_MAX_QUEUE_SIZE`), exported in the background | new spans are dropped; requests are never slowed |
| API: logs | written to stdout/file only | the collector reads at its own pace |
| API: metrics | counters and histograms summed in memory | constant memory, whatever the traffic |
| Browser (Faro) | batched in the page, sent in the background | dropped; the page is unaffected |
| Collector | `memory_limiter`, then `batch`, then a disk queue per exporter | refuses new data (senders retry) rather than crash |

### What gets sent

**Traces** — one trace can now run **from the browser to the database**. Faro adds a
`traceparent` header to the page's `/api` calls, and the API continues that trace:

```
school-web  GET /api/v1/enrollments            ← the browser's fetch (Faro)
school-api    POST /api/v1/enrollments         ← HTTP span (otelgin)
school-api      enrollments.Enroll             ← business span, with course.taken/capacity
school-api        get course:7                 ← Redis call (redisotel)
school-api        SELECT / INSERT              ← SQL (otelpgx; pgrepo runs on pgx)
```

Unexpected errors and panics are recorded on the request's span with the error message (and
stack), so a failed trace shows *why* it failed.

**Metrics** (scraped every 15 s from `:9464/metrics`):

| Metric | What |
| --- | --- |
| `http_server_request_duration_seconds_*` | Requests per route (`http_route`) and status: rate, errors, latency |
| `db_client_operation_duration_seconds_*`, `pgxpool_*` | SQL query duration; connection pool in use / idle / waits |
| `db_client_connections_*` | Redis connection pool |
| `app_cache_hits_total`, `_misses_total`, `_loads_total` | Per cache (`cache="student"`…) |
| `app_enrollments_created_total`, `app_enrollments_rejected_total{reason}` | Business counters |
| `go_memory_*`, `go_goroutine_count` | Go runtime |
| `up{job="school-api"}` | 1 while Prometheus can scrape the API |

**Logs**: the API's log file (see [Logging](#logging-zerolog)), and from the browser:
`kind=exception` (errors), `kind=measurement` (Web Vitals: LCP, FCP, TTFB, INP, CLS) and
`kind=event` (page loads, sessions, API calls).

All signals carry the service's name (`school-api`, `school-web`), version (the git commit)
and environment.

### Following an error, end to end

1. A user sees *"The server had a problem. Try again. (ref: 7bbf28dd…)"*. The frontend shows
   the trace ID from the `X-Trace-ID` response header, and **also reports the error to Faro**
   with that ID.
2. On the dashboard, **"Browser errors"** lists it; its **API trace** link opens the backend
   trace. Or paste the ID into **Explore → Tempo**.
3. The trace shows the whole path from the browser's fetch to the failing span and its error
   message, e.g. the SQL query that couldn't reach Postgres.
4. From the trace, **Logs for this span** opens the request's log lines in Loki; from a log
   line, the **Trace** link next to `trace_id` goes back.
5. **"Recent failed requests"** lists 5xx traces; the dots on latency charts are
   **exemplars**, linking a data point to a trace behind it.

### Alerts

Defined in `observability/grafana/provisioning/alerting/school-alerts.yaml`, visible in
**Alerting → Alert rules → School**:

| Alert | Fires when | Severity |
| --- | --- | --- |
| High error rate (5xx) | > 5% of requests fail, for 5 min | critical |
| Slow responses | p95 latency > 500 ms, for 10 min | warning |
| API down | Prometheus can't scrape the API (`up` = 0), for 1 min | critical |
| Database connection pool saturated | > 90% of connections in use, for 5 min | warning |

No contact point (email, Slack…) is set up, so they show in Grafana only. After editing the
file, `make -C observability restart`.

### Health checks

| Endpoint | Meaning | Use |
| --- | --- | --- |
| `GET /healthz` | The process is up and serving HTTP | Liveness: restart if it fails |
| `GET /readyz` | Postgres answers (Redis is reported, but optional) | Readiness: send traffic only when 200 |

`/readyz` returns 503 when Postgres is down, and `"degraded: running without the cache"`
(still 200) when only Redis is. Both are left out of traces and logs.

### Privacy

SQL spans record the statement but not its parameters, and Redis spans record the command
name but not its arguments (which hold cached students' names and emails). Faro records
page URLs and errors, not form contents.

## Frontend layout

The frontend is a **static single page app**, organised like MyApp's web project:
thin routes, feature components that fetch their own data, and a services layer that
talks to the API. `adapter-static` builds it into plain HTML, JS and CSS in
`frontend/build/`, which nginx or a CDN can serve. No Node process runs in production.

```
frontend/
├── nginx.conf                     production example: serves build/, forwards /api to Go and /collect to the collector
├── vite.config.ts                 adapter-static (fallback 200.html) + dev proxy for /api and /collect
└── src/
    ├── app.css                    Tailwind import + a few shared classes (.btn, .input, .card, …)
    ├── hooks.client.ts            starts Faro; reports unexpected page errors to it
    ├── lib/
    │   ├── observability/faro.ts  Grafana Faro: browser errors, Web Vitals, tracing of /api calls
    │   ├── services/              API calls only, no UI
    │   │   ├── shared.ts          request() → { success, message, data }, never throws
    │   │   ├── students/
    │   │   │   ├── index.ts       readStudents, readStudent, createStudent, updateStudent, deleteStudent
    │   │   │   └── types.ts       TS mirror of the Go DTOs
    │   │   ├── teachers/          same shape
    │   │   ├── courses/           same shape
    │   │   └── enrollments/       readEnrollments, enroll, withdraw
    │   ├── features/              UI, one folder per feature, files named after their component
    │   │   ├── dashboard/Dashboard.svelte
    │   │   ├── students/
    │   │   │   ├── StudentList.svelte     list + search + delete (fetches its own data)
    │   │   │   ├── NewStudent.svelte      create page
    │   │   │   ├── StudentDetails.svelte  edit + enrollments
    │   │   │   ├── StudentForm.svelte     the form, used by NewStudent and StudentDetails
    │   │   │   ├── StudentTable.svelte
    │   │   │   └── form.ts                FormData → StudentInput
    │   │   ├── teachers/          same shape (TeacherList, NewTeacher, …)
    │   │   ├── courses/           same shape (CourseList, NewCourse, …)
    │   │   └── enrollments/       EnrollForm.svelte, EnrollmentTable.svelte
    │   └── ui/
    │       ├── FormError.svelte   inline error for the current page
    │       ├── flash.svelte.ts    shared success message ("Student deleted"), see below
    │       └── Flash.svelte       shows it at the top of every page (in +layout.svelte)
    └── routes/                    one line each: render the feature component
        ├── +layout.ts             ssr = false: everything renders in the browser
        ├── +page.svelte           <Dashboard />
        ├── students/              <StudentList /> · new/ <NewStudent /> · [id]/ <StudentDetails id=… />
        ├── teachers/              same
        └── courses/               same
```

How the frontend talks to the backend:

- **Components fetch their own data.** A feature component calls a service function when
  it mounts and keeps the result in `$state` (`busy`, `error` and the data). List pages use
  `$effect` instead of `onMount` so they fetch again when `?q=` changes. Detail pages do
  the same when the `id` changes, because the router keeps the component mounted.
- **Services never throw.** Every call returns `{ success, message, data }`, like MyApp's
  `services/shared.ts`. A component checks `ret.success` and puts `ret.message` in
  `FormError`. The message is Go's text for a 4xx, "Could not reach the API" when Go is
  down, and a generic message for other 5xx errors.
- **Same origin, so no CORS.** The browser calls `/api/v1/...` on its own origin. In dev,
  Vite proxies `/api` to Go (`API_TARGET`). In production, nginx does the same
  (`nginx.conf`).
- **Forms need JavaScript.** `StudentForm` (and the other forms) stops the browser's own submit, parses the
  fields (`form.ts`) and calls the page's `onsave`. Tables call `ondelete` and `onwithdraw`.
- **Tailwind CSS 4** handles styling, mostly as utility classes in the markup.
- **Svelte 5 runes** are used throughout: `$props()`, `$state()`, `$derived()`, `$effect()`.
- **Imports** use `#lib/...` (Node subpath imports, defined in `package.json`). This is the
  SvelteKit 3 replacement for `$lib`.

### Reminder: who talks to whom

```
dev:   Browser  ⇄  Vite (5173)  ─/api→  Go API  ⇄  PostgreSQL
prod:  Browser  ⇄  nginx        ─/api→  Go API  ⇄  PostgreSQL
                (static files + JSON)
```

What happens on submit, using [NewCourse.svelte](frontend/src/lib/features/courses/NewCourse.svelte) as the example:

1. The user types. The text stays in the browser's input fields; nothing tracks keystrokes.
2. On submit, `CourseForm` calls `preventDefault()`, parses the fields and calls `onsave(input)`.
3. `NewCourse` calls `createCourse(input)` from `services/courses`, which sends JSON to Go.
4. **Success:** `goto('/courses/<id>')` opens the new course.
   **Failure:** `error = ret.message` and `FormError` shows it. Nothing re-renders the form,
   so what the user typed stays put.

On a details page, a successful save sets `saved` and fetches the record again.

**Success messages** use the one piece of shared state in the app,
[flash.svelte.ts](frontend/src/lib/ui/flash.svelte.ts). It is a `.svelte.ts` file because
the page that sets the message is often not the page that shows it: `StudentDetails`
deletes the student, runs `await goto('/students')`, then calls `showMessage(...)`, and the
list page shows "Student deleted". The layout clears the message on every navigation.

## Production build

```bash
cd frontend && make build     # writes build/ (200.html + _app/ + static files)
make preview                  # serves build/ locally with the same /api proxy
```

To deploy, copy `build/` and `nginx.conf` into an nginx container. Change `api:8080` in
`nginx.conf` to wherever the Go API runs.

## Ideas for next steps

- Add a feature end to end, e.g. `grades` (student × course × score): migration → model →
  repo → service → handler → routes on the backend, then service → feature components → route
  on the frontend.
- Write service tests with a fake `Repository`. The interfaces are there for that.
- Add pagination (`?page=&limit=`) to the list endpoints.
- Add auth: a Gin middleware that checks a token, plus sending that token from `lib/api.ts`.
- Move the capacity check into a transaction so concurrent enrollments can't overfill a course.
