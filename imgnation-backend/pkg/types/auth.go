package types

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RefreshClaims struct {
	UserID uuid.UUID `json:"id"`
	jwt.RegisteredClaims
}

type AccessClaims struct {
	UserID        uuid.UUID `json:"id"`
	Roles         []string  `json:"roles"`
	UserCreatedAt time.Time `json:"user_created_at"`
	jwt.RegisteredClaims
}
