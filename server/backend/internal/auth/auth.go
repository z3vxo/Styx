package auth

import (
	"crypto/rand"
	"encoding/base64"
	"styx/internal/cfg"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Auth struct {
	AccessSec  []byte
	AccessTTL  time.Duration
	RefreshSec []byte
	RefreshTTL time.Duration
}

type Claims struct {
	Usename string `json:"username"`
	jwt.RegisteredClaims
}

func New(cfg *cfg.Cfg) *Auth {
	return &Auth{AccessSec: cfg.Access, RefreshSec: cfg.Refresh, AccessTTL: cfg.AccessTTL, RefreshTTL: cfg.RefreshTTL}
}

type Tokens struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

func (a *Auth) NewJwtPair(user string) (Tokens, error) {
	access, err := a.sign(user, a.AccessSec, a.AccessTTL)
	if err != nil {
		return Tokens{}, err
	}

	refresh, err := a.sign(user, a.RefreshSec, a.RefreshTTL)
	if err != nil {
		return Tokens{}, err
	}

	return Tokens{Access: access, Refresh: refresh}, nil
}

func genID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(b), nil
}
func (a *Auth) sign(user string, sec []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	id, err := genID()
	if err != nil {
		return "", err
	}

	claims := Claims{
		Usename: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id,
			Subject:   user,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(sec)
}
