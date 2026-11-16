package db

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"styx/internal/packager"
)

func TestGenerateAgentName(t *testing.T) {
	for i := 0; i < 200; i++ {
		name := GenerateAgentName()
		figure, symbol, ok := strings.Cut(name, "-")
		if !ok || !slices.Contains(agentNameFigures, figure) || !slices.Contains(agentNameSymbols, symbol) {
			t.Fatalf("GenerateAgentName() = %q", name)
		}
	}
}

func TestInsertAgentStoresCampaignAndName(t *testing.T) {
	conn, err := Setup(filepath.Join(t.TempDir(), "styx.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	database := &DB{conn: conn}
	info := packager.AgentRegister{AgentID: 42, Hostname: "host-a", Domain: "example.com"}
	inserted, err := database.InsertAgent(info, "campaign-a")
	if err != nil {
		t.Fatal(err)
	}
	if inserted.ID != 42 || inserted.Hostname != "host-a" || inserted.Domain != "example.com" || inserted.CreatedAt == "" {
		t.Fatalf("inserted = %+v", inserted)
	}
	figure, symbol, ok := strings.Cut(inserted.Name, "-")
	if !ok || !slices.Contains(agentNameFigures, figure) || !slices.Contains(agentNameSymbols, symbol) {
		t.Fatalf("inserted name = %q", inserted.Name)
	}

	clients, err := database.ListClients("campaign-a")
	if err != nil {
		t.Fatal(err)
	}
	if clients.Total != 1 || len(clients.Entrys) != 1 {
		t.Fatalf("clients = %+v", clients)
	}
	client := clients.Entrys[0]
	if client != inserted {
		t.Fatalf("stored client = %+v, inserted = %+v", client, inserted)
	}

	other, err := database.ListClients("campaign-b")
	if err != nil {
		t.Fatal(err)
	}
	if other.Total != 0 {
		t.Fatalf("agent stored under the wrong campaign: %+v", other)
	}
}
