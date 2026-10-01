# Exercises

Symptoms only. Finding *where* is half the skill. Each item has a test that
proves it (named in brackets) unless noted. Numbers match `HINTS.md`.

Heads-up: bugs overlap. Fixing one can *reveal* another that was hidden behind
it. That's normal in real codebases. Re-run the tests after each fix.

---

## Level 0: warm-up (no code changes)

- [ ] Get the API running and `GET /authors` returning data.
- [ ] `go test ./...`. Confirm the two `TestAuthorStore_*` tests pass. If they
      don't, your setup is broken, not the code.
- [ ] Read `author_store.go`. Without looking back, answer:
  - When do you use `QueryContext` vs `QueryRowContext` vs `ExecContext`?
  - Why `defer rows.Close()`? Why check `rows.Err()` after the loop?
  - How does a NULL `bio` survive `Scan`?
  - What's the difference between `%w` and `%v` in `fmt.Errorf`, and who cares?
  - `Update` and `Delete` detect "no such id" in two different ways. What are they?
- [ ] Do Levels 1–3 of `sql/exercises.sql` in psql.

---

## Level 1: bug hunt in the books API

1. **`GET /books` returns `null`**, even though there are 14 books in the DB.
   `GET /books?page=-1` is a 500.  [`TestBookStore_List_Pagination`]
2. **Once #1 is fixed, `GET /books` is a 500.** Read the server log.  [`TestBookStore_Get_NullPublishedYear`]
   Your fix must survive a round trip: `GET` a book with no year, `PUT` it back unchanged, and the year must still be NULL, not `0`.
3. **An empty books table comes back as `null`, not `[]`.**  [`TestBookStore_List_EmptyIsNotNil`]
4. **`List` leaks a database connection when a row fails to scan.** You can't see it
   through HTTP; work out *why not* (hint: context). A background job using
   `context.Background()` would slowly eat the pool until the app hangs.
   [`TestBookStore_List_ReleasesConnectionOnError`]
5. **Paging isn't stable.** Books inserted in the same transaction have identical
   `created_at`. Can the same book show up on page 1 and page 2?  [`TestBookStore_List_Pagination`]
6. **`GET /books/1`: title and ISBN are swapped.**  [`TestBookStore_Get_ReturnsCorrectFields`]
7. **`GET /books/9999` is a 500, should be 404.** The log says "no rows". Two
   separate things are wrong.  [`TestBookStore_Get_MissingIsErrNotFound`]
8. **`POST /books` responds with `"id": 0`** and zero timestamps, though the row is
   in the DB.  [`TestBookStore_Create_FillsGeneratedFields`]
9. **`POST /books` with an unknown `author_id` is a 500**, should be 404.  [`TestBookStore_Create_UnknownAuthorIsNotFound`]
10. **`POST /books` with body `this is not json` creates a blank book.** Also: nothing
    stops `total_copies: -5` or an empty title. (No test. Fix it in the handler,
    following `authors.go`.)
11. **`PUT /books/6` ...** now look at every other book. 😬 Reset the DB afterwards.
    Also `updated_at` never changes.  [`TestBookStore_Update_OnlyChangesTargetRow`]
12. **`PUT /books/9999` returns 200.**  [`TestBookStore_Update_MissingIsErrNotFound`]
13. **`DELETE /books/9999` returns 204.**  [`TestBookStore_Delete_MissingIsErrNotFound`]
14. **`DELETE /books/1` is a 500**, should be 409.  [`TestBookStore_Delete_WithLoansIsConflict`]
15. **Search.** Try `q=' OR '1'='1' --` (URL-encoded in requests.http). Then think about
    what `q='; DROP TABLE loans; --` could do.  [`TestBookStore_Search_IsNotInjectable`]
16. **`GET /authors/1/stats` is a 500.** Read the error text closely.  [`TestBookStore_CountByAuthor`]
17. **`GET /books/with-authors` works, but count its queries** (see README: `log_statement`).
    With 2000 books, how many round trips? Rewrite it as one query.
    Compare `go test ./internal/store -run ListWithAuthors_Timing -v` before and after.
18. **Updating `total_copies` leaves `available_copies` alone.** Change a book from
    3 copies to 1 while 2 are on loan. What should happen? (Design question: decide, then implement.)

## Level 2: loans (transactions)

19. **`POST /loans/1/return` returns 200, but:** the loan still shows `returned_at: null`
    in psql, `available_copies` went *up*, and you can return the same loan again and again.
    Three bugs in one short function.
    [`TestLoanStore_Return_MarksReturnedAndRestocks`, `TestLoanStore_Return_TwiceIsAlreadyReturned`, `TestLoanStore_Return_MissingIsErrNotFound`]

## Level 3: build it yourself

20. **`MemberStore`**: all five methods, from scratch. [`TestMemberStore_*`]
    Then add input validation to `handlers/members.go`.
21. **`LoanStore.Checkout`**: a real transaction with row locking. [`TestLoanStore_Checkout_*`]
    First make the simple tests pass. Then look at `_ConcurrentLastCopy`: 10 goroutines
    fight over 1 copy. Do exercise 7.2 in `sql/exercises.sql` first.
    Once it passes, remove `FOR UPDATE` and run that test 10 times (`-count=10`). What happens?
22. **`LoanStore.ListOverdue`**: one JOIN query. [`TestLoanStore_ListOverdue`]
23. **`LoanStore.ListByMember`**: with an optional filter, *without* building SQL by
    string concatenation. [`TestLoanStore_ListByMember`]

## Level 4: fix the schema

Create `migrations/003_constraints.sql` and apply it:
`Get-Content migrations/003_constraints.sql | docker compose exec -T db psql -U library -d library`
(Also add it to the `db_test` volumes in `docker-compose.yml` so tests use it.)

24. **ISBNs aren't unique.** Add a UNIQUE constraint. It'll fail on the seed data.
    Find the duplicates with SQL, decide what to do about them, then add it.
    Map the violation to `ErrConflict` in `BookStore`.
25. **Nothing stops `available_copies` going negative or above `total_copies`.** Add CHECK
    constraints. Now `Checkout` has a second line of defence. Try to break it in psql.
26. **Foreign keys have no indexes.** Postgres indexes primary keys and UNIQUE columns,
    not foreign keys. Add them. Use `EXPLAIN ANALYZE` (sql Level 6) to show the difference.
27. **`updated_at` depends on every UPDATE remembering to set it.** Write a trigger that sets it.
28. **A member can't hold two copies of the same book at once**. Enforce that with a
    *partial* unique index (`... WHERE returned_at IS NULL`).
29. **`loans.due_at` could be before `borrowed_at`.** Should the DB forbid it?

## Level 5: stretch

- `PATCH /books/{id}` updating only the fields sent (look at `COALESCE`, or build the SET
  list safely).
- `GET /books?author_id=2&year_from=1980&sort=title`: dynamic filters with placeholders
  only, and an allow-list for `sort`.
- Cursor pagination (`?after=<id>`) instead of OFFSET. Why is OFFSET slow on page 10,000?
- Soft delete (`deleted_at`). What happens to every query you've written?
- `GET /stats`: most borrowed books, members with most overdue loans (window functions).
- Handler tests with `net/http/httptest`.
- Port one store to native `pgx` (`pgxpool`, `pgx.CollectRows`) and compare.
- Put the migrations under a real tool (golang-migrate or goose) with up/down files.
