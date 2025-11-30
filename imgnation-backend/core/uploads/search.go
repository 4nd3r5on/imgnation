package uploads

import (
	"context"
	"slices"

	"imgnation-backend/pkg/types"
	"imgnation-backend/repository"

	"github.com/google/uuid"
)

func SearchUploads(
	ctx context.Context,
	repo *repository.Repo,
	authData *types.AccessClaims,
	searchParams *types.SearchUploadsParams,
	paginationParams *types.PaginationParams,
) (*types.UploadsSearchResults, error) {
	userID := uuid.Nil
	isAdmin := false
	if authData != nil {
		userID = authData.UserID
		// TODO: Don't hardcode admin role
		isAdmin = slices.Contains(authData.Roles, "admin")
	}
	return repo.UploadsReg.SearchUploads(
		ctx, searchParams, paginationParams, userID, !isAdmin,
	)
}
