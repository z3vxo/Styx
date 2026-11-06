package db

// CREATE TABLE IF NOT EXISTS clients(
//
//	id INTEGER PRIMARY KEY AUTOINCREMENT,
//	ClientName TEXT NOT NULL,
//	ClientID   INTEGER NOT NULL,
//	CampaignUUID TEXT NOT NULL,
//	Hostname TEXT NOT NULL DEFAULT '',
//	Domain TEXT NOT NULL DEFAULT ''
//
// created_at,
//
//		FOREIGN KEY(CampaignUUID) REFERENCES campaigns(uuid) ON DELETE CASCADE
//	);`,
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
