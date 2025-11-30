package auth

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"imgnation-backend/config"
	"imgnation-backend/core/common"
	"imgnation-backend/core/users"
	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"

	usersRepo "imgnation-backend/repository/users"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
	"golang.org/x/crypto/bcrypt"
)

// Signup creates a new user account and returns authentication tokens
func Signup(
	ctx context.Context,
	cfg *config.ServerCfg,
	repo usersRepo.UsersRepo,
	data *usersRepo.NewUserReqData,
	currentUserClaims *types.AccessClaims,
) (tokens *Tokens, userID uuid.UUID, err error) {
	if err := validateSignupData(data); err != nil {
		return nil, uuid.Nil, werr.Wrapf(xerr.ErrInvalidArgument, "%s", err.Error())
	}
	data.Email = strings.ToLower(data.Email)

	isAdmin := false
	if currentUserClaims != nil {
		isAdmin = slices.Contains(currentUserClaims.Roles, users.RoleAdmin)
	}

	roles := data.Roles
	if !isAdmin || len(roles) == 0 {
		roles = []string{users.DefaultRole}
	} else {
		// keep only unique roles (exclude duplicates)
		rolesMap := map[string]struct{}{}
		for _, role := range data.Roles {
			rolesMap[role] = struct{}{}
		}
		roles = make([]string, len(rolesMap))
		i := 0
		for role := range rolesMap {
			roles[i] = role
			i++
		}
	}

	taken, err := repo.CheckUsernameTaken(ctx, data.Username)
	if err != nil {
		return nil, uuid.Nil, werr.Wrapf(err, "failed to check username availability")
	}
	if taken {
		return nil, uuid.Nil, werr.Wrapf(xerr.ErrEntityExists, "username already taken")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, uuid.Nil, werr.Wrapf(err, "failed to hash password")
	}

	// Create user data for repository
	newUserID, err := repo.CreateUser(ctx, usersRepo.NewUserRepoData{
		ID:           uuid.New(),
		Username:     data.Username,
		Name:         data.Name,
		Email:        data.Email,
		Roles:        roles,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, uuid.Nil, werr.Wrapf(err, "failed to create user")
	}

	// Generate tokens for the new user
	now := time.Now()
	accessToken, err := IssueAccessToken(cfg.Auth.JwtAccessSecret, newUserID, roles, now, AccessTokenTTL)
	if err != nil {
		return nil, uuid.Nil, werr.Wrapf(err, "failed to issue access token")
	}
	refreshToken, err := IssueRefreshToken(cfg.Auth.JwtRefreshSecret, newUserID)
	if err != nil {
		return nil, uuid.Nil, werr.Wrapf(err, "failed to issue refresh token")
	}

	return &Tokens{
		AccesToken:   accessToken,
		RefreshToken: refreshToken,
	}, newUserID, nil
}

// validateSignupData validates the signup request data
func validateSignupData(data *usersRepo.NewUserReqData) error {
	if data == nil {
		return errors.New("signup data is required")
	}
	username, isValid := ValidateUsername(data.Username)
	if !isValid {
		return werr.Wrapf(xerr.ErrInvalidArgument, "invalid username format")
	}
	data.Username = username

	if strings.TrimSpace(data.Username) == "" {
		return errors.New("username is required")
	}

	if strings.TrimSpace(data.Name) == "" {
		return werr.Wrapf(xerr.ErrInvalidArgument, "name is required")
	}
	if !common.IsValidEmail(data.Email) {
		return werr.Wrapf(xerr.ErrInvalidArgument, "invalid email format")
	}
	if err := common.ValidatePassword(data.Password); err != nil {
		return err
	}
	return nil
}

func InitAdmin(
	ctx context.Context,
	cfg *config.ServerCfg,
	repo usersRepo.UsersRepo,
	data *usersRepo.NewUserReqData,
) (err error) {
	if err := validateSignupData(data); err != nil {
		return werr.Wrapf(xerr.ErrInvalidArgument, "%s", err.Error())
	}

	data.Email = strings.ToLower(data.Email)
	// keep only unique roles (exclude duplicates)
	rolesMap := map[string]struct{}{}
	for _, role := range data.Roles {
		rolesMap[role] = struct{}{}
	}
	roles := make([]string, len(rolesMap))
	i := 0
	for role := range rolesMap {
		roles[i] = role
		i++
	}

	taken, err := repo.CheckUsernameTaken(ctx, data.Username)
	if err != nil {
		return werr.Wrapf(err, "failed to check username availability")
	}
	if taken {
		return werr.Wrapf(xerr.ErrEntityExists, "username already taken")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return werr.Wrapf(err, "failed to hash password")
	}

	// Create user data for repository
	_, err = repo.CreateUser(ctx, usersRepo.NewUserRepoData{
		Username:     data.Username,
		Name:         data.Name,
		Email:        data.Email,
		Roles:        roles,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return werr.Wrapf(err, "failed to create user")
	}
	return nil
}
