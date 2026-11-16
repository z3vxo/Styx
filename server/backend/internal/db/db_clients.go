package db

import "styx/internal/packager"

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
	Hostname  string `json:"hostname"`
	Domain    string `json:"domain"`
	CreatedAt string `json:"created_at"`
}

type Clients struct {
	Total  int      `json:"total"`
	Entrys []Client `json:"clients"`
}

func (db *DB) ListClients(id string) (*Clients, error) {
	q := `SELECT ClientName, ClientID, Hostname, Domain, created_at FROM clients WHERE CampaignUUID = ?`

	rows, err := db.conn.Query(q, id)
	if err != nil {
		return nil, err
	}

	var clients []Client

	for rows.Next() {
		var c Client
		if err := rows.Scan(&c.Name, &c.ID, &c.Hostname, &c.Domain, &c.CreatedAt); err != nil {
			return nil, err
		}

		clients = append(clients, c)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return &Clients{Total: len(clients), Entrys: clients}, nil
}

func (db *DB) InsertAgent(info packager.AgentRegister, campaignID string) (Client, error) {
	q := `INSERT INTO Clients(ClientName, ClientID, IsAdmin, Pid, PPid, CampaignUUID, Hostname, Username, Domain, Country, Arch) VALUES(?,?,?,?,?,?,?,?,?,?,?) RETURNING ClientName, ClientID, Hostname, Domain, created_at`

	var client Client
	err := db.conn.QueryRow(q, GenerateAgentName(), info.AgentID, info.IsAdmin, info.Pid, info.PPid, campaignID, info.Hostname, info.Username, info.Domain, info.Country, info.Arch).
		Scan(&client.Name, &client.ID, &client.Hostname, &client.Domain, &client.CreatedAt)
	if err != nil {
		return Client{}, err
	}
	return client, nil
}
