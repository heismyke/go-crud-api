package models

import "time"

// Pointer fields (*string, *time.Time) are how this project represents
// nullable columns: nil <-> NULL. The alternative is sql.NullString etc.

type Author struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Bio       *string   `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
}

type Book struct {
	ID              int64     `json:"id"`
	AuthorID        int64     `json:"author_id"`
	Title           string    `json:"title"`
	ISBN            string    `json:"isbn"`
	PublishedYear   int       `json:"published_year"`
	TotalCopies     int       `json:"total_copies"`
	AvailableCopies int       `json:"available_copies"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type BookWithAuthor struct {
	Book
	AuthorName string `json:"author_name"`
}

type Member struct {
	ID       int64     `json:"id"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
	JoinedAt time.Time `json:"joined_at"`
}

type Loan struct {
	ID         int64      `json:"id"`
	BookID     int64      `json:"book_id"`
	MemberID   int64      `json:"member_id"`
	BorrowedAt time.Time  `json:"borrowed_at"`
	DueAt      time.Time  `json:"due_at"`
	ReturnedAt *time.Time `json:"returned_at"`
}

// OverdueLoan is a read model: it's built from a JOIN, not a single table.
type OverdueLoan struct {
	LoanID      int64     `json:"loan_id"`
	BookID      int64     `json:"book_id"`
	BookTitle   string    `json:"book_title"`
	MemberID    int64     `json:"member_id"`
	MemberEmail string    `json:"member_email"`
	DueAt       time.Time `json:"due_at"`
	DaysOverdue int       `json:"days_overdue"`
}
