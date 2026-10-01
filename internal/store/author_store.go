package store

// AuthorStore is the REFERENCE implementation. It has no planted bugs.
// Read it closely before touching the other stores: every pattern you need
// (query vs query-row vs exec, scanning, NULLs, not-found, constraint errors,
// RETURNING, RowsAffected) shows up here once.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"library/internal/models"
)

type AuthorStore struct {
	db *sql.DB
}

func NewAuthorStore(db *sql.DB) *AuthorStore {
	return &AuthorStore{db: db}
}

// List: many rows -> QueryContext + rows.Next loop.
func (s *AuthorStore) List(ctx context.Context) ([]models.Author, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, bio, created_at
		FROM authors
		ORDER BY name, id`)
	if err != nil {
		return nil, fmt.Errorf("list authors: %w", err)
	}
	// Until rows are closed, the connection can't go back to the pool.
	defer rows.Close()

	authors := []models.Author{} // non-nil, so JSON encodes [] instead of null
	for rows.Next() {
		var a models.Author
		// Scan targets must match the SELECT list: same order, compatible types.
		// a.Bio is *string, so a NULL bio scans as nil instead of erroring.
		if err := rows.Scan(&a.ID, &a.Name, &a.Bio, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan author: %w", err)
		}
		authors = append(authors, a)
	}
	// rows.Next() returns false both when it's done AND when it failed.
	// rows.Err() is how you tell the difference.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate authors: %w", err)
	}
	return authors, nil
}

// Get: exactly one row -> QueryRowContext. Its errors are deferred to Scan.
func (s *AuthorStore) Get(ctx context.Context, id int64) (models.Author, error) {
	var a models.Author
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, bio, created_at
		FROM authors
		WHERE id = $1`, id).
		Scan(&a.ID, &a.Name, &a.Bio, &a.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Author{}, ErrNotFound
	}
	if err != nil {
		// %w (not %v) keeps the original error inspectable with errors.Is / errors.As.
		return models.Author{}, fmt.Errorf("get author %d: %w", id, err)
	}
	return a, nil
}

// Create: INSERT ... RETURNING lets the database fill in id/defaults and hand
// them back in the same round trip. Because it returns a row, use QueryRow.
func (s *AuthorStore) Create(ctx context.Context, a *models.Author) error {
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO authors (name, bio)
		VALUES ($1, $2)
		RETURNING id, created_at`,
		a.Name, a.Bio).
		Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return fmt.Errorf("create author: %w", err)
	}
	return nil
}

// Update: using RETURNING here means "no row came back" == "no such id",
// so sql.ErrNoRows doubles as the not-found check.
func (s *AuthorStore) Update(ctx context.Context, id int64, a *models.Author) error {
	err := s.db.QueryRowContext(ctx, `
		UPDATE authors
		SET name = $1, bio = $2
		WHERE id = $3
		RETURNING id, created_at`,
		a.Name, a.Bio, id).
		Scan(&a.ID, &a.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update author %d: %w", id, err)
	}
	return nil
}

// Delete: no rows to read back -> ExecContext, then RowsAffected for not-found.
func (s *AuthorStore) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM authors WHERE id = $1`, id)
	if isForeignKeyViolation(err) {
		// books.author_id still points at this author.
		return fmt.Errorf("author %d still has books: %w", id, ErrConflict)
	}
	if err != nil {
		return fmt.Errorf("delete author %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete author %d: %w", id, err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
