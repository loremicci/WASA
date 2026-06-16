package api

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/loremicci/WASA/service/api/reqcontext"
	"net/http"
)

type LoginRequest struct {
	Name string `json:"name"`
}

type LoginResponse struct {
	Identifier string `json:"identifier"`
}

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(req.Name) < 3 || len(req.Name) > 16 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := rt.db.DoLogin(req.Name)
	if err != nil {
		ctx.Logger.WithError(err).Error("login failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(LoginResponse{Identifier: id})
}

func (rt *_router) doLogout(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	err := rt.db.DoLogout(GetUserIDFromContext(r))
	if err != nil {
		ctx.Logger.WithError(err).Error("logout failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
