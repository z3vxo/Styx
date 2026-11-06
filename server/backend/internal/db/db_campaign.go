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
	Name         string `json:"name"`
	Id           string `json:"id"`
	TotalClients int    `json:"total_clients"`
	TotalFiles   int    `json:"total_files"`
	//TotalBuilds  int    `json:"total_builds"`
	CreatedAt string `json:"created_at"`
}

func (db *DB) ListCampaigns() (*Campaigns, error) {

	q := `SELECT c.Name, c.uuid,
		(SELECT COUNT(*) FROM clients AS cl WHERE cl.CampaignUUID = c.uuid),
		(SELECT COUNT(*) FROM user_uploads AS f WHERE f.CampaignUUID = c.uuid),
		c.created_at
		FROM campaigns AS c`

	rows, err := db.conn.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []Campaign
	for rows.Next() {
		var c Campaign
		if err := rows.Scan(&c.Name, &c.Id, &c.TotalClients, &c.TotalFiles, &c.CreatedAt); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &Campaigns{Total: len(campaigns), Entry: campaigns}, nil

}

func (db *DB) DeleteCampaign(id string) ([]string, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRow(`SELECT 1 FROM campaigns WHERE uuid = ?`, id).Scan(&exists); err != nil {
		return nil, err
	}
	rows, err := tx.Query(`SELECT FileUUID FROM user_uploads WHERE CampaignUUID = ?`, id)
	if err != nil {
		return nil, err
	}
	var fileIDs []string
	for rows.Next() {
		var fileID string
		if err := rows.Scan(&fileID); err != nil {
			rows.Close()
			return nil, err
		}
		fileIDs = append(fileIDs, fileID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	for _, query := range []string{
		`DELETE FROM user_uploads WHERE CampaignUUID = ?`,
		`DELETE FROM clients WHERE CampaignUUID = ?`,
		`DELETE FROM campaigns WHERE uuid = ?`,
	} {
		if _, err := tx.Exec(query, id); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return fileIDs, nil
}
