package store

import (
	"errors"
	"testing"
	"time"

	"library/internal/models"
)

func TestMemberStore_CRUD(t *testing.T) {
	resetDB(t)
	s := NewMemberStore(testDB)
	ctx := t.Context()

	empty, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List on empty table: %v", err)
	}
	if empty == nil {
		t.Error("List on empty table returned nil, want []")
	}

	m := models.Member{Email: "barbara@example.com", FullName: "Barbara Liskov"}
	if err := s.Create(ctx, &m); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if m.ID == 0 || m.JoinedAt.IsZero() {
		t.Fatalf("Create did not fill ID/JoinedAt: %+v", m)
	}

	got, err := s.Get(ctx, m.ID)
	if err != nil || got.Email != m.Email || got.FullName != m.FullName {
		t.Fatalf("Get = %+v, %v", got, err)
	}

	upd := models.Member{Email: "liskov@example.com", FullName: "Barbara Liskov"}
	if err := s.Update(ctx, m.ID, &upd); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got, _ := s.Get(ctx, m.ID); got.Email != "liskov@example.com" {
		t.Errorf("after Update, Email = %q", got.Email)
	}

	seedMember(t, "zed@example.com")
	list, err := s.List(ctx)
	if err != nil || len(list) != 2 {
		t.Fatalf("List = %d members, err %v; want 2", len(list), err)
	}
	if list[0].FullName != "Barbara Liskov" {
		t.Errorf("List not ordered by full_name: %q before %q", list[0].FullName, list[1].FullName)
	}

	if err := s.Delete(ctx, m.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get(ctx, m.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get after Delete: want ErrNotFound, got %v", err)
	}
	if err := s.Delete(ctx, m.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("second Delete: want ErrNotFound, got %v", err)
	}
	if err := s.Update(ctx, m.ID, &upd); !errors.Is(err, ErrNotFound) {
		t.Errorf("Update missing: want ErrNotFound, got %v", err)
	}
}

func TestMemberStore_DuplicateEmailIsConflict(t *testing.T) {
	resetDB(t)
	seedMember(t, "ada@example.com")
	other := seedMember(t, "grace@example.com")
	s := NewMemberStore(testDB)

	dup := models.Member{Email: "ada@example.com", FullName: "Imposter"}
	if err := s.Create(t.Context(), &dup); !errors.Is(err, ErrConflict) {
		t.Errorf("Create with taken email: want ErrConflict, got %v", err)
	}
	if err := s.Update(t.Context(), other, &dup); !errors.Is(err, ErrConflict) {
		t.Errorf("Update to taken email: want ErrConflict, got %v", err)
	}
}

func TestMemberStore_DeleteWithLoansIsConflict(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Ted Chiang")
	bid := seedBook(t, aid, "Exhalation", "1", 1, 1)
	mid := seedMember(t, "ada@example.com")
	seedLoan(t, bid, mid, -time.Hour, true)

	if err := NewMemberStore(testDB).Delete(t.Context(), mid); !errors.Is(err, ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}
}
