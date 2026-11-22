package db

import (
	"strings"

	"styx/internal/packager"
)

func (db *DB) ClientExists(clientID int32) (bool, error) {
	var exists bool
	err := db.conn.QueryRow(`SELECT EXISTS(SELECT 1 FROM clients WHERE ClientID = ?)`, clientID).Scan(&exists)
	return exists, err
}

func (db *DB) ClientInCampaign(clientID int32, campaignID string) (bool, error) {
	var exists bool
	err := db.conn.QueryRow(`SELECT EXISTS(SELECT 1 FROM clients WHERE ClientID = ? AND CampaignUUID = ?)`, clientID, campaignID).Scan(&exists)
	return exists, err
}

type Client struct {
	Name      string `json:"name"`
	ID        int32  `json:"client_id"`
	Username  string `json:"username"`
	Hostname  string `json:"hostname"`
	Domain    string `json:"domain"`
	Country   string `json:"country"`
	IsAdmin   bool   `json:"is_admin"`
	CreatedAt string `json:"created_at"`
}

type Clients struct {
	Total  int      `json:"total"`
	Entrys []Client `json:"clients"`
}

// cleanClient strips NUL terminators written by older agent builds.
func cleanClient(c Client) Client {
	c.Name = strings.TrimRight(c.Name, "\x00")
	c.Username = strings.TrimRight(c.Username, "\x00")
	c.Hostname = strings.TrimRight(c.Hostname, "\x00")
	c.Domain = strings.TrimRight(c.Domain, "\x00")
	c.Country = strings.TrimRight(c.Country, "\x00")
	return c
}

func (db *DB) ListClients(id string) (*Clients, error) {
	q := `SELECT ClientName, ClientID, Username, Hostname, Domain, Country, IsAdmin, created_at FROM clients WHERE CampaignUUID = ?`

	rows, err := db.conn.Query(q, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []Client

	for rows.Next() {
		var c Client
		var isAdmin int
		if err := rows.Scan(&c.Name, &c.ID, &c.Username, &c.Hostname, &c.Domain, &c.Country, &isAdmin, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.IsAdmin = isAdmin != 0
		c = cleanClient(c)

		clients = append(clients, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &Clients{Total: len(clients), Entrys: clients}, nil
}

func (db *DB) InsertAgent(info packager.AgentRegister, campaignID string) (Client, error) {
	q := `INSERT INTO Clients(ClientName, ClientID, IsAdmin, Pid, PPid, CampaignUUID, Hostname, Username, Domain, Country, Arch) VALUES(?,?,?,?,?,?,?,?,?,?,?) RETURNING ClientName, ClientID, Username, Hostname, Domain, Country, IsAdmin, created_at`

	var client Client
	var isAdmin int
	err := db.conn.QueryRow(q, GenerateAgentName(), info.AgentID, info.IsAdmin, info.Pid, info.PPid, campaignID, info.Hostname, info.Username, info.Domain, info.Country, info.Arch).
		Scan(&client.Name, &client.ID, &client.Username, &client.Hostname, &client.Domain, &client.Country, &isAdmin, &client.CreatedAt)
	if err != nil {
		return Client{}, err
	}
	client.IsAdmin = isAdmin != 0
	return cleanClient(client), nil
}
