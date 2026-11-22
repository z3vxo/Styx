package db

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestInsertTaskPersistsResolvedFields(t *testing.T) {
	conn, err := Setup(filepath.Join(t.TempDir(), "styx.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	database := &DB{conn: conn}

	moduleID, code := int32(1), int32(1)
	req := NewTaskReq{
		Cmd:        "ls",
		Args:       []Arg{{Type: "string", Value: "/tmp"}},
		ClientID:   42,
		CampaignID: "campaign-a",
		Kind:       "run-command",
		Module:     "file-system",
		ModuleID:   &moduleID,
		Command:    "ls",
		Code:       &code,
	}
	taskID, err := database.InsertTask(req)
	if err != nil {
		t.Fatal(err)
	}
	if taskID <= 0 {
		t.Fatalf("task ID = %d", taskID)
	}

	var (
		gotTaskID   int32
		gotClientID int32
		campaignID  string
		kind        string
		module      string
		gotModuleID int64
		command     string
		gotCode     int64
		args        string
		status      string
	)
	err = conn.QueryRow(`SELECT TaskID, ClientID, CampaignID, Kind, Module, ModuleID, Command, Code, Args, Status FROM tasks WHERE TaskID = ?`, taskID).
		Scan(&gotTaskID, &gotClientID, &campaignID, &kind, &module, &gotModuleID, &command, &gotCode, &args, &status)
	if err != nil {
		t.Fatal(err)
	}
	if gotTaskID != taskID || gotClientID != 42 || campaignID != "campaign-a" {
		t.Fatalf("row identity = %d/%d/%q", gotTaskID, gotClientID, campaignID)
	}
	if kind != "run-command" || module != "file-system" || gotModuleID != 1 || command != "ls" || gotCode != 1 {
		t.Fatalf("resolved fields = %q/%q/%d/%q/%d", kind, module, gotModuleID, command, gotCode)
	}
	if status != "pending" {
		t.Fatalf("status = %q", status)
	}
	var gotArgs []Arg
	if err := json.Unmarshal([]byte(args), &gotArgs); err != nil {
		t.Fatal(err)
	}
	if len(gotArgs) != 1 || gotArgs[0].Type != "string" || gotArgs[0].Value != "/tmp" {
		t.Fatalf("args = %+v", gotArgs)
	}
}

func TestGetTasksReturnsAndMarksSent(t *testing.T) {
	conn, err := Setup(filepath.Join(t.TempDir(), "styx.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	database := &DB{conn: conn}

	if _, err := database.InsertTask(NewTaskReq{
		Cmd: "list-modules", ClientID: 42, CampaignID: "campaign-a", Kind: "list-modules",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := database.InsertTask(NewTaskReq{
		Cmd: "list-modules", ClientID: 7, CampaignID: "campaign-a", Kind: "list-modules",
	}); err != nil {
		t.Fatal(err)
	}

	tasks, err := database.GetTasks(42)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Fatalf("tasks = %+v", tasks)
	}
	if tasks[0].Kind != "list-modules" || tasks[0].ClientID != 42 || tasks[0].CampaignID != "campaign-a" {
		t.Fatalf("task = %+v", tasks[0])
	}

	var status string
	if err := conn.QueryRow(`SELECT Status FROM tasks WHERE TaskID = ?`, tasks[0].TaskID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "sent" {
		t.Fatalf("status = %q, want sent", status)
	}

	again, err := database.GetTasks(42)
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != 0 {
		t.Fatalf("tasks replayed: %+v", again)
	}

	other, err := database.GetTasks(7)
	if err != nil {
		t.Fatal(err)
	}
	if len(other) != 1 {
		t.Fatalf("other client tasks = %+v", other)
	}
}

func TestInsertExitAgentTask(t *testing.T) {
	conn, err := Setup(filepath.Join(t.TempDir(), "styx.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	database := &DB{conn: conn}

	code := int32(0xf1)
	taskID, err := database.InsertTask(NewTaskReq{
		Cmd: "exit", ClientID: 42, CampaignID: "campaign-a", Kind: "exit-agent", Code: &code,
	})
	if err != nil {
		t.Fatal(err)
	}

	var kind string
	var gotCode int32
	if err := conn.QueryRow(`SELECT Kind, Code FROM tasks WHERE TaskID = ?`, taskID).Scan(&kind, &gotCode); err != nil {
		t.Fatal(err)
	}
	if kind != "exit-agent" || gotCode != code {
		t.Fatalf("task = %q/%#x, want exit-agent/%#x", kind, gotCode, code)
	}
}
