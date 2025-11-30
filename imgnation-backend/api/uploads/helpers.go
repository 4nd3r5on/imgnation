package uploadsApi

import (
	"context"
	"errors"
	"net/http"

	"imgnation-backend/config"
	"imgnation-backend/core/auth"
	"imgnation-backend/core/uploads"
	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"
	"imgnation-backend/repository"
	"imgnation-backend/repository/storage"

	"github.com/google/uuid"
)

func StorageServeContent(
	ctx context.Context, s storage.IStorage,
	w http.ResponseWriter, r *http.Request,
	objKey string,
) {
	obj, err := s.Get(ctx, objKey)
	if err != nil {
		xerr.HttpHandleError(ctx, w, r, err)
	}
	w.Header().Set("Content-Type", obj.ContentType)
	http.ServeContent(w, r, obj.Key, obj.ModTime, obj.Content)
}

func getFileAccess(
	ctx context.Context,
	cfg *config.ServerCfg,
	repo *repository.Repo,
	cache uploads.AccessCache,
	w http.ResponseWriter,
	r *http.Request,
	uploadID uuid.UUID,
) (access *uploads.UploadAccess, payload *types.AccessClaims, ok bool) {
	payload, err := auth.AuthAccess(r, cfg.Auth.JwtAccessSecret)
	if err != nil && !errors.Is(err, xerr.ErrUnauthorized) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, nil, false
	}
	permissions, err := uploads.CheckUploadAccess(
		ctx, repo, cache,
		uploadID, payload,
	)
	if err != nil {
		xerr.HttpHandleError(ctx, w, r, err)
		return nil, nil, false
	}
	return &permissions, payload, true
}
