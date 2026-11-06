package clients

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
	r.Get("/list/{id}", a.ListClients)
	return r
}

func (a *API) ListClients(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id must be present in url"})
		return
	}

	clients, err := a.db.ListClients(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(clients)
}
