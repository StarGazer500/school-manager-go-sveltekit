# School Manager: Go (Gin) + SvelteKit

A small school management app for learning Gin and SvelteKit. It manages
**students**, **teachers**, **courses** and **enrollments**. Both halves use a
feature-based layout: code is grouped by business feature, not by technical layer.

```
gosvelte/
├── backend/    Go + Gin + pgrepo + PostgreSQL → http://localhost:8080
└── frontend/   SvelteKit 3 + Svelte 5         → http://localhost:5173
```

You need **Go** (any version from 1.21 up: it downloads Go 1.27.3, which `backend/go.mod`
asks for, automatically), **Node 22.17+** and **Docker** (for PostgreSQL and Redis).

## Running it

With `make`, from the project root:

```bash
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
| backend  | `LOG_FORMAT`   | `console` (readable, colored) or `json` (one JSON object per line, for production) |
| backend  | `LOG_LEVEL`    | `info` (also `debug`, `warn`, `error`)                     |
| frontend | `API_TARGET`   | `http://localhost:8080` (where the dev proxy sends `/api`) |

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
│   ├── api/main.go              entrypoint: config → DB → migrate → server → graceful shutdown
│   └── migrate/main.go          `make migrate-*` commands: up, down, status, new
└── internal/
    ├── core/                    one package (`core`): infrastructure set up at startup
    │   ├── config.go            env vars (LoadConfig)
    │   ├── database.go          pgx pool (OpenDatabase), Postgres errors → app errors (MapError)
    │   ├── migrate.go           goose migrations (Migrate, Migrator)
    │   └── cache.go             Redis (OpenRedis), caches for services (NewCache[T]), not-found helpers
    ├── logger/                  builds the zerolog logger (console or JSON)
    ├── server/server.go         composition root: middleware + wires every feature together
    ├── shared/                  one package (`shared`) used by every feature:
    │   ├── commonutils.go       error kinds (NotFound / Conflict / Invalid, type AppError) +
    │   │                        Gin helpers: errors → HTTP status, ID parsing, validation messages
    │   └── middleware.go        request logging + panic recovery
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
        └── enrollments/         same shape, no cache (+ depends on courses)
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
16:18:55 INF request method=GET path=/api/v1/students status=200 latency=13.6 request_id=c6e5c72a9ef7b5de
16:18:55 WRN request method=POST path=/api/v1/students status=409 latency=5.1 request_id=e6365961ff72870f
```

With `LOG_FORMAT=json`, each line is a JSON object instead, which log tools can filter by
field (e.g. every `status` ≥ 500).

- **One line per request** from `shared.RequestLogger`: `INF` for success, `WRN` for 4xx,
  `ERR` for 5xx. `latency` is in milliseconds.
- **Every request gets an ID.** It's in each log line for that request and in the
  `X-Request-ID` response header, so an error seen in the browser's network tab can be
  found in the logs. If a proxy already sent an `X-Request-ID`, that one is reused.
- **How code gets the logger.** There's no global logger. The rule is *injected at startup,
  carried in `ctx` during a request*:
  - **Outside a request** (startup, migrations, Redis, the caches' error handler), `main.go`
    passes the logger as a parameter: `core.Migrate(ctx, db, log)`,
    `core.OpenRedis(ctx, url, log)`, `core.NewCache[T](rdb, prefix, log)`, `server.New(db, rdb, log)`.
  - **Inside a request**, the middleware puts a logger with the request ID into the
    request's context. Get it with `zerolog.Ctx(ctx)`, and your line carries the same
    `request_id`:

    ```go
    zerolog.Ctx(ctx).Info().Int64("course_id", id).Msg("course is full")
    ```

  Outside a request `zerolog.Ctx(ctx)` returns a logger that writes **nothing**, so pass
  a logger in there instead.
- **go-redis's own messages** are sent through zerolog at debug level (`source=go-redis`),
  so they only show with `LOG_LEVEL=debug` and JSON output stays valid.

- **Unexpected errors** (anything that isn't a `shared.AppError`) are logged by `shared.ErrorResponse` with
  the real error, while the client only sees `internal server error`. **Panics** are
  caught by `shared.Recoverer`, logged with a stack trace, and answered with a 500.

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
| GET                     | `/health`                                |                              |

```bash
curl -X POST localhost:8080/api/v1/students \
  -H 'content-type: application/json' \
  -d '{"firstName":"Ada","lastName":"Lovelace","email":"ada@school.test","gradeLevel":10}'
```

## Frontend layout

The frontend is a **static single page app**, organised like MyApp's web project:
thin routes, feature components that fetch their own data, and a services layer that
talks to the API. `adapter-static` builds it into plain HTML, JS and CSS in
`frontend/build/`, which nginx or a CDN can serve. No Node process runs in production.

```
frontend/
├── nginx.conf                     production example: serves build/, forwards /api to Go
├── vite.config.ts                 adapter-static (fallback 200.html) + dev proxy for /api
└── src/
    ├── app.css                    Tailwind import + a few shared classes (.btn, .input, .card, …)
    ├── lib/
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
