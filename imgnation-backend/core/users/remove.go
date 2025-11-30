package users

import (
	"context"

	"imgnation-backend/pkg/types"
	usersRepo "imgnation-backend/repository/users"

	"github.com/google/uuid"
)

// RemoveUser removes a user (admin only)
func RemoveUser(ctx context.Context, userID uuid.UUID, repo usersRepo.UsersRepo, currentUserClaims *types.AccessClaims) {
	// TODO: Check permissions - only admins should be able to remove users
	// TODO: Return unauthorized error if not admin

	// TODO: Validate user ID format
	// TODO: Return invalid user ID error if empty or invalid format

	// TODO: Check if user exists before attempting removal
	// TODO: Return user not found error if user doesn't exist

	// TODO: Cannot remove self - check if userID matches current user's ID
	// TODO: Return error "cannot remove yourself" if trying to remove self

	// Remove user
	err := repo.RemoveUser(ctx, userID)
	if err != nil {
		// TODO: Handle repository error from RemoveUser
	}
}
