package campaigns

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	authmgr "styx/internal/auth"
	"styx/internal/db"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
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
	r.Post("/new", a.NewCampaign)
	r.Get("/list", a.ListCampaigns)
	r.Post("/delete/{id}", a.DeleteCampaign)
	r.Delete("/delete/{id}", a.DeleteCampaign)
	return r
}

type NewCamapaignReq struct {
	Name string `json:"name"`
}

func (a *API) NewCampaign(w http.ResponseWriter, r *http.Request) {
	var req NewCamapaignReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if err := a.db.InsertNewCampaign(req.Name); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}

func (a *API) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	data, err := a.db.ListCampaigns()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func (a *API) DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "must provide campaign uuid in url"})
		return
	}

	fileIDs, err := a.db.DeleteCampaign(id)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "campaign not found"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "could not locate stored files"})
		return
	}
	base := filepath.Join(home, ".local", "share", "styx", "user-files")
	for _, fileID := range fileIDs {
		if _, err := uuid.Parse(fileID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid stored file ID"})
			return
		}
		if err := os.RemoveAll(filepath.Join(base, fileID)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "could not remove stored files"})
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}
