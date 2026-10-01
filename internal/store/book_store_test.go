package store

import (
	"errors"
	"testing"
	"time"

	"library/internal/db"
	"library/internal/models"
)

func newBookStore() *BookStore {
	return NewBookStore(testDB, NewAuthorStore(testDB))
}

func TestBookStore_Create_FillsGeneratedFields(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Ted Chiang")

	b := models.Book{AuthorID: aid, Title: "Exhalation", ISBN: "9781101947883", TotalCopies: 3}
	if err := newBookStore().Create(t.Context(), &b); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if b.ID == 0 {
		t.Fatal("Create did not set b.ID")
	}
	if b.CreatedAt.IsZero() || b.UpdatedAt.IsZero() {
		t.Error("Create did not set CreatedAt/UpdatedAt")
	}
	if b.AvailableCopies != 3 {
		t.Errorf("b.AvailableCopies = %d, want 3", b.AvailableCopies)
	}
}

func TestBookStore_Create_UnknownAuthorIsNotFound(t *testing.T) {
	resetDB(t)
	b := models.Book{AuthorID: 999, Title: "Orphan", ISBN: "x", TotalCopies: 1}
	err := newBookStore().Create(t.Context(), &b)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound for unknown author, got %v", err)
	}
}

func TestBookStore_Get_ReturnsCorrectFields(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Octavia E. Butler")
	id := seedBook(t, aid, "Kindred", "9780807083697", 4, 3)

	b, err := newBookStore().Get(t.Context(), id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if b.Title != "Kindred" {
		t.Errorf("Title = %q, want %q", b.Title, "Kindred")
	}
	if b.ISBN != "9780807083697" {
		t.Errorf("ISBN = %q, want %q", b.ISBN, "9780807083697")
	}
	if b.TotalCopies != 4 || b.AvailableCopies != 3 {
		t.Errorf("copies = %d/%d, want 3/4", b.AvailableCopies, b.TotalCopies)
	}
}

func TestBookStore_Get_MissingIsErrNotFound(t *testing.T) {
	resetDB(t)
	_, err := newBookStore().Get(t.Context(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestBookStore_Get_NullPublishedYear(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "N. K. Jemisin")
	id := seedBook(t, aid, "Untitled", "TBD", 1, 1)
	testDB.Exec(`UPDATE books SET published_year = NULL WHERE id = $1`, id)

	if _, err := newBookStore().Get(t.Context(), id); err != nil {
		t.Fatalf("Get on a book with NULL published_year: %v", err)
	}
}

func TestBookStore_List_Pagination(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Terry Pratchett")
	// One statement -> every row gets the exact same created_at.
	_, err := testDB.Exec(`
		INSERT INTO books (author_id, title, isbn, total_copies, available_copies)
		SELECT $1, 'Discworld ' || n, 'isbn-' || n, 1, 1
		FROM generate_series(1, 5) AS n`, aid)
	if err != nil {
		t.Fatal(err)
	}
	// Give them real years so a NULL-scan bug doesn't mask this test.
	testDB.Exec(`UPDATE books SET published_year = 1990`)

	s := newBookStore()
	seen := map[int64]bool{}
	for page, want := range map[int]int{1: 2, 2: 2, 3: 1} {
		books, err := s.List(t.Context(), page, 2)
		if err != nil {
			t.Fatalf("List page %d: %v", page, err)
		}
		if len(books) != want {
			t.Errorf("page %d: got %d books, want %d", page, len(books), want)
		}
		for _, b := range books {
			if seen[b.ID] {
				t.Errorf("book %d appeared on more than one page", b.ID)
			}
			seen[b.ID] = true
		}
	}
	if len(seen) != 5 {
		t.Errorf("paging through everything saw %d distinct books, want 5", len(seen))
	}
}

func TestBookStore_List_EmptyIsNotNil(t *testing.T) {
	resetDB(t)
	books, err := newBookStore().List(t.Context(), 1, 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if books == nil {
		t.Fatal("List on an empty table returned nil; JSON would encode it as null instead of []")
	}
}

func TestBookStore_List_ReleasesConnectionOnError(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Anonymous")
	_, err := testDB.Exec(`
		INSERT INTO books (author_id, title, isbn, published_year)
		SELECT $1, 'Draft ' || n, 'isbn-' || n, NULL
		FROM generate_series(1, 25) AS n`, aid)
	if err != nil {
		t.Fatal(err)
	}

	// Private pool so other tests can't affect the numbers.
	pool, err := db.Open(testDSN)
	if err != nil {
		t.Fatal(err)
	}
	s := NewBookStore(pool, NewAuthorStore(pool))
	_, _ = s.List(t.Context(), 1, 10) // may or may not error, depending on what you've fixed

	if inUse := pool.Stats().InUse; inUse != 0 {
		t.Fatalf("List returned but left %d connection(s) checked out of the pool", inUse)
	}
	pool.Close()
}

func TestBookStore_Search_FindsMatches(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Terry Pratchett")
	seedBook(t, aid, "Small Gods", "1", 1, 1)
	seedBook(t, aid, "Going Postal", "2", 1, 1)

	books, err := newBookStore().Search(t.Context(), "gods")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(books) != 1 || books[0].Title != "Small Gods" {
		t.Fatalf("Search(gods) = %+v", books)
	}
}

func TestBookStore_Search_IsNotInjectable(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Terry Pratchett")
	seedBook(t, aid, "Small Gods", "1", 1, 1)
	seedBook(t, aid, "Going Postal", "2", 1, 1)

	for _, q := range []string{
		`' OR '1'='1' --`,
		`'; DROP TABLE loans; --`,
	} {
		books, err := newBookStore().Search(t.Context(), q)
		if err != nil {
			t.Errorf("Search(%q) errored: %v (a safe query treats this as plain text)", q, err)
			continue
		}
		if len(books) != 0 {
			t.Errorf("Search(%q) returned %d books; user input was executed as SQL", q, len(books))
		}
	}
}

func TestBookStore_Update_OnlyChangesTargetRow(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Octavia E. Butler")
	target := seedBook(t, aid, "Kindred", "111", 1, 1)
	other := seedBook(t, aid, "Dawn", "222", 1, 1)
	testDB.Exec(`UPDATE books SET updated_at = now() - interval '1 day'`)

	b := models.Book{Title: "Kindred (2nd ed.)", ISBN: "111", TotalCopies: 1}
	if err := newBookStore().Update(t.Context(), target, &b); err != nil {
		t.Fatalf("Update: %v", err)
	}

	var otherTitle string
	testDB.QueryRow(`SELECT title FROM books WHERE id = $1`, other).Scan(&otherTitle)
	if otherTitle != "Dawn" {
		t.Fatalf("updating book %d renamed book %d to %q", target, other, otherTitle)
	}

	var updatedAt time.Time
	testDB.QueryRow(`SELECT updated_at FROM books WHERE id = $1`, target).Scan(&updatedAt)
	if time.Since(updatedAt) > time.Minute {
		t.Errorf("updated_at was not bumped (still %s)", updatedAt)
	}
}

func TestBookStore_Update_MissingIsErrNotFound(t *testing.T) {
	resetDB(t)
	b := models.Book{Title: "Ghost", ISBN: "0", TotalCopies: 1}
	err := newBookStore().Update(t.Context(), 999, &b)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestBookStore_Delete_MissingIsErrNotFound(t *testing.T) {
	resetDB(t)
	err := newBookStore().Delete(t.Context(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestBookStore_Delete_WithLoansIsConflict(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Ursula K. Le Guin")
	bid := seedBook(t, aid, "The Dispossessed", "1", 1, 1)
	mid := seedMember(t, "ada@example.com")
	seedLoan(t, bid, mid, -time.Hour, true)

	err := newBookStore().Delete(t.Context(), bid)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}
}

func TestBookStore_CountByAuthor(t *testing.T) {
	resetDB(t)
	a1 := seedAuthor(t, "A")
	a2 := seedAuthor(t, "B")
	seedBook(t, a1, "One", "1", 1, 1)
	seedBook(t, a1, "Two", "2", 1, 1)
	seedBook(t, a2, "Three", "3", 1, 1)

	n, err := newBookStore().CountByAuthor(t.Context(), a1)
	if err != nil {
		t.Fatalf("CountByAuthor: %v", err)
	}
	if n != 2 {
		t.Fatalf("CountByAuthor = %d, want 2", n)
	}
}

func TestBookStore_ListWithAuthors(t *testing.T) {
	resetDB(t)
	le := seedAuthor(t, "Ursula K. Le Guin")
	bu := seedAuthor(t, "Octavia E. Butler")
	seedBook(t, le, "A Wizard of Earthsea", "1", 1, 1)
	seedBook(t, bu, "Kindred", "2", 1, 1)
	seedBook(t, le, "The Dispossessed", "3", 1, 1)

	got, err := newBookStore().ListWithAuthors(t.Context())
	if err != nil {
		t.Fatalf("ListWithAuthors: %v", err)
	}
	want := map[string]string{
		"A Wizard of Earthsea": "Ursula K. Le Guin",
		"Kindred":              "Octavia E. Butler",
		"The Dispossessed":     "Ursula K. Le Guin",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rows, want %d", len(got), len(want))
	}
	for _, b := range got {
		if want[b.Title] != b.AuthorName {
			t.Errorf("%q: author %q, want %q", b.Title, b.AuthorName, want[b.Title])
		}
	}
}

// Not a correctness test: logs how long ListWithAuthors takes with lots of
// books. Run it before and after you fix the N+1 query:
//
//	go test ./internal/store -run ListWithAuthors_Timing -v
func TestBookStore_ListWithAuthors_Timing(t *testing.T) {
	resetDB(t)
	_, err := testDB.Exec(`
		INSERT INTO authors (name) SELECT 'Author ' || n FROM generate_series(1, 500) AS n;
		INSERT INTO books (author_id, title, isbn, published_year)
		SELECT (n % 500) + 1, 'Book ' || n, 'isbn-' || n, 2000 FROM generate_series(1, 2000) AS n;`)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	got, err := newBookStore().ListWithAuthors(t.Context())
	if err != nil {
		t.Skipf("ListWithAuthors errored (fix the other book bugs first): %v", err)
	}
	t.Logf("ListWithAuthors: %d rows in %s", len(got), time.Since(start))
}
