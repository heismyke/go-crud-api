# go-crud-lab

A small library API (authors, books, members, loans) in Go + PostgreSQL,
**deliberately unfinished and broken**. You learn CRUD and SQL by fixing it.

- `database/sql` with the `pgx` driver, hand-written SQL (no ORM)
- Go 1.22+ `net/http` routing (`GET /books/{id}`), no framework
- Integration tests against a real Postgres that go from red to green as you fix things

## What's in here

```
cmd/api/main.go              wiring + HTTP server
internal/db/db.go            opening the connection pool
internal/models/             plain structs
internal/store/
  author_store.go            ✅ REFERENCE — correct, heavily commented. Read first.
  book_store.go              🐛 looks plausible, has many bugs
  loan_store.go              🐛 + 🚧 one buggy method, three to write
  member_store.go            🚧 write it from scratch
  *_test.go                  the tests that tell you when you're done
internal/handlers/           HTTP layer (authors = reference; books has bugs too)
migrations/                  schema + seed data (the schema has weaknesses on purpose)
sql/exercises.sql            ~35 psql exercises, basic → locking
requests.http                clickable requests for every endpoint
EXERCISES.md                 your task list: symptoms, not locations
HINTS.md                     progressive hints, collapsed, peek only when stuck
```

## Setup (PowerShell)

```powershell
docker compose up -d            # dev DB on :5440 (seeded), test DB on :5441
go run ./cmd/api                # API on http://localhost:8080
```

Then in another terminal:

```powershell
curl.exe localhost:8080/health
curl.exe localhost:8080/authors
go test ./...                   # expect lots of FAILs, that's the point
```

psql (it's inside the container, so no local install needed):

```powershell
docker compose exec db psql -U library -d library
```

Reset the dev DB to the seed data whenever you've wrecked it:

```powershell
docker compose down -v; docker compose up -d
```

## Workflow that works well

1. `git init; git add .; git commit -m "starting point"` so `git diff` shows exactly what you changed.
2. Read `author_store.go` top to bottom. Every pattern you need is in there once.
3. Pick a symptom from `EXERCISES.md`, reproduce it (requests.http or psql), find it, fix it.
4. Run the matching test: `go test ./internal/store -run TestBookStore_Get -v`
5. Stuck for 20+ minutes? Open one hint in `HINTS.md`, not all of them.

Useful: watch every query the app sends to Postgres —

```powershell
docker compose exec db psql -U library -d library -c "ALTER SYSTEM SET log_statement = 'all'" -c "SELECT pg_reload_conf()"
docker compose logs -f db
```
