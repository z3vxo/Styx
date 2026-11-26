package clients

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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
	r.Get("/info/{client_id}", a.GetClientInfo)
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

func ToInt32(raw string) (int32, error) {
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || id < 0 {
		return 0, errors.New("invalid client id")
	}
	return int32(id), nil
}

func (a *API) GetClientInfo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "client_id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id must be present in url"})
		return
	}

	clientID, err := ToInt32(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid client_id"})
		return
	}

	info, err := a.db.GetClientInfo(clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "client not found for given ID"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed looking up client info"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&info)
	return

}
