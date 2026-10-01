package handlers

import (
	"net/http"
)

type checkoutInput struct {
	BookID   int64 `json:"book_id"`
	MemberID int64 `json:"member_id"`
}

// POST /loans  {"book_id": 1, "member_id": 5}
func (a *API) checkout(w http.ResponseWriter, r *http.Request) {
	var in checkoutInput
	if err := decodeJSON(w, r, &in); err != nil {
		badRequest(w, err.Error())
		return
	}
	if in.BookID <= 0 || in.MemberID <= 0 {
		badRequest(w, "book_id and member_id are required")
		return
	}
	loan, err := a.Loans.Checkout(r.Context(), in.BookID, in.MemberID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, loan)
}

// POST /loans/{id}/return
func (a *API) returnLoan(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	loan, err := a.Loans.Return(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, loan)
}

func (a *API) listOverdue(w http.ResponseWriter, r *http.Request) {
	loans, err := a.Loans.ListOverdue(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, loans)
}
