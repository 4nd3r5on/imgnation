package upload

import (
	"context"
	"errors"
	"regexp"
	"slices"
	"strings"

	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"
	usersRepo "imgnation-backend/repository/users"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
)

func ProcessAccessMetadata(
	ctx context.Context,
	users usersRepo.UsersRepo,
	authData *types.AccessClaims,
	access AccessControl,
) (types.AccessControl, error) {
	levelNum, ok := AccessLevelMap[access.Level]
	if !ok {
		return types.AccessControl{}, werr.Wrapf(xerr.ErrInvalidArgument, "access level %s is not supported", access.Level)
	}
	if authData == nil {
		if levelNum > types.AccessLevelAnyone {
			return types.AccessControl{}, errors.New("access level for unauthorized users should be always 'anyone'")
		}
		return types.AccessControl{
			Level:        levelNum,
			AllowedUsers: []uuid.UUID{},
		}, nil
	}

	// TODO: Stop hardcoding admin role
	isAdmin := slices.Contains(authData.Roles, "admin")
	if !isAdmin && levelNum == types.AccessLevelAdmin {
		return types.AccessControl{}, werr.Wrapf(xerr.ErrUnauthorized,
			"access level admin is not allowed for non-admin users")
	}

	allowedIDs := make([]uuid.UUID, 0, len(access.AllowedUserEmails))
	emailIdPairs, err := users.GetIDsByEmails(ctx, access.AllowedUserEmails)
	if err != nil {
		return types.AccessControl{}, werr.Wrapf(err, "failed to get E-Mails by UserIDs")
	}
	for _, emailIdPair := range emailIdPairs {
		allowedIDs = append(allowedIDs, emailIdPair.ID)
	}
	return types.AccessControl{
		Level:        levelNum,
		AllowedUsers: allowedIDs,
	}, nil
}

func ValidateDescription(s string) error {
	if len(s) > 300 {
		return werr.Wrapf(xerr.ErrInvalidArgument, "description argument is too long")
	}
	return nil
}

func ProcessTags(tags []string) (processed []string) {
	tagsMap := map[string]struct{}{}

	tagRegexp, _ := regexp.Compile("^[a-z0-9][a-z0-9_]*$")

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if len(tag) > 32 || len(tag) == 0 || !tagRegexp.MatchString(tag) {
			continue
		}
		tagsMap[tag] = struct{}{}
	}

	processed = make([]string, 0, len(tagsMap))
	for tag := range tagsMap {
		processed = append(processed, tag)
	}
	return processed
}
