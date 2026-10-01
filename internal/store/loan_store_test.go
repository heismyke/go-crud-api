package store

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestLoanStore_Return_MarksReturnedAndRestocks(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Ursula K. Le Guin")
	bid := seedBook(t, aid, "The Dispossessed", "1", 1, 0) // the one copy is out
	mid := seedMember(t, "ada@example.com")
	lid := seedLoan(t, bid, mid, 24*time.Hour, false)

	l, err := NewLoanStore(testDB).Return(t.Context(), lid)
	if err != nil {
		t.Fatalf("Return: %v", err)
	}
	if l.ReturnedAt == nil {
		t.Error("returned loan has ReturnedAt == nil")
	}

	var returnedAt *time.Time
	testDB.QueryRow(`SELECT returned_at FROM loans WHERE id = $1`, lid).Scan(&returnedAt)
	if returnedAt == nil {
		t.Error("loans.returned_at is still NULL in the database; the change was never saved")
	}
	if got := availableCopies(t, bid); got != 1 {
		t.Errorf("available_copies = %d, want 1", got)
	}
}

func TestLoanStore_Return_TwiceIsAlreadyReturned(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Ursula K. Le Guin")
	bid := seedBook(t, aid, "The Dispossessed", "1", 1, 0)
	mid := seedMember(t, "ada@example.com")
	lid := seedLoan(t, bid, mid, 24*time.Hour, false)
	s := NewLoanStore(testDB)

	if _, err := s.Return(t.Context(), lid); err != nil {
		t.Fatalf("first Return: %v", err)
	}
	if _, err := s.Return(t.Context(), lid); !errors.Is(err, ErrAlreadyReturned) {
		t.Errorf("second Return: want ErrAlreadyReturned, got %v", err)
	}
	if got := availableCopies(t, bid); got != 1 {
		t.Errorf("available_copies = %d after returning one loan twice, want 1", got)
	}
}

func TestLoanStore_Return_MissingIsErrNotFound(t *testing.T) {
	resetDB(t)
	_, err := NewLoanStore(testDB).Return(t.Context(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestLoanStore_Checkout_HappyPath(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Octavia E. Butler")
	bid := seedBook(t, aid, "Kindred", "1", 2, 2)
	mid := seedMember(t, "grace@example.com")

	l, err := NewLoanStore(testDB).Checkout(t.Context(), bid, mid)
	if err != nil {
		t.Fatalf("Checkout: %v", err)
	}
	if l.ID == 0 || l.BookID != bid || l.MemberID != mid {
		t.Errorf("loan = %+v", l)
	}
	if l.ReturnedAt != nil {
		t.Error("new loan already has ReturnedAt set")
	}
	if d := l.DueAt.Sub(l.BorrowedAt); d < LoanPeriod-time.Minute || d > LoanPeriod+time.Minute {
		t.Errorf("due %s after borrowing, want %s", d, LoanPeriod)
	}
	if got := availableCopies(t, bid); got != 1 {
		t.Errorf("available_copies = %d, want 1", got)
	}
}

func TestLoanStore_Checkout_NoCopies(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "N. K. Jemisin")
	bid := seedBook(t, aid, "The Fifth Season", "1", 1, 0)
	mid := seedMember(t, "alan@example.com")

	_, err := NewLoanStore(testDB).Checkout(t.Context(), bid, mid)
	if !errors.Is(err, ErrNoCopies) {
		t.Fatalf("want ErrNoCopies, got %v", err)
	}
}

func TestLoanStore_Checkout_MissingBookOrMember(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "N. K. Jemisin")
	bid := seedBook(t, aid, "The Fifth Season", "1", 1, 1)
	mid := seedMember(t, "alan@example.com")
	s := NewLoanStore(testDB)

	if _, err := s.Checkout(t.Context(), 999, mid); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing book: want ErrNotFound, got %v", err)
	}
	if _, err := s.Checkout(t.Context(), bid, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing member: want ErrNotFound, got %v", err)
	}
	if got := availableCopies(t, bid); got != 1 {
		t.Errorf("available_copies = %d after a failed checkout, want 1 (was the transaction rolled back?)", got)
	}
}

func TestLoanStore_Checkout_LoanLimit(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Terry Pratchett")
	shelf := seedBook(t, aid, "Guards! Guards!", "1", 10, 10-MaxActiveLoans)
	mid := seedMember(t, "edsger@example.com")
	for range MaxActiveLoans {
		seedLoan(t, shelf, mid, 24*time.Hour, false)
	}
	seedLoan(t, shelf, mid, -24*time.Hour, true) // returned loans don't count

	_, err := NewLoanStore(testDB).Checkout(t.Context(), shelf, mid)
	if !errors.Is(err, ErrLoanLimit) {
		t.Fatalf("want ErrLoanLimit, got %v", err)
	}
	if got := availableCopies(t, shelf); got != 10-MaxActiveLoans {
		t.Errorf("available_copies changed to %d after a refused checkout", got)
	}
}

func TestLoanStore_Checkout_ConcurrentLastCopy(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Ted Chiang")
	bid := seedBook(t, aid, "Exhalation", "1", 1, 1) // exactly one copy
	const n = 10
	members := make([]int64, n)
	for i := range members {
		members[i] = seedMember(t, fmt.Sprintf("racer%d@example.com", i))
	}

	s := NewLoanStore(testDB)
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		ok, empty int
		other     []error
	)
	start := make(chan struct{})
	for _, mid := range members {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := s.Checkout(t.Context(), bid, mid)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				ok++
			case errors.Is(err, ErrNoCopies):
				empty++
			default:
				other = append(other, err)
			}
		}()
	}
	close(start)
	wg.Wait()

	if len(other) > 0 {
		t.Fatalf("unexpected errors: %v", other)
	}
	if ok != 1 || empty != n-1 {
		t.Errorf("%d checkouts succeeded and %d got ErrNoCopies; want 1 and %d", ok, empty, n-1)
	}
	var loans int
	testDB.QueryRow(`SELECT count(*) FROM loans WHERE book_id = $1`, bid).Scan(&loans)
	if got := availableCopies(t, bid); got != 0 || loans != 1 {
		t.Errorf("after the race: available_copies = %d, loans = %d; want 0 and 1", got, loans)
	}
}

func TestLoanStore_ListOverdue(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Octavia E. Butler")
	b1 := seedBook(t, aid, "Kindred", "1", 5, 5)
	b2 := seedBook(t, aid, "Dawn", "2", 5, 5)
	ada := seedMember(t, "ada@example.com")
	alan := seedMember(t, "alan@example.com")

	late := seedLoan(t, b1, ada, -3*24*time.Hour-time.Hour, false)   // overdue ~3 days
	later := seedLoan(t, b2, alan, -10*24*time.Hour-time.Hour, false) // overdue ~10 days
	seedLoan(t, b1, alan, -5*24*time.Hour, true)                      // overdue but returned
	seedLoan(t, b2, ada, 5*24*time.Hour, false)                       // not due yet

	got, err := NewLoanStore(testDB).ListOverdue(t.Context())
	if err != nil {
		t.Fatalf("ListOverdue: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d overdue loans, want 2: %+v", len(got), got)
	}
	if got[0].LoanID != later || got[1].LoanID != late {
		t.Errorf("order = [%d %d], want most overdue first [%d %d]", got[0].LoanID, got[1].LoanID, later, late)
	}
	if got[0].BookTitle != "Dawn" || got[0].MemberEmail != "alan@example.com" {
		t.Errorf("first row joined wrong: %+v", got[0])
	}
	if got[0].DaysOverdue != 10 || got[1].DaysOverdue != 3 {
		t.Errorf("DaysOverdue = %d, %d; want 10, 3", got[0].DaysOverdue, got[1].DaysOverdue)
	}
}

func TestLoanStore_ListByMember(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Terry Pratchett")
	bid := seedBook(t, aid, "Small Gods", "1", 5, 5)
	ada := seedMember(t, "ada@example.com")
	quiet := seedMember(t, "quiet@example.com")
	seedLoan(t, bid, ada, 24*time.Hour, false)
	seedLoan(t, bid, ada, -24*time.Hour, true)
	s := NewLoanStore(testDB)

	all, err := s.ListByMember(t.Context(), ada, false)
	if err != nil || len(all) != 2 {
		t.Errorf("all loans: got %d, err %v; want 2", len(all), err)
	}
	active, err := s.ListByMember(t.Context(), ada, true)
	if err != nil || len(active) != 1 {
		t.Errorf("active loans: got %d, err %v; want 1", len(active), err)
	}
	none, err := s.ListByMember(t.Context(), quiet, false)
	if err != nil || none == nil || len(none) != 0 {
		t.Errorf("member with no loans: got %v, err %v; want [] (non-nil, empty)", none, err)
	}
	if _, err := s.ListByMember(t.Context(), 999, false); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown member: want ErrNotFound, got %v", err)
	}
}
