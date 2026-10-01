package handlers

import (
	"net/http"

	"library/internal/models"
)

type bookInput struct {
	AuthorID      int64  `json:"author_id"`
	Title         string `json:"title"`
	ISBN          string `json:"isbn"`
	PublishedYear int    `json:"published_year"`
	TotalCopies   int    `json:"total_copies"`
}

func (in bookInput) toModel() models.Book {
	return models.Book{
		AuthorID:      in.AuthorID,
		Title:         in.Title,
		ISBN:          in.ISBN,
		PublishedYear: in.PublishedYear,
		TotalCopies:   in.TotalCopies,
	}
}

// GET /books?page=1&limit=20
func (a *API) listBooks(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 20)

	books, err := a.Books.List(r.Context(), page, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, books)
}

// GET /books/search?q=gods
func (a *API) searchBooks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	books, err := a.Books.Search(r.Context(), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, books)
}

func (a *API) listBooksWithAuthors(w http.ResponseWriter, r *http.Request) {
	books, err := a.Books.ListWithAuthors(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, books)
}

func (a *API) getBook(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	book, err := a.Books.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (a *API) createBook(w http.ResponseWriter, r *http.Request) {
	var in bookInput
	decodeJSON(w, r, &in)

	book := in.toModel()
	if err := a.Books.Create(r.Context(), &book); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, book)
}

func (a *API) updateBook(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	var in bookInput
	if err := decodeJSON(w, r, &in); err != nil {
		badRequest(w, err.Error())
		return
	}
	book := in.toModel()
	if err := a.Books.Update(r.Context(), id, &book); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (a *API) deleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	if err := a.Books.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
