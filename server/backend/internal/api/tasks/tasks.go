package tasks

import (
	"encoding/json"
	"net/http"
	authmgr "styx/internal/auth"
	"styx/internal/db"

	"github.com/go-chi/chi"
)

type API struct {
	db   *db.DB
	auth *authmgr.Auth
}

func New(database *db.DB, a *authmgr.Auth) *API {
	return &API{db: database, auth: a}
}

func (a *API) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(a.auth.AuthWare())
	r.Post("/new", a.NewTask)
	return r
}

func (a *API) NewTask(w http.ResponseWriter, r *http.Request) {
	var req db.NewTask
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if err := a.db.InsertTask(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}
