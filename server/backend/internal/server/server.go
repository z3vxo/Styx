package server

import (
	"fmt"
	"net/http"
	"styx/internal/agentsrv"
	campaigns "styx/internal/api/Campaigns"
	authapi "styx/internal/api/auth"
	clients "styx/internal/api/clients"
	files "styx/internal/api/files"
	tasks "styx/internal/api/tasks"
	authmgr "styx/internal/auth"
	"styx/internal/cfg"
	"styx/internal/db"
	"styx/internal/modules"
	"styx/internal/web"
	"styx/internal/ws"

	"github.com/go-chi/chi"
)

type Server struct {
	DB          *db.DB
	modules     *modules.Registry
	Auth        *authmgr.Auth
	AuthAPI     *authapi.API
	CampaignAPI *campaigns.API
	UploadsAPI  *files.API
	ClientsAPI  *clients.API
	TasksAPI    *tasks.API
	Cfg         *cfg.Cfg
	WSHandler   *ws.WS
	AgentSrv    *agentsrv.Agent
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
	a := authmgr.New(c)

	m, err := modules.New()
	if err != nil {
		return nil, err
	}

	return &Server{
		DB:          db,
		modules:     m,
		Cfg:         c,
		Auth:        a,
		AuthAPI:     authapi.New(db, a),
		CampaignAPI: campaigns.New(db, a),
		UploadsAPI:  files.New(db, a),
		ClientsAPI:  clients.New(db, a),
		TasksAPI:    tasks.New(db, a, m),
		WSHandler:   ws.New(db, a),
		AgentSrv:    agentsrv.New(db, m),
	}, nil
}

func (s *Server) Run() error {
	r := chi.NewRouter()

	r.Route("/rest", func(r chi.Router) {
		r.Mount("/auth", s.AuthAPI.Routes())
		r.Mount("/campaigns", s.CampaignAPI.Routes())
		r.Mount("/uploads", s.UploadsAPI.Routes())
		r.Mount("/clients", s.ClientsAPI.Routes())
		r.Mount("/tasks", s.TasksAPI.Routes())
		r.Get("/ws/connect", s.WSHandler.HandleWS)
	})
	r.HandleFunc("/*", s.AgentSrv.HandleAgentSrv)

	r.Handle("/login", web.Handler())
	r.Handle("/dashboard", web.Handler())
	r.Handle("/assets/*", web.Handler())

	listen := fmt.Sprintf("%s:%d", s.Cfg.Bind, s.Cfg.Port)

	server := &http.Server{Addr: listen, Handler: r}

	fmt.Printf("[*] Listening on %s\n", listen)
	return server.ListenAndServe()

}
