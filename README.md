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
asks for, automatically), **Node 22.17+** and **Docker** (for PostgreSQL).

## Running it

With `make`, from the project root:

```bash
make dev        # Postgres (Docker) + backend on :8080 + frontend on :5173 (Ctrl+C stops the servers)
make check      # go vet/test + svelte-check
```

Each folder has its own Makefile too. Run `make` in `backend/` or `frontend/` to list
the targets (`run`, `db-up`, `db-down`, `db-reset`, `db-shell`, `migrate-*`, `dev`, `check`, …). The
frontend Makefile switches to the Node version in `.nvmrc` automatically. Postgres keeps
running after Ctrl+C; stop it with `make -C backend db-down`.

> **pgrepo is a private module.** Go can only download it with
> `GOPRIVATE=github.com/StarGazer500` set (and SSH access to the repo). `backend/Makefile`
> sets this for you. If you run `go` commands directly, export it first, or run
> `go env -w GOPRIVATE=github.com/StarGazer500` once.

Without `make`:

**Backend.** The API applies the database migrations itself when it starts:

```bash
cd backend
docker compose up -d --wait db              # PostgreSQL on localhost:5433
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
| frontend | `API_TARGET`   | `http://localhost:8080` (where the dev proxy sends `/api`) |

The database runs from [backend/docker-compose.yml](backend/docker-compose.yml) on port
**5433**, so it doesn't clash with another Postgres on the default 5432. Its data lives in
a Docker volume and survives restarts; `make -C backend db-reset` deletes it.

## Backend layout

```
backend/
├── docker-compose.yml           PostgreSQL 18 for local development
├── migrations/                  numbered SQL files (goose), embedded into the binary
├── cmd/
│   ├── api/main.go              entrypoint: config → DB → migrate → server → graceful shutdown
│   └── migrate/main.go          `make migrate-*` commands: up, down, status, new
└── internal/
    ├── config/                  env vars
    ├── database/                open the pgx pool, run migrations, map Postgres errors → app errors
    ├── server/server.go         composition root: wires every feature together
    ├── shared/
    │   ├── apperr/              error kinds: NotFound / Conflict / Invalid
    │   └── httpx/               Gin helpers: error → HTTP status, ID parsing, validation messages
    └── features/
        ├── students/
        │   ├── model.go         DB entity with `db` tags + TableName() (+ list filter)
        │   ├── dto.go           JSON request/response shapes + mapping
        │   ├── repository.go    interface + pgrepo implementation
        │   ├── service.go       business rules (no HTTP, no SQL)
        │   ├── handler.go       HTTP ↔ service
        │   └── routes.go        mounts endpoints
        ├── teachers/            same shape
        ├── courses/             same shape (+ optional teacher FK)
        └── enrollments/         same shape (+ depends on courses)
```

A request flows **router → handler → service → repository → DB**, and each layer has one job:

- **Handler**: parse/validate input (`ShouldBindJSON` + `binding:` tags), call the service, write JSON.
- **Service**: business rules, e.g. "a course can't take more students than its capacity"
  ([enrollments/service.go](backend/internal/features/enrollments/service.go)).
- **Repository**: database access only, through [pgrepo](https://github.com/StarGazer500/pgrepo),
  which writes the SQL from each model's `db` tags. It returns `apperr` errors, so services
  never see SQL or driver errors.
- **Errors**: services return `apperr.NotFound(...)` and similar, and `httpx.Error` turns them into 404/409/400.
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
