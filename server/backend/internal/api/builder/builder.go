package builder

import (
	"encoding/json"
	"fmt"
	"net/http"
	authmgr "styx/internal/auth"
	"styx/internal/db"

	"github.com/go-chi/chi"
)

type API struct {
	db         *db.DB
	auth       *authmgr.Auth
	BuilderInt *Builder
}

func New(database *db.DB, authService *authmgr.Auth, backendHost string, backendPort int) *API {
	return &API{db: database, auth: authService, BuilderInt: NewBuilder(backendHost, backendPort)}
}

func (a *API) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/build", a.Build)
	r.Get("/list/{campaignID}", a.ListBuilds)

	return r
}

type BuildReq struct {
	CampaingID string `json:"campaign_id"`
	FileType   string `json:"file_type"`
	FileName   string `json:"file_name"`
	Domain     string `json:"domain"`
	Port       int    `json:"port"` 
}

func (a *API) Build(w http.ResponseWriter, r *http.Request) {
	var req BuildReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed decoding json"})
		return
	}

	paths, err := a.db.GetPaths(req.CampaingID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed getting paths"})
		return
	}

	if len(paths) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "No http paths found for given campaingID"})
		return
	}

	fmt.Printf("%s\n", a.BuilderInt.BackendURL)
	fmt.Printf("%d\n", a.BuilderInt.BackendPort)
	w.WriteHeader(http.StatusOK)

}

func (a *API) ListBuilds(w http.ResponseWriter, r *http.Request) {

}
