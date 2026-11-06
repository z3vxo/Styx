package db

type NewTask struct {
	CampaignID string `json:"campaign_id"`
	ClientID   string `json:"client_id"`
	TaskName   string `json:"task_name"`
	Args       []Arg  `json:"args"`
}

type Arg struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// var CmdToCodeMap = map[string]int {
// 	""
// }

/*
 * run = run exe/dll/shellcode
 * load-module <name> = load specific module
 * unload-module <name> = unload specfic module
 * ls/cp/rm/cat/whoami = as the command states
 *
 */
func (db *DB) InsertTask(info NewTask) error {
	return nil
}
