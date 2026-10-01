package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"library/internal/store"
)

type API struct {
	DB      *sql.DB
	Authors *store.AuthorStore
	Books   *store.BookStore
	Members *store.MemberStore
	Loans   *store.LoanStore
}

// Routes uses Go 1.22+ method-and-wildcard patterns ("GET /books/{id}").
// Read a wildcard with r.PathValue("id").
func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", a.health)

	mux.HandleFunc("GET /authors", a.listAuthors)
	mux.HandleFunc("POST /authors", a.createAuthor)
	mux.HandleFunc("GET /authors/{id}", a.getAuthor)
	mux.HandleFunc("PUT /authors/{id}", a.updateAuthor)
	mux.HandleFunc("DELETE /authors/{id}", a.deleteAuthor)
	mux.HandleFunc("GET /authors/{id}/stats", a.authorStats)

	mux.HandleFunc("GET /books", a.listBooks)
	mux.HandleFunc("POST /books", a.createBook)
	mux.HandleFunc("GET /books/search", a.searchBooks)
	mux.HandleFunc("GET /books/with-authors", a.listBooksWithAuthors)
	mux.HandleFunc("GET /books/{id}", a.getBook)
	mux.HandleFunc("PUT /books/{id}", a.updateBook)
	mux.HandleFunc("DELETE /books/{id}", a.deleteBook)

	mux.HandleFunc("GET /members", a.listMembers)
	mux.HandleFunc("POST /members", a.createMember)
	mux.HandleFunc("GET /members/{id}", a.getMember)
	mux.HandleFunc("PUT /members/{id}", a.updateMember)
	mux.HandleFunc("DELETE /members/{id}", a.deleteMember)
	mux.HandleFunc("GET /members/{id}/loans", a.listMemberLoans)

	mux.HandleFunc("POST /loans", a.checkout)
	mux.HandleFunc("POST /loans/{id}/return", a.returnLoan)
	mux.HandleFunc("GET /loans/overdue", a.listOverdue)

	return logRequests(mux)
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := a.DB.PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db unreachable", "error": err.Error()})
		return
	}
	s := a.DB.Stats()
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"pool": map[string]int{
			"max_open": s.MaxOpenConnections,
			"open":     s.OpenConnections,
			"in_use":   s.InUse,
			"idle":     s.Idle,
		},
	})
}
