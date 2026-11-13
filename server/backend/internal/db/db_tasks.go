package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"styx/internal/packager"
	"styx/internal/shared"
)

type NewTaskReq struct {
	Cmd        string `json:"cmd"`
	Args       []Arg  `json:"args"`
	ClientID   int32  `json:"client_id"`
	CampaignID string `json:"campaign_id"`

	Kind     string `json:"-"`
	Module   string `json:"-"`
	ModuleID *int32 `json:"-"`
	Command  string `json:"-"`
	Code     *int32 `json:"-"`
}

type Arg struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

func (db *DB) InsertTask(task NewTaskReq) (int32, error) {
	args := task.Args
	if args == nil {
		args = []Arg{}
	}
	jsonArgs, err := json.Marshal(args)
	if err != nil {
		return 0, err
	}

	var module any
	if task.Module != "" {
		module = task.Module
	}
	var command any
	if task.Command != "" {
		command = task.Command
	}
	var moduleID any
	if task.ModuleID != nil {
		moduleID = *task.ModuleID
	}
	var code any
	if task.Code != nil {
		code = *task.Code
	}

	q := `INSERT INTO tasks(TaskID, ClientID, CampaignID, Kind, Module, ModuleID, Command, Code, Args) VALUES(?,?,?,?,?,?,?,?,?)`
	for attempt := 0; attempt < 5; attempt++ {
		taskID, err := shared.GenRandInt32()
		if err != nil {
			return 0, err
		}
		_, err = db.conn.Exec(q, taskID, task.ClientID, task.CampaignID, task.Kind, module, moduleID, command, code, string(jsonArgs))
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				continue
			}
			return 0, err
		}
		return taskID, nil
	}
	return 0, fmt.Errorf("could not allocate a unique task id")
}

func (db *DB) GetTasks(agentID int32) ([]packager.TaskRecord, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`SELECT id, TaskID, ClientID, CampaignID, Kind, Module, ModuleID, Command, Code, Args, Status FROM tasks WHERE ClientID = ? AND Status = 'pending' ORDER BY id`, agentID)
	if err != nil {
		return nil, err
	}

	var entry []packager.TaskRecord
	var ids []int64
	for rows.Next() {
		var task packager.TaskRecord
		var id int64
		var module, command sql.NullString
		var moduleID, code sql.NullInt64
		if err := rows.Scan(&id, &task.TaskID, &task.ClientID, &task.CampaignID, &task.Kind, &module, &moduleID, &command, &code, &task.Args, &task.Status); err != nil {
			rows.Close()
			return nil, err
		}
		if module.Valid {
			task.Module = module.String
		}
		if moduleID.Valid {
			task.ModuleID = int32(moduleID.Int64)
		}
		if command.Valid {
			task.Command = command.String
		}
		if code.Valid {
			task.Code = int32(code.Int64)
		}
		entry = append(entry, task)
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	for _, id := range ids {
		if _, err := tx.Exec(`UPDATE tasks SET Status = 'sent', sent_at = CURRENT_TIMESTAMP WHERE id = ?`, id); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return entry, nil
}
