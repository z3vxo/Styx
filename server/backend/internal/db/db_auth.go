package db

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

func (db *DB) VerifyAuth(user, pass string) error {
	var encoded string
	if err := db.conn.QueryRow(`SELECT Pass FROM users WHERE User = ?`, user).Scan(&encoded); err != nil {
		return err
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" {
		return fmt.Errorf("invalid stored password hash")
	}
	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil ||
		parts[3] != fmt.Sprintf("m=%d,t=%d,p=%d", memory, iterations, parallelism) ||
		memory < 8*uint32(parallelism) || memory > 256*1024 || iterations == 0 || iterations > 10 || parallelism == 0 || parallelism > 16 {
		return fmt.Errorf("invalid stored password parameters")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 16 || len(salt) > 64 {
		return fmt.Errorf("invalid stored password salt")
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(want) < 16 || len(want) > 64 {
		return fmt.Errorf("invalid stored password hash")
	}
	actual := argon2.IDKey([]byte(pass), salt, iterations, memory, parallelism, uint32(len(want)))
	if subtle.ConstantTimeCompare(actual, want) != 1 {
		return ErrInvalidCredentials
	}
	return nil
}
