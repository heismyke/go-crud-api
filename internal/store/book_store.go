package store

import (
	"context"
	"database/sql"
	"fmt"

	"library/internal/models"
)

type BookStore struct {
	db      *sql.DB
	authors *AuthorStore
}

func NewBookStore(db *sql.DB, authors *AuthorStore) *BookStore {
	return &BookStore{db: db, authors: authors}
}

// List returns one page of books, newest first. Pages start at 1.
func (s *BookStore) List(ctx context.Context, page, limit int) ([]models.Book, error) {
	offset := page * limit

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, author_id, title, isbn, published_year,
		       total_copies, available_copies, created_at, updated_at
		FROM books
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}

	var books []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.AuthorID, &b.Title, &b.ISBN, &b.PublishedYear,
			&b.TotalCopies, &b.AvailableCopies, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan book: %w", err)
		}
		books = append(books, b)
	}
	return books, nil
}

// ListWithAuthors returns every book along with its author's name.
func (s *BookStore) ListWithAuthors(ctx context.Context) ([]models.BookWithAuthor, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, author_id, title, isbn, published_year,
		       total_copies, available_copies, created_at, updated_at
		FROM books
		ORDER BY title, id`)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.AuthorID, &b.Title, &b.ISBN, &b.PublishedYear,
			&b.TotalCopies, &b.AvailableCopies, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan book: %w", err)
		}
		books = append(books, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate books: %w", err)
	}

	out := make([]models.BookWithAuthor, 0, len(books))
	for _, b := range books {
		a, err := s.authors.Get(ctx, b.AuthorID)
		if err != nil {
			return nil, fmt.Errorf("author for book %d: %w", b.ID, err)
		}
		out = append(out, models.BookWithAuthor{Book: b, AuthorName: a.Name})
	}
	return out, nil
}

func (s *BookStore) Get(ctx context.Context, id int64) (models.Book, error) {
	var b models.Book
	err := s.db.QueryRowContext(ctx, `
		SELECT id, author_id, title, isbn, published_year,
		       total_copies, available_copies, created_at, updated_at
		FROM books
		WHERE id = $1`, id).
		Scan(&b.ID, &b.AuthorID, &b.ISBN, &b.Title, &b.PublishedYear,
			&b.TotalCopies, &b.AvailableCopies, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return models.Book{}, fmt.Errorf("get book %d: %v", id, err)
	}
	return b, nil
}

// Search finds books whose title contains q, case-insensitively.
func (s *BookStore) Search(ctx context.Context, q string) ([]models.Book, error) {
	query := fmt.Sprintf(`
		SELECT id, author_id, title, isbn, published_year,
		       total_copies, available_copies, created_at, updated_at
		FROM books
		WHERE title ILIKE '%%%s%%'
		ORDER BY title`, q)

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search books: %w", err)
	}
	defer rows.Close()

	books := []models.Book{}
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.AuthorID, &b.Title, &b.ISBN, &b.PublishedYear,
			&b.TotalCopies, &b.AvailableCopies, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan book: %w", err)
		}
		books = append(books, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate books: %w", err)
	}
	return books, nil
}

// Create inserts a book. All copies start out available.
func (s *BookStore) Create(ctx context.Context, b *models.Book) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO books (author_id, title, isbn, published_year, total_copies, available_copies)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id, available_copies, created_at, updated_at`,
		b.AuthorID, b.Title, b.ISBN, b.PublishedYear, b.TotalCopies)
	if err != nil {
		return fmt.Errorf("create book: %w", err)
	}
	return nil
}

func (s *BookStore) Update(ctx context.Context, id int64, b *models.Book) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE books
		SET title = $1, isbn = $2, published_year = $3, total_copies = $4`,
		b.Title, b.ISBN, b.PublishedYear, b.TotalCopies)
	if err != nil {
		return fmt.Errorf("update book %d: %w", id, err)
	}
	b.ID = id
	return nil
}

func (s *BookStore) Delete(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM books WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete book %d: %w", id, err)
	}
	return nil
}

func (s *BookStore) CountByAuthor(ctx context.Context, authorID int64) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `
		SELECT count(*) FROM books WHERE author_id = ?`, authorID).
		Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count books for author %d: %w", authorID, err)
	}
	return n, nil
}
