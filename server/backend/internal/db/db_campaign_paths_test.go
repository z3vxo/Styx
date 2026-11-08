package db

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestInsertNewCampaignStoresPathsAtomically(t *testing.T) {
	conn, err := Setup(filepath.Join(t.TempDir(), "styx.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	database := &DB{conn: conn}

	if err := database.InsertNewCampaign("first", []string{"/alpha", "/beta"}); err != nil {
		t.Fatal(err)
	}
	var campaignID string
	if err := conn.QueryRow(`SELECT uuid FROM campaigns WHERE Name = ?`, "first").Scan(&campaignID); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/alpha", "/beta"} {
		got, err := database.GetCampaignIdFromPath(path)
		if err != nil || got != campaignID {
			t.Fatalf("path %q resolves to %q, err = %v; want %q", path, got, err, campaignID)
		}
	}
	if _, err := conn.Exec(`INSERT INTO campaigns (Name, uuid) VALUES (?, ?)`, "legacy", "legacy-id"); err != nil {
		t.Fatal(err)
	}
	listed, err := database.ListCampaigns()
	if err != nil {
		t.Fatal(err)
	}
	if listed.Total != 2 || len(listed.Entry) != 2 {
		t.Fatalf("campaign list has total %d and %d entries; want 2", listed.Total, len(listed.Entry))
	}
	if !reflect.DeepEqual(listed.Entry[0].Paths, []string{"/alpha", "/beta"}) {
		t.Fatalf("stored paths = %v, want [/alpha /beta]", listed.Entry[0].Paths)
	}
	if listed.Entry[1].Paths == nil || len(listed.Entry[1].Paths) != 0 {
		t.Fatalf("legacy paths = %v, want empty array", listed.Entry[1].Paths)
	}

	if _, err := conn.Exec(`CREATE TRIGGER reject_path BEFORE INSERT ON campaign_paths
		WHEN NEW.Path = '/broken' BEGIN SELECT RAISE(ABORT, 'rejected path'); END`); err != nil {
		t.Fatal(err)
	}
	if err := database.InsertNewCampaign("second", []string{"/ok", "/broken"}); err == nil {
		t.Fatal("expected path insert to fail")
	}
	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM campaigns WHERE Name = ?`, "second").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("failed campaign left %d campaign rows", count)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM campaign_paths WHERE Path = ?`, "/ok").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("failed campaign left %d path rows", count)
	}

	if _, err := database.DeleteCampaign(campaignID); err != nil {
		t.Fatal(err)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM campaign_paths WHERE uuid = ?`, campaignID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("deleted campaign left %d path rows", count)
	}
}
