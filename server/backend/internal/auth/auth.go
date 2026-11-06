package authmgr

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
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

func (a *Auth) Verify(token string) (*Claims, error) {
	return a.verify(token, a.AccessSec)
}
func (a *Auth) VerifyRefresh(token string) (*Claims, error) {
	return a.verify(token, a.RefreshSec)
}

func (a *Auth) verify(token string, secret []byte) (*Claims, error) {
	claims := &Claims{}

	tokenStr, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	if !tokenStr.Valid {
		return nil, err
	}
	return claims, err
}

type contextKey string

const claimsKey contextKey = "claims"

func (a *Auth) AuthWare() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			scheme, tokenStr, ok := strings.Cut(header, " ")
			if !ok || !strings.EqualFold(scheme, "Bearer") || tokenStr == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "missing Authorization header or token"})
				return
			}
			claims, err := a.Verify(tokenStr)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid token"})
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
