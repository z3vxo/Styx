package campaigns

import (
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
	r.Get("/list", a.ListCampaign)
	return r
}

func (a *API) NewCampaign(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusOK)
}

func (a *API) ListCampaign(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusOK)
}
