package campaigns

import (
	"encoding/json"
	"net/http"
	"styx/internal/db"

	"github.com/go-chi/chi"
)

type API struct {
	db *db.DB
}

func New(database *db.DB) *API {
	return &API{db: database}
}

func (a *API) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/new", a.NewCampaign)
	r.Get("/list", a.ListCampaigns)
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
	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "must provide campaign uuid in url"})
		return
	}

	if err := a.db.DeleteCampaign(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "OK"})

}
