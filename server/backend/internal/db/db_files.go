package db

import "os"

type FileInfo struct {
	FileName     string `json:"file_name"`
	CampaignUUID string `json:"campaign_uuid"`
	FileSize     int64  `json:"file_size"`
	FileType     string `json:"file_type"`
	FileUUID     string
	Handle       *os.File
}

type Files struct {
	Total int    `json:"total"`
	Entry []File `json:"files"`
}

type File struct {
	FileName   string `json:"file_name"`
	FileSize   int64  `json:"file_size"`
	FileType   string `json:"file_type"`
	FileUUID   string `json:"file_uuid"`
	UploadedAt string `json:"uploaded_at"`
}

func (db *DB) InsertFile(info FileInfo) error {
	q := `INSERT INTO user_uploads(FileUUID, FileName, FileSize, FileType, HasReloc,CampaignUUID) VALUES(?,?,?,?,?,?)`
	_, err := db.conn.Exec(q, info.FileUUID, info.FileName, info.FileSize, info.FileType, 1, info.CampaignUUID)
	return err
}

func (db *DB) DeleteFile(id string) error {
	_, err := db.conn.Exec(`DELETE FROM user_uploads WHERE fileUUID = ?`, id)
	return err
}

func (db *DB) ListUploads(id string) (*Files, error) {
	q := `SELECT FileUUID, FileName, FileSize, FileType, uploaded_at FROM user_uploads WHERE CampaignUUID = ?`

	rows, err := db.conn.Query(q, id)
	if err != nil {
		return nil, err
	}

	var files []File
	for rows.Next() {
		var f File
		if err := rows.Scan(&f.FileUUID, &f.FileName, &f.FileSize, &f.FileType, &f.UploadedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return &Files{Total: len(files), Entry: files}, nil
}
