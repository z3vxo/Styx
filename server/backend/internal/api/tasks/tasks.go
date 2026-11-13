package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	authmgr "styx/internal/auth"
	"styx/internal/db"
	"styx/internal/modules"

	"github.com/go-chi/chi"
)

type API struct {
	db      *db.DB
	auth    *authmgr.Auth
	modules *modules.Registry
}

func New(database *db.DB, a *authmgr.Auth, m *modules.Registry) *API {
	return &API{db: database, auth: a, modules: m}
}

func (a *API) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(a.auth.AuthWare())
	r.Post("/new", a.NewTask)
	return r
}

func (a *API) NewTask(w http.ResponseWriter, r *http.Request) {
	var req db.NewTaskReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Cmd == "" || req.ClientID <= 0 || req.CampaignID == "" {
		writeError(w, http.StatusBadRequest, "cmd, client_id and campaign_id are required")
		return
	}
	if err := a.resolve(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Kind == "help" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"help": "help, load-module <name>, unload-module <name>, list-modules, run-command <module> <command> [args]"})
		return
	}

	taskID, err := a.db.InsertTask(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int32{"task_id": taskID})
}

func (a *API) resolve(req *db.NewTaskReq) error {
	switch req.Cmd {
	case "help":
		req.Kind = "help"
	case "load-module", "unload-module":
		name, ok := stringArg(req.Args, 0)
		if !ok {
			return fmt.Errorf("%s requires a module name", req.Cmd)
		}
		module, ok := a.modules.Module(name)
		if !ok {
			return fmt.Errorf("unknown module %q", name)
		}
		moduleID := int32(module.ID)
		req.Kind = req.Cmd
		req.Module = name
		req.ModuleID = &moduleID
	case "list-modules":
		req.Kind = "list-modules"
	case "run-command":
		moduleName, ok := stringArg(req.Args, 0)
		if !ok {
			return fmt.Errorf("run-command requires a module name")
		}
		command, ok := stringArg(req.Args, 1)
		if !ok {
			return fmt.Errorf("run-command requires a command")
		}
		module, ok := a.modules.Module(moduleName)
		if !ok {
			return fmt.Errorf("unknown module %q", moduleName)
		}
		code, ok := module.Commands[command]
		if !ok {
			return fmt.Errorf("unknown command %q in module %q", command, moduleName)
		}
		moduleID := int32(module.ID)
		commandCode := int32(code)
		req.Kind = "run-command"
		req.Module = moduleName
		req.ModuleID = &moduleID
		req.Command = command
		req.Code = &commandCode
		req.Args = req.Args[2:]
	default:
		moduleName, module, code, ok := a.modules.ResolveCommand(req.Cmd)
		if !ok {
			return fmt.Errorf("unknown command %q", req.Cmd)
		}
		moduleID := int32(module.ID)
		commandCode := int32(code)
		req.Kind = "run-command"
		req.Module = moduleName
		req.ModuleID = &moduleID
		req.Command = req.Cmd
		req.Code = &commandCode
	}
	return nil
}

func stringArg(args []db.Arg, index int) (string, bool) {
	if index < 0 || index >= len(args) {
		return "", false
	}
	value, ok := args[index].Value.(string)
	if !ok || value == "" {
		return "", false
	}
	return value, true
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
