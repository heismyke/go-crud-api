package store

// These are integration tests: they run real SQL against the db_test container.
// Start it with:  docker compose up -d db_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"library/internal/db"
)

var (
	testDB  *sql.DB
	testDSN string
)

func TestMain(m *testing.M) {
	testDSN = os.Getenv("TEST_DATABASE_URL")
	if testDSN == "" {
		testDSN = "postgres://library:library@localhost:5441/library_test?sslmode=disable"
	}
	var err error
	testDB, err = db.Open(testDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\ncannot reach the test database: %v\nstart it with: docker compose up -d db_test\n\n", err)
		os.Exit(1)
	}
	testDB.SetMaxOpenConns(20)
	testDB.SetMaxIdleConns(20)

	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

// resetDB wipes every table so each test starts from nothing.
func resetDB(t *testing.T) {
	t.Helper()
	if _, err := testDB.Exec(`TRUNCATE loans, books, members, authors RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("reset db: %v", err)
	}
}

// The seed helpers use raw SQL on purpose, so a bug in the store under test
// can't break the test's setup.

func seedAuthor(t *testing.T, name string) int64 {
	t.Helper()
	var id int64
	if err := testDB.QueryRow(`INSERT INTO authors (name) VALUES ($1) RETURNING id`, name).Scan(&id); err != nil {
		t.Fatalf("seed author: %v", err)
	}
	return id
}

func seedBook(t *testing.T, authorID int64, title, isbn string, total, available int) int64 {
	t.Helper()
	var id int64
	err := testDB.QueryRow(`
		INSERT INTO books (author_id, title, isbn, published_year, total_copies, available_copies)
		VALUES ($1, $2, $3, 2000, $4, $5)
		RETURNING id`, authorID, title, isbn, total, available).Scan(&id)
	if err != nil {
		t.Fatalf("seed book: %v", err)
	}
	return id
}

func seedMember(t *testing.T, email string) int64 {
	t.Helper()
	var id int64
	if err := testDB.QueryRow(`INSERT INTO members (email, full_name) VALUES ($1, $1) RETURNING id`, email).Scan(&id); err != nil {
		t.Fatalf("seed member: %v", err)
	}
	return id
}

// seedLoan inserts a loan due at now+dueIn (negative = already overdue).
// It does NOT touch books.available_copies; seed the book accordingly.
func seedLoan(t *testing.T, bookID, memberID int64, dueIn time.Duration, returned bool) int64 {
	t.Helper()
	var id int64
	err := testDB.QueryRow(`
		INSERT INTO loans (book_id, member_id, borrowed_at, due_at, returned_at)
		VALUES ($1, $2, $3, $3, CASE WHEN $4::boolean THEN now() END)
		RETURNING id`, bookID, memberID, time.Now().Add(dueIn), returned).Scan(&id)
	if err != nil {
		t.Fatalf("seed loan: %v", err)
	}
	// borrowed_at = due_at above keeps the INSERT short; push borrowed_at back.
	if _, err := testDB.Exec(`UPDATE loans SET borrowed_at = due_at - interval '14 days' WHERE id = $1`, id); err != nil {
		t.Fatalf("seed loan: %v", err)
	}
	return id
}

func availableCopies(t *testing.T, bookID int64) int {
	t.Helper()
	var n int
	if err := testDB.QueryRow(`SELECT available_copies FROM books WHERE id = $1`, bookID).Scan(&n); err != nil {
		t.Fatalf("read available_copies: %v", err)
	}
	return n
}
