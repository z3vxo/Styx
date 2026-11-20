package tasks

import (
	"testing"

	"styx/internal/db"
	"styx/internal/modules"
)

func TestResolveBuiltinsAndModuleCommands(t *testing.T) {
	api := &API{modules: modules.Default()}

	load := db.NewTaskReq{Cmd: "load-module", Args: []db.Arg{{Type: "string", Value: "file-system"}}}
	if err := api.resolve(&load); err != nil {
		t.Fatal(err)
	}
	if load.Kind != "load-module" || load.Module != "file-system" || load.ModuleID == nil || *load.ModuleID != 1 {
		t.Fatalf("load = %+v", load)
	}

	unload := db.NewTaskReq{Cmd: "unload-module", Args: []db.Arg{{Type: "string", Value: "file-system"}}}
	if err := api.resolve(&unload); err != nil {
		t.Fatal(err)
	}
	if unload.Kind != "unload-module" || unload.Module != "file-system" || unload.ModuleID == nil || *unload.ModuleID != 1 {
		t.Fatalf("unload = %+v", unload)
	}

	unknownUnload := db.NewTaskReq{Cmd: "unload-module", Args: []db.Arg{{Type: "string", Value: "missing"}}}
	if err := api.resolve(&unknownUnload); err == nil {
		t.Fatal("unknown module unloaded")
	}

	list := db.NewTaskReq{Cmd: "list-modules"}
	if err := api.resolve(&list); err != nil {
		t.Fatal(err)
	}
	if list.Kind != "list-modules" || list.ModuleID != nil {
		t.Fatalf("list = %+v", list)
	}

	bare := db.NewTaskReq{Cmd: "ls"}
	if err := api.resolve(&bare); err != nil {
		t.Fatal(err)
	}
	if bare.Kind != "run-command" || bare.Module != "file-system" || bare.Command != "ls" || bare.Code == nil || *bare.Code != 1 {
		t.Fatalf("bare = %+v", bare)
	}

	explicit := db.NewTaskReq{Cmd: "run-command", Args: []db.Arg{{Type: "string", Value: "file-system"}, {Type: "string", Value: "cat"}, {Type: "string", Value: "C:\\"}}}
	if err := api.resolve(&explicit); err != nil {
		t.Fatal(err)
	}
	if explicit.Kind != "run-command" || explicit.Command != "cat" || explicit.Code == nil || *explicit.Code != 4 {
		t.Fatalf("explicit = %+v", explicit)
	}

	unknown := db.NewTaskReq{Cmd: "nope"}
	if err := api.resolve(&unknown); err == nil {
		t.Fatal("unknown command resolved")
	}
}

func TestResolveValidatesArguments(t *testing.T) {
	api := &API{modules: modules.Default()}

	if err := api.resolve(&db.NewTaskReq{Cmd: "cat"}); err == nil {
		t.Fatal("cat accepted without a path")
	}
	if err := api.resolve(&db.NewTaskReq{Cmd: "cat", Args: []db.Arg{{Type: "string", Value: 1}}}); err == nil {
		t.Fatal("cat accepted a non-string path")
	}
	if err := api.resolve(&db.NewTaskReq{Cmd: "ls", Args: []db.Arg{{Type: "string", Value: "/"}, {Type: "string", Value: "x"}}}); err == nil {
		t.Fatal("ls accepted extra arguments")
	}
	if err := api.resolve(&db.NewTaskReq{Cmd: "pwd", Args: []db.Arg{{Type: "string", Value: "x"}}}); err == nil {
		t.Fatal("pwd accepted arguments")
	}
	if err := api.resolve(&db.NewTaskReq{Cmd: "load-module", Args: []db.Arg{{Type: "string", Value: "file-system"}, {Type: "string", Value: "x"}}}); err == nil {
		t.Fatal("load-module accepted extra arguments")
	}

	ls := db.NewTaskReq{Cmd: "ls"}
	if err := api.resolve(&ls); err != nil {
		t.Fatal(err)
	}
	if len(ls.Args) != 1 || ls.Args[0].Value != "." {
		t.Fatalf("ls default args = %+v", ls.Args)
	}

	registry := &modules.Registry{Modules: map[string]modules.Modules{
		"test": {
			ID: 2,
			Commands: map[string]modules.Command{
				"seek": {Code: 1, Args: []modules.CommandArg{{Name: "offset", Type: "int32", Required: true}}},
			},
		},
	}}
	intAPI := &API{modules: registry}
	if err := intAPI.resolve(&db.NewTaskReq{Cmd: "seek", Args: []db.Arg{{Type: "string", Value: "1"}}}); err == nil {
		t.Fatal("seek accepted a string offset")
	}
	seek := db.NewTaskReq{Cmd: "seek", Args: []db.Arg{{Type: "int32", Value: float64(1)}}}
	if err := intAPI.resolve(&seek); err != nil {
		t.Fatal(err)
	}
}
