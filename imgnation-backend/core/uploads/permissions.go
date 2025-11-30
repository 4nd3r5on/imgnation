package uploads

import (
	"context"
	"fmt"
	"slices"
	"time"

	"imgnation-backend/core/users"
	"imgnation-backend/pkg/ds"
	"imgnation-backend/pkg/types"
	"imgnation-backend/repository"

	"github.com/google/uuid"

	"github.com/safeblock-dev/werr"
)

const AccessItemTTL = 15 * time.Minute

type AccessItem struct {
	UploaderIDs  map[uuid.UUID]struct{}
	AllowedUsers map[uuid.UUID]struct{}
	Level        types.AccessLevel
}

type UploadAccess struct {
	IsUploadCreator bool
	Delete          bool
	Read            bool
}

type AccessCache ds.ICache[uuid.UUID, *AccessItem]

func NewAccessCache() AccessCache {
	return ds.NewCache[uuid.UUID, *AccessItem](AccessItemTTL)
}

func CheckUploadAccess(
	ctx context.Context,
	repo *repository.Repo,
	cache AccessCache,
	uploadID uuid.UUID,
	authData *types.AccessClaims,
) (access UploadAccess, err error) {
	accessItem, err := GetUploadAccess(ctx, repo, cache, uploadID, authData)
	if err != nil {
		return UploadAccess{}, err
	}
	fmt.Printf("%+v\n", accessItem.Level == types.AccessLevelAnyone)
	if authData == nil || authData.UserID == uuid.Nil {
		return UploadAccess{
			Read:            accessItem.Level == types.AccessLevelAnyone,
			Delete:          false,
			IsUploadCreator: false,
		}, nil
	}

	if _, ok := accessItem.UploaderIDs[authData.UserID]; ok {
		return UploadAccess{
			Read:            true,
			Delete:          true,
			IsUploadCreator: true,
		}, nil
	}

	isAdmin := slices.Contains(authData.Roles, users.RoleAdmin)

	if accessItem.Level <= types.AccessLevelAuthorized {
		access.Read = true
		access.Delete = isAdmin
	} else if accessItem.Level == types.AccessLevelPrivate {
		_, access.Read = accessItem.AllowedUsers[authData.UserID]
		access.Read = access.Read || isAdmin
		access.Delete = isAdmin
	} else if accessItem.Level == types.AccessLevelAdmin {
		return UploadAccess{
			Read:            isAdmin,
			Delete:          isAdmin,
			IsUploadCreator: false,
		}, nil
	} else {
		return UploadAccess{
			Read:            isAdmin,
			Delete:          isAdmin,
			IsUploadCreator: false,
		}, nil
	}

	return access, nil
}

// Uses cache + saves view
func GetUploadAccess(
	ctx context.Context,
	repo *repository.Repo,
	cache AccessCache,
	uploadID uuid.UUID,
	authData *types.AccessClaims,
) (access *AccessItem, err error) {
	access, ok := cache.Get(uploadID)
	if ok {
		return access, nil
	}
	access, err = getUploadAccessAddView(ctx, repo, uploadID, authData)
	if err != nil {
		return nil, werr.Wrapf(err, "failed to upload access")
	}
	cache.Set(uploadID, access)
	return access, nil
}

// Internal function to get access info from repo
func getUploadAccess(ctx context.Context, reg types.UploadsReg, uploadID uuid.UUID) (*AccessItem, error) {
	records, err := reg.GetUploadRecords(ctx, uploadID)
	if err != nil {
		return nil, werr.Wrapf(err, "failed to get upload records")
	}
	access := AccessItem{
		UploaderIDs:  make(map[uuid.UUID]struct{}, len(records)),
		AllowedUsers: make(map[uuid.UUID]struct{}, 0),
	}
	for _, record := range records {
		access.Level = min(access.Level, record.Access.Level)
		access.UploaderIDs[record.UserID] = struct{}{}
		for _, user := range record.Access.AllowedUsers {
			access.AllowedUsers[user] = struct{}{}
		}
	}
	return &access, nil
}
