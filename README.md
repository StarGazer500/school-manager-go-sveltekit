# School Manager: Go (Gin) + SvelteKit

A small school management app for learning Gin and SvelteKit. It manages
**students**, **teachers**, **courses** and **enrollments**. Both halves use a
feature-based layout: code is grouped by business feature, not by technical layer.

```
gosvelte/
├── backend/    Go + Gin + database/sql + SQLite → http://localhost:8080
└── frontend/   SvelteKit 3 + Svelte 5     → http://localhost:5173
```

## Running it

With `make`, from the project root:

```bash
make dev        # backend on :8080 + frontend on :5173 (Ctrl+C stops both)
make check      # go vet/test + svelte-check
```

Each folder has its own Makefile too. Run `make` in `backend/` or `frontend/` to list
the targets (`run`, `build`, `reset-db`, `dev`, `check`, …). The frontend Makefile
switches to the Node version in `.nvmrc` automatically.

Without `make`:

**Backend** (Go 1.23+). This creates `school.db` on first run:

```bash
cd backend
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

| Where    | Variable  | Default                        |
| -------- | --------- | ------------------------------ |
| backend  | `PORT`    | `8080`                         |
| backend  | `DB_PATH` | `school.db`                    |
| frontend | `API_URL` | `http://localhost:8080/api/v1` |

## Backend layout

```
backend/
├── cmd/api/main.go              entrypoint: config → DB → migrate → server → graceful shutdown
└── internal/
    ├── config/                  env vars
    ├── database/                open SQLite, run schemas, map SQLite errors → app errors
    ├── server/server.go         composition root: wires every feature together
    ├── shared/
    │   ├── apperr/              error kinds: NotFound / Conflict / Invalid
    │   └── httpx/               Gin helpers: error → HTTP status, ID parsing, validation messages
    └── features/
        ├── students/
        │   ├── schema.go        CREATE TABLE owned by this feature
        │   ├── model.go         DB entity (+ list filter)
        │   ├── dto.go           JSON request/response shapes + mapping
        │   ├── repository.go    interface + plain SQL implementation
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
- **Repository**: SQL only. It returns `apperr` errors so services never see SQL or driver errors.
- **Errors**: services return `apperr.NotFound(...)` and similar, and `httpx.Error` turns them into 404/409/400.
  Unknown errors become 500 and are logged.

Things worth reading in the code:

- **Cross-feature dependency.** `enrollments` needs to look up courses. It defines a tiny
  `CourseReader` interface on *its* side, and `courses.Service` happens to satisfy it.
  The two features stay loosely coupled.
- **The database enforces integrity.** Foreign keys are real (`PRAGMA foreign_keys`).
  Deleting a teacher sets their courses' `teacher_id` to NULL, and deleting a student or
  course cascades to its enrollments.
- **Read models.** `enrollments.Detail` is a JOIN result (student name, course code). It is
  not a table.

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

```
frontend/src/
├── env.ts                         declares API_URL (read via `$app/env/private`)
├── app.css                        Tailwind import + a few shared classes (.btn, .input, .card, …)
├── lib/
│   ├── server/api.ts              server-only fetch wrapper, ApiError, form-failure helpers
│   ├── ui/                        shared, feature-agnostic components
│   └── features/
│       ├── students/
│       │   ├── types.ts           TS mirror of the Go DTOs
│       │   ├── api.server.ts      calls to /students (server-only)
│       │   ├── form.ts            FormData → StudentInput
│       │   └── components/        StudentForm, StudentTable
│       ├── teachers/              same shape
│       ├── courses/               same shape
│       └── enrollments/           api + shared enroll/withdraw actions + components
└── routes/                        thin pages that load data and delegate to features
    ├── +page(.server).ts/.svelte  dashboard
    ├── students/                  list · new/ · [id]/ (edit + enrollments)
    ├── teachers/                  list · new/ · [id]/ (edit + courses taught)
    └── courses/                   list · new/ · [id]/ (edit + roster)
```

How the frontend talks to the backend:

- **Only the SvelteKit server talks to Go.** `load` functions in `+page.server.ts` fetch data,
  and **form actions** handle writes. The browser never calls the Go API directly, so there
  is no CORS to configure and `API_URL` stays private. `*.server.ts` files and anything under
  `/server/` can't be imported into browser code.
- **Forms work without JavaScript.** They are plain `<form method="POST">`. `use:enhance`
  upgrades them to fetch requests without a page reload when JS is available.
- **Errors.** A 4xx from Go becomes `fail(status, { error, values })`, so the page shows the
  message and re-fills the form. A 404 in `load` shows the 404 page. 5xx or a backend that
  is down shows `+error.svelte`.
- **Tailwind CSS 4** handles styling, mostly as utility classes in the markup. The forms
  plugin gives inputs a clean default look.
- **Svelte 5 runes** are used throughout: `$props()`, `$derived()`.
- **Imports** use `#lib/...` (Node subpath imports, defined in `package.json`). This is the
  SvelteKit 3 replacement for `$lib`.

### Reminder: who talks to whom

```
Browser  ⇄  SvelteKit server  ⇄  Go API  ⇄  SQLite
         (HTML, form posts)   (JSON over HTTP)
```

The browser only talks to SvelteKit, and only SvelteKit talks to Go. SvelteKit acts as
the frontend's own backend.

**`data` vs `form` on a page**, using [courses/new](frontend/src/routes/courses/new/+page.svelte) as the example:

```svelte
<CourseForm values={form?.values} teachers={data.teachers} />
```

- **`data`** is what the server gives you when the page *loads*. `load()` in
  `+page.server.ts` fetches teachers from Go, and they arrive as `data.teachers`
  (the dropdown options). `load()` re-runs after every form submit.
- **`form`** is what the server gives back after you *submit*. It is `undefined` until a
  submit happens, and it only has content if the submit failed.

What happens on submit:

1. The user types. The text stays in the browser's input fields; nothing tracks keystrokes.
2. On submit, the browser sends the fields as form data to the SvelteKit action.
3. The action parses it (`parseCourseForm`) and sends it as JSON to Go (`api()` in `lib/server/api.ts`).
4. **Success:** `redirect(303, ...)` goes to the new course's page.
   **Failure:** `failFromApi` returns `fail(status, { error, values })`.
5. The page re-renders with that object as `form`. `FormError` shows `form.error`, and
   `CourseForm` refills its inputs from `form.values`, so the user doesn't have to retype.

`values` is just a name chosen here, not a SvelteKit keyword. On edit pages,
`values={form?.values ?? data.course}` means "show the unsaved input if a save failed,
otherwise show the saved record."

## Ideas for next steps

- Add a feature end to end, e.g. `grades` (student × course × score): schema → model →
  repo → service → handler → routes on the backend, then types → api → components → route
  on the frontend.
- Write service tests with a fake `Repository`. The interfaces are there for that.
- Add pagination (`?page=&limit=`) to the list endpoints.
- Add auth: a Gin middleware plus SvelteKit `hooks.server.ts` with `locals.user`.
- Move the capacity check into a transaction so concurrent enrollments can't overfill a course.
