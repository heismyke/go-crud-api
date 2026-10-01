package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"library/internal/models"
)

const (
	MaxActiveLoans = 3
	LoanPeriod     = 14 * 24 * time.Hour
)

type LoanStore struct {
	db *sql.DB
}

func NewLoanStore(db *sql.DB) *LoanStore {
	return &LoanStore{db: db}
}

// Checkout lends a copy of a book to a member.
//
// TODO(you): implement. Requirements:
//   - Book doesn't exist                    -> ErrNotFound
//   - Member doesn't exist                  -> ErrNotFound
//   - available_copies is 0                 -> ErrNoCopies
//   - member already has MaxActiveLoans     -> ErrLoanLimit
//     (an active loan is one with returned_at IS NULL)
//   - otherwise insert a loan due LoanPeriod from now, decrement
//     books.available_copies, and return the new loan.
//
// All of that has to happen atomically: if any step fails, nothing changes.
// It also has to be safe when two requests race for the last copy
// (see TestLoanStore_Checkout_ConcurrentLastCopy).
//
// Useful things to look up: db.BeginTx, tx.Rollback/Commit,
// SELECT ... FOR UPDATE.
func (s *LoanStore) Checkout(ctx context.Context, bookID, memberID int64) (models.Loan, error) {
	return models.Loan{}, ErrNotImplemented
}

// Return marks a loan as returned and puts the copy back on the shelf.
//   - Loan doesn't exist       -> ErrNotFound
//   - Loan already returned    -> ErrAlreadyReturned
func (s *LoanStore) Return(ctx context.Context, loanID int64) (models.Loan, error) {
	var l models.Loan

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return l, fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		UPDATE loans
		SET returned_at = now()
		WHERE id = $1
		RETURNING id, book_id, member_id, borrowed_at, due_at, returned_at`, loanID).
		Scan(&l.ID, &l.BookID, &l.MemberID, &l.BorrowedAt, &l.DueAt, &l.ReturnedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return l, ErrNotFound
	}
	if err != nil {
		return l, fmt.Errorf("mark loan %d returned: %w", loanID, err)
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE books
		SET available_copies = available_copies + 1
		WHERE id = $1`, l.BookID)
	if err != nil {
		return l, fmt.Errorf("restock book %d: %w", l.BookID, err)
	}

	return l, nil
}

// ListOverdue returns active loans whose due date has passed, most overdue
// first, with the book title and member email filled in.
// DaysOverdue is whole days (rounded down) since due_at.
//
// TODO(you): implement. One query with JOINs, not one query per loan.
func (s *LoanStore) ListOverdue(ctx context.Context) ([]models.OverdueLoan, error) {
	return nil, ErrNotImplemented
}

// ListByMember returns a member's loans, newest first. With activeOnly, only
// loans that haven't been returned. Unknown member -> ErrNotFound (an empty
// list would hide typos in the id).
//
// TODO(you): implement.
func (s *LoanStore) ListByMember(ctx context.Context, memberID int64, activeOnly bool) ([]models.Loan, error) {
	return nil, ErrNotImplemented
}
