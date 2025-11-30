package uploads

import (
	"context"
	"errors"

	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"
	"imgnation-backend/repository"

	reactionsRepo "imgnation-backend/repository/reactions"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
)

func addView(
	ctx context.Context,
	repo *repository.Repo,
	uploadID uuid.UUID,
	authData *types.AccessClaims,
) (err error) {
	const reaction = "view"
	userID := uuid.Nil
	if authData != nil {
		userID = authData.UserID
	}

	if err = repo.ReactionsRepo.AddReaction(ctx, reactionsRepo.ReactionOpts{
		UploadID: uploadID,
		UserID:   userID,
		Reaction: reaction,
	}); err != nil {
		if errors.Is(err, xerr.ErrEntityExists) {
			return nil
		}
		return err
	}
	_, err = repo.UploadsReg.AddReactionCount(ctx, uploadID, reaction, 1)
	return err
}

// middleware to getting access info that tries to add users to file viewers
func getUploadAccessAddView(
	ctx context.Context,
	repo *repository.Repo,
	uploadID uuid.UUID,
	authData *types.AccessClaims,
) (*AccessItem, error) {
	access, err := getUploadAccess(ctx, repo.UploadsReg, uploadID)
	if err != nil {
		return access, err
	}

	if authData != nil {
		err := addView(ctx, repo, uploadID, authData)
		if err != nil {
			if errors.Is(err, xerr.ErrEntityExists) {
				return access, nil
			}
			return access, werr.Wrapf(err, "failed to add upload ID %s to viewed", uploadID.String())
		}
	}
	return access, nil
}

func ToggleReaction(
	ctx context.Context,
	repo *repository.Repo,
	uploadID uuid.UUID,
	authData *types.AccessClaims,
	reaction string,
) (isSet bool, err error) {
	if authData == nil {
		return false, werr.Wrapf(xerr.ErrUnauthorized, "authorization required to set reaction")
	}
	res, err := repo.ReactionsCache.ToggleReaction(ctx, reactionsRepo.ReactionOpts{
		UserID:   authData.UserID,
		UploadID: uploadID,
		Reaction: reaction,
	}, repo.ReactionsRepo)
	if err != nil {
		if errors.Is(err, xerr.ErrEntityNotFound) {
			return false, nil
		}
		return false, werr.Wrapf(err, "failed to remove like on upload ID %s", uploadID.String())
	}
	return res.IsSet, nil
}

type UploadIdReactionPair struct {
	UploadID uuid.UUID
	Reaction string
}

func ApplyReactions(
	ctx context.Context,
	repo *repository.Repo,
) error {
	reactionOpts, err := repo.ReactionsCache.FlushPendingReactions()
	if err != nil {
	}
	toggleResults, err := repo.ReactionsRepo.ToggleReactions(ctx, reactionOpts)
	if err != nil {
		return err
	}
	reactionCounters := map[UploadIdReactionPair]int64{}
	for _, result := range toggleResults {
		key := UploadIdReactionPair{
			UploadID: result.UploadID,
			Reaction: result.Reaction,
		}
		oldVal := reactionCounters[key]
		newVal := oldVal
		if result.IsSet {
			newVal++
		} else {
			newVal--
		}
		reactionCounters[key] = newVal
	}

	for reaction, count := range reactionCounters {
		_, err := repo.UploadsReg.AddReactionCount(ctx, reaction.UploadID, reaction.Reaction, count)
		if err != nil {
			return err
		}
	}
	return nil
}
