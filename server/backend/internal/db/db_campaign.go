package db

import (
	"database/sql"

	"github.com/google/uuid"
)

func (db *DB) InsertNewCampaign(name string, paths []string) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	campaignID := id.String()
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := `INSERT INTO campaigns(Name, uuid) VALUES(?,?)`
	_, err = tx.Exec(q, name, campaignID)
	if err != nil {
		return err
	}
	q = `INSERT INTO campaign_paths(uuid, Path) VALUES (?,?)`
	for _, path := range paths {
		_, err := tx.Exec(q, campaignID, path)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

type Campaigns struct {
	Total int        `json:"total"`
	Entry []Campaign `json:"campaigns"`
}

type Campaign struct {
	Name         string   `json:"name"`
	Id           string   `json:"id"`
	TotalClients int      `json:"total_clients"`
	TotalFiles   int      `json:"total_files"`
	Paths        []string `json:"paths"`
	CreatedAt    string   `json:"created_at"`
}

func (db *DB) ListCampaigns() (*Campaigns, error) {

	q := `SELECT c.Name, c.uuid,
		(SELECT COUNT(*) FROM clients AS cl WHERE cl.CampaignUUID = c.uuid),
		(SELECT COUNT(*) FROM user_uploads AS f WHERE f.CampaignUUID = c.uuid),
		c.created_at, cp.Path
		FROM campaigns AS c
		LEFT JOIN campaign_paths AS cp ON cp.uuid = c.uuid
		ORDER BY c.id, cp.id`

	rows, err := db.conn.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []Campaign
	indexByID := make(map[string]int)
	for rows.Next() {
		var c Campaign
		var path sql.NullString
		if err := rows.Scan(&c.Name, &c.Id, &c.TotalClients, &c.TotalFiles, &c.CreatedAt, &path); err != nil {
			return nil, err
		}
		index, exists := indexByID[c.Id]
		if !exists {
			c.Paths = []string{}
			campaigns = append(campaigns, c)
			index = len(campaigns) - 1
			indexByID[c.Id] = index
		}
		if path.Valid {
			campaigns[index].Paths = append(campaigns[index].Paths, path.String)
		}
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
		`DELETE FROM campaign_paths WHERE uuid = ?`,
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

func (db *DB) GetCampaignIdFromPath(path string) (string, error) {
	var uuid string
	q := `SELECT c.uuid FROM campaign_paths AS cp JOIN campaigns AS c ON c.uuid = cp.uuid WHERE cp.Path = ?`
	err := db.conn.QueryRow(q, path).Scan(&uuid)
	return uuid, err

}
