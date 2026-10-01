package store

import (
	"context"
	"database/sql"

	"library/internal/models"
)

// MemberStore is yours to write from scratch. Use AuthorStore as the pattern.
//
// Requirements (the tests in member_store_test.go check these):
//   - List:   all members ordered by full_name; empty table -> [] (not nil)
//   - Get:    unknown id -> ErrNotFound
//   - Create: fill in ID and JoinedAt; duplicate email -> ErrConflict
//   - Update: change email + full_name; unknown id -> ErrNotFound;
//     duplicate email -> ErrConflict
//   - Delete: unknown id -> ErrNotFound; member with loans -> ErrConflict
type MemberStore struct {
	db *sql.DB
}

func NewMemberStore(db *sql.DB) *MemberStore {
	return &MemberStore{db: db}
}

func (s *MemberStore) List(ctx context.Context) ([]models.Member, error) {
	return nil, ErrNotImplemented
}

func (s *MemberStore) Get(ctx context.Context, id int64) (models.Member, error) {
	return models.Member{}, ErrNotImplemented
}

func (s *MemberStore) Create(ctx context.Context, m *models.Member) error {
	return ErrNotImplemented
}

func (s *MemberStore) Update(ctx context.Context, id int64, m *models.Member) error {
	return ErrNotImplemented
}

func (s *MemberStore) Delete(ctx context.Context, id int64) error {
	return ErrNotImplemented
}
