-- ============================================================================
-- psql practice. Run against the DEV database (it has seed data):
--
--   docker compose exec db psql -U library -d library
--
-- Inside psql you can run this whole file with:  \i /sql/exercises.sql
-- but it's better to copy one exercise at a time and write your answer under it.
--
-- psql survival kit:
--   \dt              list tables            \d books       describe a table
--   \x auto          wide rows go vertical  \timing on     show query time
--   \e               edit last query in $EDITOR
--   \q               quit
--   Every statement ends with ;   (no ; = psql waits for more input)
--
-- Wrecked the data? Reset everything:
--   docker compose down -v; docker compose up -d
-- ============================================================================


-- ---------------------------------------------------------------------------
-- Level 1: SELECT, WHERE, ORDER BY, LIMIT
-- ---------------------------------------------------------------------------

-- 1.1  All books published before 1980, oldest first. Show title and year.

-- 1.2  Books with no published_year. (Hint: `= NULL` does not work. Why?)

-- 1.3  The 3 most recent books by published_year. Where do NULL years end up
--      with ORDER BY ... DESC? Make them go last.

-- 1.4  Books whose title contains "the" in any case.

-- 1.5  Members who joined today, and who've never borrowed anything.
--      (Skip the second part until Level 3 if it's hard.)


-- ---------------------------------------------------------------------------
-- Level 2: aggregates, GROUP BY, HAVING
-- ---------------------------------------------------------------------------

-- 2.1  How many books are in the library? How many physical copies?

-- 2.2  Number of books per author_id.

-- 2.3  Same, but only authors with more than 2 books.
--      What's the difference between WHERE and HAVING?

-- 2.4  For each book: total_copies - available_copies = copies out on loan.
--      Then check it against the loans table. Do they agree? (They should.)

-- 2.5  Find duplicate ISBNs: each ISBN that appears more than once, and how many times.


-- ---------------------------------------------------------------------------
-- Level 3: JOINs
-- ---------------------------------------------------------------------------

-- 3.1  Every book with its author's name (INNER JOIN).

-- 3.2  Every author with how many books they have — INCLUDING authors with 0.
--      Does count(*) or count(b.id) give the right answer for 'Anonymous'? Why?

-- 3.3  Active loans: member name, book title, due date. Soonest due first.

-- 3.4  Overdue loans with days overdue. This is the query LoanStore.ListOverdue needs.

-- 3.5  Members who have never borrowed anything. Write it three ways:
--        a) LEFT JOIN ... WHERE ... IS NULL
--        b) NOT EXISTS
--        c) NOT IN          (then read about why NOT IN + NULLs is a trap)

-- 3.6  Books that have never been borrowed.


-- ---------------------------------------------------------------------------
-- Level 4: subqueries, CTEs, window functions
-- ---------------------------------------------------------------------------

-- 4.1  Books whose total_copies is above the average total_copies.

-- 4.2  Each member's most recent loan (one row per member).
--      Try it with DISTINCT ON, then with row_number() OVER (PARTITION BY ...).

-- 4.3  Rank authors by number of loans of their books (all time), with rank().

-- 4.4  Using a CTE (WITH ...): per author, their number of books AND their number of
--      active loans, in one result.

-- 4.5  For each book, its published_year and the previous book's year by the same
--      author (lag()).


-- ---------------------------------------------------------------------------
-- Level 5: writes. Wrap each in BEGIN; ... ROLLBACK; while experimenting!
-- ---------------------------------------------------------------------------

-- 5.1  INSERT a new author and a book for them in one go, using the id from
--      INSERT ... RETURNING. (Try a CTE: WITH a AS (INSERT ... RETURNING id) INSERT ...)

-- 5.2  UPSERT: insert a member with email ada@example.com; on conflict, update
--      their full_name instead. (INSERT ... ON CONFLICT (email) DO UPDATE ...)

-- 5.3  Mark every overdue loan older than 30 days as returned, and see which
--      rows changed with RETURNING. Then ROLLBACK.

-- 5.4  UPDATE ... FROM: set available_copies on every book to
--      total_copies - (number of active loans for that book).
--      This "recomputes" the counter from the source of truth.

-- 5.5  DELETE the author 'Anonymous'. Then try to delete 'Terry Pratchett'.
--      Read the error. Which constraint stopped you?


-- ---------------------------------------------------------------------------
-- Level 6: indexes & EXPLAIN
-- ---------------------------------------------------------------------------

-- 6.1  Make the tables bigger so the planner has something to think about:
--        INSERT INTO books (author_id, title, isbn, published_year)
--        SELECT (random()*5)::int + 1, 'Filler ' || n, 'fill-' || n, 1900 + (random()*120)::int
--        FROM generate_series(1, 200000) AS n;
--        ANALYZE books;

-- 6.2  EXPLAIN ANALYZE SELECT * FROM books WHERE author_id = 3;
--      Note the plan type and time. Now:  CREATE INDEX ON books (author_id);
--      Run EXPLAIN ANALYZE again. What changed?

-- 6.3  EXPLAIN ANALYZE SELECT * FROM books WHERE title ILIKE '%gods%';
--      Does a plain btree index on title help? Why not?
--      (Stretch: look up the pg_trgm extension and a GIN index.)

-- 6.4  \di  — list indexes. Which ones did Postgres create for you, and why?


-- ---------------------------------------------------------------------------
-- Level 7: transactions & locking (needs TWO psql windows: A and B)
-- ---------------------------------------------------------------------------

-- 7.1  A: BEGIN; UPDATE books SET available_copies = 0 WHERE id = 11;
--      B: SELECT available_copies FROM books WHERE id = 11;   -- what does B see?
--      A: COMMIT;   B: run it again.

-- 7.2  A: BEGIN; SELECT * FROM books WHERE id = 11 FOR UPDATE;
--      B: BEGIN; SELECT * FROM books WHERE id = 11 FOR UPDATE;  -- what happens?
--      A: COMMIT;   -- and now B?
--      This is exactly what LoanStore.Checkout needs.

-- 7.3  While B is stuck in 7.2, open a third window and run:
--        SELECT pid, state, wait_event_type, query FROM pg_stat_activity WHERE datname = 'library';
