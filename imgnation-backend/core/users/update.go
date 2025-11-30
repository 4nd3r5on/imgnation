package users

import (
	"context"
	"slices"

	"imgnation-backend/pkg/types"
	usersRepo "imgnation-backend/repository/users"

	"github.com/google/uuid"
)

// ChangeUserRoles modifies user roles (admin/moderator only)
func ChangeUserRoles(ctx context.Context, userID uuid.UUID, addRoles, removeRoles []string, repo usersRepo.UsersRepo, currentUserClaims *types.AccessClaims) {
	isAdmin := slices.Contains(currentUserClaims.Roles, RoleAdmin)
	if !isAdmin {
		// TODO: Return unauthorized error if insufficient permissions
	}
	if userID == currentUserClaims.UserID {
		// TODO: Cannot modify own roles - check if userID matches current user's ID
	}

	repo.GetUser(ctx, currentUserClaims.UserID)

	// TODO: Validate that user exists before attempting role change
	// TODO: Return user not found error if user doesn't exist

	err := repo.UpdateUser(ctx, userID, usersRepo.UpdateUser{
		AddRoles:    addRoles,
		RemoveRoles: removeRoles,
	})
	if err != nil {
		// TODO: Handle repository error from UpdateUser
	}
}
