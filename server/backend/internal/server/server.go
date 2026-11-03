package server

import (
	"fmt"
	"net/http"
	campaigns "styx/internal/api/Campaigns"
	authapi "styx/internal/api/auth"
	"styx/internal/auth"
	"styx/internal/cfg"
	"styx/internal/db"

	"github.com/go-chi/chi"
)

type Server struct {
	DB          *db.DB
	Auth        *auth.Auth
	AuthAPI     *authapi.API
	CampaignAPI *campaigns.API
	Cfg         *cfg.Cfg
}

func New() (*Server, error) {
	db, err := db.New()
	if err != nil {
		return nil, err
	}

	c, err := cfg.Load()
	if err != nil {
		return nil, err
	}
	a := auth.New(c)

	return &Server{
		DB:          db,
		Cfg:         c,
		Auth:        a,
		AuthAPI:     authapi.New(db, a),
		CampaignAPI: campaigns.New(db),
	}, nil
}

func (s *Server) Run() error {
	r := chi.NewRouter()

	r.Route("/rest", func(r chi.Router) {
		r.Mount("/auth", s.AuthAPI.Routes())
		//r.Get("/ws/connect")
		r.Mount("/campaigns", s.CampaignAPI.Routes())
	})

	listen := fmt.Sprintf("%s:%d", s.Cfg.Bind, s.Cfg.Port)

	server := &http.Server{Addr: listen, Handler: r}

	fmt.Printf("[*] Listening on %s\n", listen)
	return server.ListenAndServe()

}
