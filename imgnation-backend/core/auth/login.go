package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"imgnation-backend/config"
	"imgnation-backend/core/common"
	"imgnation-backend/pkg/xerr"

	usersRepo "imgnation-backend/repository/users"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
	"golang.org/x/crypto/bcrypt"
)

type LoginReqData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(
	ctx context.Context,
	cfg *config.ServerCfg,
	repo usersRepo.UsersRepo,
	data *LoginReqData,
) (tokens *Tokens, userID uuid.UUID, err error) {
	if !common.IsValidEmail(data.Email) {
		return nil, uuid.Nil, werr.Wrapf(xerr.ErrInvalidArgument, "invalid email")
	}

	user, err := repo.GetUserByEmail(ctx, strings.ToLower(data.Email))
	if err != nil {
		if errors.Is(err, xerr.ErrEntityNotFound) {
			return nil, uuid.Nil, werr.Wrapf(xerr.ErrUnauthorized, "incorrect email or password")
		}
		return nil, uuid.Nil, werr.Wrapf(err, "failed to get user info")
	}
	if user.ID == uuid.Nil {
		return nil, uuid.Nil, errors.New("how the duck UserID in the DB is nil? It suppose to be reserved for anon user")
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(data.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, uuid.Nil, werr.Wrapf(xerr.ErrUnauthorized, "incorrect email or password")
		}
		return nil, uuid.Nil, werr.Wrapf(err, "failed to compare hash with password")
	}

	now := time.Now()
	accessToken, err := IssueAccessToken(cfg.Auth.JwtAccessSecret, user.ID, user.Roles, now, AccessTokenTTL)
	if err != nil {
		return nil, uuid.Nil, werr.Wrapf(err, "failed to issue access token")
	}
	refreshToken, err := IssueRefreshToken(cfg.Auth.JwtRefreshSecret, user.ID)
	if err != nil {
		return nil, uuid.Nil, werr.Wrapf(err, "failed to issue refresh token")
	}

	return &Tokens{
		AccesToken:   accessToken,
		RefreshToken: refreshToken,
	}, user.ID, nil
}
