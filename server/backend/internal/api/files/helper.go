package files

import (
	"fmt"
	"os"
	"path/filepath"
	"styx/internal/db"

	"github.com/google/uuid"
)

func (a *API) Insert(data db.FileInfo, id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.uploadsMap == nil {
		a.uploadsMap = make(map[string]db.FileInfo)
	}
	a.uploadsMap[id] = data

}

func (a *API) Delete(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	entry, ok := a.uploadsMap[id]
	if !ok {
		return nil
	}
	f := entry.Handle
	delete(a.uploadsMap, id)
	if f == nil {
		return nil
	}
	return f.Close()
}

func (a *API) Get(id string) (db.FileInfo, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	data, ok := a.uploadsMap[id]
	return data, ok
}

func (a *API) PrepeareDisk(info db.FileInfo) (string, error) {
	if info.FileName == "" || info.FileName == "." || info.FileName == ".." || filepath.Base(info.FileName) != info.FileName {
		return "", fmt.Errorf("invalid file name")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	uploadID := id.String()

	userfilesDir := filepath.Join(home, ".local", "share", "styx", "user-files", uploadID)
	if err := os.Mkdir(userfilesDir, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(userfilesDir, info.FileName)
	f, err := os.Create(filePath)
	if err != nil {
		os.Remove(userfilesDir)
		return "", err
	}

	info.Handle = f
	info.FileUUID = uploadID

	a.Insert(info, uploadID)

	return uploadID, nil
}

func (a *API) DeleteLocal(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid file id")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if err := a.Delete(id); err != nil {
		return err
	}

	path := filepath.Join(home, ".local", "share", "styx", "user-files", id)
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return a.db.DeleteFile(id)
}
