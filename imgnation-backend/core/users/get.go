package users

import (
	"context"

	"imgnation-backend/pkg/types"

	usersRepo "imgnation-backend/repository/users"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
)

// GetUser retrieves user information
func GetUser(ctx context.Context, userID uuid.UUID, repo usersRepo.UsersRepo, currentUserClaims *types.AccessClaims) (usersRepo.StoredUser, error) {
	user, err := repo.GetUser(ctx, userID)
	if err != nil {
		return usersRepo.StoredUser{}, werr.Wrapf(err, "failed to get user")
	}
	return user, nil
}

func GetUserByUsername(ctx context.Context, username string, repo usersRepo.UsersRepo, currentUserClaims *types.AccessClaims) (usersRepo.StoredUser, error) {
	user, err := repo.GetUserByUsername(ctx, username)
	if err != nil {
		return usersRepo.StoredUser{}, werr.Wrapf(err, "failed to get user")
	}
	return user, nil
}

// GetUsersCount returns total number of users (moderator only)
func GetUsersCount(ctx context.Context, repo usersRepo.UsersRepo) (count int64, err error) {
	count, err = repo.GetUsersCount(ctx)
	if err != nil {
		return 0, werr.Wrapf(err, "failed to get users count")
	}
	return count, nil
}
