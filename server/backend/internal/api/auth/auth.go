package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	authmgr "styx/internal/auth"
	"styx/internal/db"

	"github.com/go-chi/chi"
)

type API struct {
	db   *db.DB
	auth *authmgr.Auth
}

func New(database *db.DB, authService *authmgr.Auth) *API {
	return &API{db: database, auth: authService}
}

func (a *API) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/login", a.login)
	r.Post("/refresh", a.refresh)
	return r
}

type LoginReq struct {
	User string `json:"username"`
	Pass string `json:"password"`
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var req LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed decoding json", http.StatusInternalServerError)
		return
	}

	if err := a.db.VerifyAuth(req.User, req.Pass); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "invalid creds", http.StatusUnauthorized)
			return
		}
		http.Error(w, "failed checking db", http.StatusInternalServerError)
		return
	}

	tokens, err := a.auth.NewJwtPair(req.User)
	if err != nil {
		http.Error(w, "failed generating jwt", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokens)

}

type ResfreshReq struct {
	RefreshToken string `json:"token"`
}

func (a *API) refresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req ResfreshReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if req.RefreshToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "refresh token is required"})
		return
	}

	claims, err := a.auth.VerifyRefresh(req.RefreshToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	tokens, err := a.auth.NewJwtPair(claims.Usename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokens)

}
