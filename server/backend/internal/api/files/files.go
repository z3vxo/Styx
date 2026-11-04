package files

import (
	"encoding/json"
	"io"
	"net/http"
	"styx/internal/db"
	"sync"

	"github.com/go-chi/chi"
)

type API struct {
	db         *db.DB
	uploadsMap map[string]db.FileInfo
	mu         sync.RWMutex
}

func New(database *db.DB) *API {
	return &API{db: database, uploadsMap: make(map[string]db.FileInfo)}
}

func (a *API) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/start", a.StartUpload)
	r.Put("/begin/{id}", a.BeginUpload)
	r.Delete("/delete/{id}", a.DeleteUpload)
	r.Get("/list/{campaignID}", a.ListUploads)
	return r
}

func (a *API) ListUploads(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "campaignID")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "must provide file id in url"})
		return
	}

	data, err := a.db.ListUploads(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func (a *API) StartUpload(w http.ResponseWriter, r *http.Request) {
	var req db.FileInfo

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	id, err := a.PrepeareDisk(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (a *API) DeleteUpload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "must provide file id in url"})
		return
	}

	if err := a.DeleteLocal(id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}

func (a *API) BeginUpload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "must provide file id in url"})
		return
	}

	f, ok := a.Get(id)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "file id not found"})
		return
	}

	if _, err := io.Copy(f.Handle, r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed writing upload"})
		return
	}

	if err := a.db.InsertFile(f); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}

	w.WriteHeader(http.StatusOK)
}
