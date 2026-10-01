# Hints

Open **one** hint at a time. Each item goes from a gentle hint to a strong one.
There are no full solutions here on purpose.

<details><summary>1. <code>GET /books</code> returns null</summary>

- Hint 1: The handler defaults `page` to 1. What OFFSET does page 1 need?
- Hint 2: Write out the offsets for pages 1, 2, 3 with limit 20. Then look at how `offset` is computed.
- Hint 3: Negative page → negative OFFSET → Postgres error. Clamp `page` and `limit` in the handler (and cap `limit`).
</details>

<details><summary>2. NULL published_year</summary>

- Hint 1: Read the whole log line. Which column, and what Go type is it scanning into?
- Hint 2: `AuthorStore` handles a nullable column. Which field, and what type is it?
- Hint 3: Change the model field's type. The compiler will then show you every place that has to follow, including the JSON input struct. That's why the PUT round trip works afterwards.
</details>

<details><summary>3. <code>null</code> instead of <code>[]</code></summary>

- Hint: `var xs []T` vs `xs := []T{}`. Compare with `AuthorStore.List`.
</details>

<details><summary>4. Connection leak</summary>

- Hint 1: What must happen to `rows` on *every* path out of the function, including the `return` inside the loop?
- Hint 2: Compare the first lines after `QueryContext` in `AuthorStore.List`. And what does `AuthorStore.List` do after the loop that this one doesn't?
- Why HTTP hides it: `database/sql` watches the query's `ctx`. When the handler returns, the request context is cancelled and `database/sql` closes the rows for you. `t.Context()` in the test lives until the test ends, so the leak is visible there.
</details>

<details><summary>5. Unstable paging</summary>

- Hint: If two rows tie on the ORDER BY column, SQL doesn't promise any order between them. Add a unique column as a tie-breaker.
</details>

<details><summary>6. Title/ISBN swapped</summary>

- Hint: Put the SELECT column list and the Scan argument list side by side and read them in pairs. Scan matches by *position*, not by name.
</details>

<details><summary>7. 404 becomes 500</summary>

- Hint 1: What sentinel does `AuthorStore.Get` check for, and what does it return instead?
- Hint 2: Even after that, `errors.Is` can't see through one of the `fmt.Errorf` verbs. Which one?
</details>

<details><summary>8. Create returns id 0</summary>

- Hint 1: The SQL says `RETURNING`. Does anything read what's returned?
- Hint 2: `ExecContext` throws result rows away. Compare with `AuthorStore.Create`.
</details>

<details><summary>9. Unknown author → 500</summary>

- Hint: Which Postgres error code is it? `errors.go` already has a helper. `AuthorStore.Delete` shows how to use it.
</details>

<details><summary>10. Garbage body creates a blank book</summary>

- Hint: `decodeJSON` returns an error. Who looks at it? Then copy the `validate()` pattern from `authors.go`.
</details>

<details><summary>11. PUT renames every book</summary>

- Hint 1: Count the placeholders in the SQL and the arguments passed. Is `id` used anywhere?
- Hint 2: An UPDATE with no WHERE clause updates *every row*, and Postgres won't warn you. Top tip: in psql, always write the WHERE clause first, or run it inside `BEGIN; ... ROLLBACK;`.
- Hint 3: `updated_at = now()` in the SET list (or the trigger from exercise 27).
</details>

<details><summary>12–13. Update/Delete of a missing id succeed</summary>

- Hint: `AuthorStore.Update` uses `RETURNING` + `sql.ErrNoRows`, `AuthorStore.Delete` uses `RowsAffected()`. Pick one for each.
</details>

<details><summary>14. Delete book with loans → 500</summary>

- Hint: Same family as #9. Then decide: should deleting a book delete its loan history (`ON DELETE CASCADE`)? The test says no. Why is that the better choice for a library?
</details>

<details><summary>15. SQL injection</summary>

- Hint 1: Print the final `query` string for `q = ' OR '1'='1' --`. Paste it into psql.
- Hint 2: Never put user input into SQL text with `Sprintf`. Pass it as a `$1` argument.
- Hint 3: The `%` wildcards have to live somewhere. Either concatenate them *in SQL* (`'%' || $1 || '%'`) or add them to the Go string you pass as the argument.
- Stretch: what does searching for `%` or `_` do now? Look up `ESCAPE` in LIKE.
</details>

<details><summary>16. Stats 500</summary>

- Hint: `?` is MySQL/SQLite placeholder syntax. Postgres uses `$1`, `$2`, …
</details>

<details><summary>17. N+1 queries</summary>

- Hint 1: 1 query for books + 1 per book for its author = N+1.
- Hint 2: `JOIN authors a ON a.id = b.author_id` and scan `a.name` as one more column.
- Hint 3: Once it's a JOIN, `BookStore` doesn't need the `authors` field any more.
</details>

<details><summary>18. total_copies vs available_copies</summary>

- Hint: You can do arithmetic in SET: `available_copies = available_copies + ($4 - total_copies)`. In an UPDATE, the right-hand side always sees the *old* row values. What if that goes negative? (Exercise 25 is the safety net.)
</details>

<details><summary>19. Return</summary>

- Hint 1: The function begins a transaction. Where is it committed? What does the deferred `Rollback` do then?
- Hint 2: Look at which handle the second UPDATE runs on: `tx` or `s.db`? Which transaction is it part of?
- Hint 3: `WHERE id = $1 AND returned_at IS NULL`. Then "no row" means *either* missing *or* already returned. One more small query inside the tx tells you which.
</details>

<details><summary>20. MemberStore</summary>

- Hint: It's `AuthorStore` with different columns. Duplicate email → `isUniqueViolation`. Delete with loans → `isForeignKeyViolation`.
</details>

<details><summary>21. Checkout</summary>

- Hint 1: `tx, err := s.db.BeginTx(ctx, nil)`, then `defer tx.Rollback()`, then *every* query goes through `tx`, then `return loan, tx.Commit()`.
- Hint 2: `SELECT available_copies FROM books WHERE id = $1 FOR UPDATE`. The second concurrent transaction blocks on this line until the first commits, then reads the *new* value.
- Hint 3: Unknown member: check with `SELECT EXISTS (...)`, or let the INSERT fail with a FK violation and map it.
- Hint 4: `due_at`: compute it in SQL (`now() + interval '14 days'`) or pass a Go `time.Time`. Either works. Which clock is the source of truth?
- Alternative without FOR UPDATE: `UPDATE books SET available_copies = available_copies - 1 WHERE id = $1 AND available_copies > 0 RETURNING ...`. An atomic conditional update. Why does that also work?
</details>

<details><summary>22. ListOverdue</summary>

- Hint 1: `loans JOIN books JOIN members`, `WHERE returned_at IS NULL AND due_at < now()`.
- Hint 2: `now() - due_at` is an `interval`. `extract(day FROM ...)` gives whole days. Cast to `int` so it scans into a Go int.
</details>

<details><summary>23. ListByMember</summary>

- Hint 1: One fixed query can handle the optional filter: `WHERE member_id = $1 AND (NOT $2 OR returned_at IS NULL)`.
- Hint 2: An empty result can't tell you whether the member exists. Check first.
</details>
