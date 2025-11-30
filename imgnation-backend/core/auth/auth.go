package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"

	usersRepo "imgnation-backend/repository/users"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
)

const AccessTokenTTL = 30 * time.Minute

type Tokens struct {
	AccesToken, RefreshToken string
}

func ParseAuthHeader(r *http.Request) (string, error) {
	const bearerPrefix = "Bearer "
	header := strings.TrimSpace(r.Header.Get("Authorization"))

	if header == "" {
		return "", werr.Wrapf(xerr.ErrUnauthorized, "missing authorization header")
	}

	if len(header) >= len(bearerPrefix) &&
		strings.EqualFold(header[:len(bearerPrefix)], bearerPrefix) {
		token := strings.TrimSpace(header[len(bearerPrefix):])
		if token == "" {
			return "", werr.Wrapf(xerr.ErrUnauthorized, "empty bearer token")
		}
		return token, nil
	}

	return header, nil
}

// Generic auth function
func Auth(r *http.Request, secret []byte, claims jwt.Claims) (*jwt.Token, error) {
	tokenString, err := ParseAuthHeader(r)
	if err != nil {
		return nil, werr.Wrapf(err, "failed to parse authorization header")
	}
	// Create a pointer to zero value for ParseWithClaims
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, werr.Wrapf(xerr.ErrUnauthorized, "unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, werr.Wrapf(xerr.ErrUnauthorized, "invalid token: %v", err)
	}
	return token, nil
}

func AuthAccess(r *http.Request, secret []byte) (*types.AccessClaims, error) {
	claims := &types.AccessClaims{}
	token, err := Auth(r, secret, claims)
	if err != nil {
		return nil, werr.Wrapf(err, "failed to auth access token")
	}
	if !token.Valid {
		return nil, werr.Wrapf(err, "invalid token")
	}
	return claims, nil
}

func AuthRefresh(r *http.Request, secret []byte) (*types.RefreshClaims, error) {
	claims := &types.RefreshClaims{}
	token, err := Auth(r, secret, claims)
	if err != nil {
		return nil, werr.Wrapf(err, "failed to auth access token")
	}
	if !token.Valid {
		return nil, werr.Wrapf(err, "invalid token")
	}
	return claims, nil
}

func RefreshAccessToken(
	ctx context.Context,
	users usersRepo.UsersRepo,
	refreshSecret []byte,
	accessSecret []byte,
	r *http.Request,
) (string, error) {
	authData, err := AuthRefresh(r, refreshSecret)
	if err != nil {
		return "", werr.Wrapf(err, "failed to handle refresh token")
	}
	userData, err := users.GetUser(ctx, authData.UserID)
	if err != nil {
		if errors.Is(err, xerr.ErrEntityNotFound) {
			return "", werr.Wrapf(xerr.ErrUnauthorized, "user doesn't exist")
		}
		return "", werr.Wrapf(err, "failed to get user info")
	}
	accessToken, err := IssueAccessToken(
		accessSecret,
		authData.UserID,
		userData.Roles,
		userData.CreatedAt,
		AccessTokenTTL,
	)
	return accessToken, err
}

// IssueAccessToken generates a new JWT token with HMAC signing
// Parameters:
// - secret: HMAC signing secret
// - expiresIn: Token validity duration from issuance time
// Returns signed JWT token string
func IssueAccessToken(
	secret []byte,
	userID uuid.UUID,
	roles []string,
	userCreatedAt time.Time,
	expiresIn time.Duration,
) (string, error) {
	claims := types.AccessClaims{
		UserID:        userID,
		Roles:         roles,
		UserCreatedAt: userCreatedAt,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", werr.Wrapf(err, "failed to sign token")
	}
	return signedToken, nil
}

func IssueRefreshToken(
	secret []byte,
	userID uuid.UUID,
) (string, error) {
	claims := types.RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", werr.Wrapf(err, "failed to sign token")
	}
	return signedToken, nil
}
