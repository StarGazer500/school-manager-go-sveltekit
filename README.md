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
| frontend | `API_TARGET` | `http://localhost:8080` (where the dev proxy sends `/api`) |

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
dev:   Browser  ⇄  Vite (5173)  ─/api→  Go API  ⇄  SQLite
prod:  Browser  ⇄  nginx        ─/api→  Go API  ⇄  SQLite
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

- Add a feature end to end, e.g. `grades` (student × course × score): schema → model →
  repo → service → handler → routes on the backend, then service → feature components → route
  on the frontend.
- Write service tests with a fake `Repository`. The interfaces are there for that.
- Add pagination (`?page=&limit=`) to the list endpoints.
- Add auth: a Gin middleware that checks a token, plus sending that token from `lib/api.ts`.
- Move the capacity check into a transaction so concurrent enrollments can't overfill a course.
