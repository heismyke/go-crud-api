package handlers

import (
	"net/http"

	"library/internal/models"
)

// These handlers are done. They'll start working once MemberStore does.
// TODO(you): add a validate() method like authorInput has
// (email must contain "@", full_name required) and call it.

type memberInput struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

func (a *API) listMembers(w http.ResponseWriter, r *http.Request) {
	members, err := a.Members.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, members)
}

func (a *API) getMember(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	m, err := a.Members.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (a *API) createMember(w http.ResponseWriter, r *http.Request) {
	var in memberInput
	if err := decodeJSON(w, r, &in); err != nil {
		badRequest(w, err.Error())
		return
	}
	m := models.Member{Email: in.Email, FullName: in.FullName}
	if err := a.Members.Create(r.Context(), &m); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (a *API) updateMember(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	var in memberInput
	if err := decodeJSON(w, r, &in); err != nil {
		badRequest(w, err.Error())
		return
	}
	m := models.Member{Email: in.Email, FullName: in.FullName}
	if err := a.Members.Update(r.Context(), id, &m); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (a *API) deleteMember(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	if err := a.Members.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /members/{id}/loans?active=true
func (a *API) listMemberLoans(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		badRequest(w, err.Error())
		return
	}
	activeOnly := r.URL.Query().Get("active") == "true"
	loans, err := a.Loans.ListByMember(r.Context(), id, activeOnly)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, loans)
}
