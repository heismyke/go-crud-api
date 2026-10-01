package handlers

import (
	"errors"
	"net/http"
	"strings"

	"library/internal/models"
)

// Reference handlers: decode -> validate -> call store -> map errors -> respond.

type authorInput struct {
	Name string  `json:"name"`
	Bio  *string `json:"bio"`
}

func (in authorInput) validate() error {
	if strings.TrimSpace(in.Name) == "" {
		return errors.New("name is required")
	}
	return nil
}

func (a *API) listAuthors(w http.ResponseWriter, r *http.Request) {
	authors, err := a.Authors.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, authors)
}

func (a *API) getAuthor(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	author, err := a.Authors.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, author)
}

func (a *API) createAuthor(w http.ResponseWriter, r *http.Request) {
	var in authorInput
	if err := decodeJSON(w, r, &in); err != nil {
		badRequest(w, err.Error())
		return
	}
	if err := in.validate(); err != nil {
		badRequest(w, err.Error())
		return
	}
	author := models.Author{Name: strings.TrimSpace(in.Name), Bio: in.Bio}
	if err := a.Authors.Create(r.Context(), &author); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, author)
}

func (a *API) updateAuthor(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	var in authorInput
	if err := decodeJSON(w, r, &in); err != nil {
		badRequest(w, err.Error())
		return
	}
	if err := in.validate(); err != nil {
		badRequest(w, err.Error())
		return
	}
	author := models.Author{Name: strings.TrimSpace(in.Name), Bio: in.Bio}
	if err := a.Authors.Update(r.Context(), id, &author); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, author)
}

func (a *API) deleteAuthor(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	if err := a.Authors.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) authorStats(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	author, err := a.Authors.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	count, err := a.Books.CountByAuthor(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"author":     author,
		"book_count": count,
	})
}
