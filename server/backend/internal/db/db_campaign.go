package db

import "github.com/google/uuid"

func (db *DB) InsertNewCampaign(name string) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	uuid := id.String()

	q := `INSERT INTO campaigns(Name, uuid) VALUES(?,?)`
	_, err = db.conn.Exec(q, name, uuid)
	return err
}

type Campaigns struct {
	Total int        `json:"total"`
	Entry []Campaign `json:"campaigns"`
}

type Campaign struct {
	Name      string `json:"name"`
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
}

func (db *DB) ListCampaigns() (*Campaigns, error) {

	q := `SELECT Name, uuid, created_at FROM campaigns`

	rows, err := db.conn.Query(q)
	if err != nil {
		return nil, err
	}

	var campaigns []Campaign
	for rows.Next() {
		var c Campaign
		if err := rows.Scan(&c.Name, &c.Id, &c.CreatedAt); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return &Campaigns{Total: len(campaigns), Entry: campaigns}, nil

}

func (db *DB) DeleteCampaign(id string) error {
	_, err := db.conn.Exec(`DELETE FROM campaigns WHERE uuid = ?`, id)
	return err
}
