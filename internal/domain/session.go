package domain

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Tokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type Claims struct {
	UserID int64            `json:"userID"`
	Roles  map[int64]string `json:"roles"`
	Type   string           `json:"type"`
	jwt.StandardClaims
}

type RefreshSession struct {
	ID           int64     `db:"id" json:"id"`
	UserID       int64     `db:"user_id" json:"userID"`
	RefreshToken string    `db:"refresh_token" json:"refreshToken"`
	UA           string    `db:"ua" json:"ua"`
	Fingerprint  string    `db:"fingerprint" json:"fingerprint"`
	ExpiresAt    time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
}
