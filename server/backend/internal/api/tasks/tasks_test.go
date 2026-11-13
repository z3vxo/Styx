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

	explicit := db.NewTaskReq{Cmd: "run-command", Args: []db.Arg{{Type: "string", Value: "file-system"}, {Type: "string", Value: "cat"}}}
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
