package agentsrv

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"styx/internal/db"
	"styx/internal/packager"
)

type Agent struct {
	db *db.DB
}

func New(db *db.DB) *Agent {
	return &Agent{db: db}
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

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	if len(body) < 12 {
		http.Error(w, "error: missing metadata", http.StatusBadRequest)
		return
	}
	reader := bytes.NewReader(body)

	AgentID, err := packager.ParseIncomingMetaDataAndValidate(reader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	fmt.Printf("%d\n", AgentID)

	fmt.Println(r.URL.Path)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("DONE"))
}
