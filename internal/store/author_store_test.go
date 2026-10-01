package store

import (
	"errors"
	"testing"

	"library/internal/models"
)

// These should pass from the start. If they don't, your setup is broken.

func TestAuthorStore_CRUD(t *testing.T) {
	resetDB(t)
	s := NewAuthorStore(testDB)
	ctx := t.Context()

	a := models.Author{Name: "Ted Chiang"}
	if err := s.Create(ctx, &a); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if a.ID == 0 || a.CreatedAt.IsZero() {
		t.Fatalf("Create did not fill ID/CreatedAt: %+v", a)
	}

	got, err := s.Get(ctx, a.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Ted Chiang" || got.Bio != nil {
		t.Fatalf("Get = %+v", got)
	}

	bio := "Short fiction."
	upd := models.Author{Name: "Ted Chiang", Bio: &bio}
	if err := s.Update(ctx, a.ID, &upd); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ = s.Get(ctx, a.ID)
	if got.Bio == nil || *got.Bio != bio {
		t.Fatalf("after Update, Bio = %v", got.Bio)
	}

	list, err := s.List(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("List = %d authors, err %v", len(list), err)
	}

	if err := s.Delete(ctx, a.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get(ctx, a.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after Delete: want ErrNotFound, got %v", err)
	}
	if err := s.Delete(ctx, a.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second Delete: want ErrNotFound, got %v", err)
	}
	if err := s.Update(ctx, a.ID, &upd); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Update missing: want ErrNotFound, got %v", err)
	}
}

func TestAuthorStore_DeleteWithBooksIsConflict(t *testing.T) {
	resetDB(t)
	aid := seedAuthor(t, "Octavia E. Butler")
	seedBook(t, aid, "Kindred", "9780807083697", 1, 1)

	err := NewAuthorStore(testDB).Delete(t.Context(), aid)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}
}
