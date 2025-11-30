package uploadsApi

import (
	"context"
	"net/http"
	"strconv"

	"imgnation-backend/config"
	"imgnation-backend/core/process"
	"imgnation-backend/core/uploads"
	"imgnation-backend/pkg/types"
	"imgnation-backend/repository"

	"github.com/google/uuid"
)

func GetManifest(cfg *config.ServerCfg, repo *repository.Repo, cache uploads.AccessCache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		variant := r.PathValue("variant")
		uploadIdStr := r.PathValue("id")
		uploadID, err := uuid.Parse(uploadIdStr)
		if err != nil {
			http.Error(w, "failed to parse upload ID. Upload ID should be in UUID format", http.StatusBadRequest)
		}

		access, _, ok := getFileAccess(ctx, cfg, repo, cache, w, r, uploadID)
		if !ok {
			return
		}
		if !access.Read {
			http.Error(w, "file access is not allowed", http.StatusUnauthorized)
			return
		}
		objKey := process.MakeKeyFileVariant(uploadID, variant)
		StorageServeContent(ctx, repo.UploadsStorage, w, r, objKey)
	})
}

func GetInitChunk(cfg *config.ServerCfg, repo *repository.Repo, cache uploads.AccessCache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		variant := r.PathValue("variant")
		uploadIdStr := r.PathValue("id")
		uploadID, err := uuid.Parse(uploadIdStr)
		if err != nil {
			http.Error(w, "failed to parse upload ID. Upload ID should be in UUID format", http.StatusBadRequest)
		}

		access, _, ok := getFileAccess(ctx, cfg, repo, cache, w, r, uploadID)
		if !ok {
			return
		}
		if !access.Read {
			http.Error(w, "file access is not allowed", http.StatusUnauthorized)
			return
		}
		objKey := process.MakeKeyChunk(uploadID, variant, types.ChunkTypeInit, 0)
		StorageServeContent(ctx, repo.UploadsStorage, w, r, objKey)
	})
}

func GetSegmentChunk(cfg *config.ServerCfg, repo *repository.Repo, cache uploads.AccessCache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		variant := r.PathValue("variant")
		chunkIdxStr := r.PathValue("chunk_idx")
		uploadIdStr := r.PathValue("id")
		uploadID, err := uuid.Parse(uploadIdStr)
		if err != nil {
			http.Error(w, "failed to parse upload ID. Upload ID should be in UUID format", http.StatusBadRequest)
		}
		chunkIdx, err := strconv.ParseInt(chunkIdxStr, 10, 64)
		if err != nil {
			http.Error(w, "failed to parse Chunk Idx. Chunk Idx should be an integer", http.StatusBadRequest)
		}

		access, _, ok := getFileAccess(ctx, cfg, repo, cache, w, r, uploadID)
		if !ok {
			return
		}
		if !access.Read {
			http.Error(w, "file access is not allowed", http.StatusUnauthorized)
			return
		}
		objKey := process.MakeKeyChunk(uploadID, variant, types.ChunkTypeSegment, chunkIdx)
		StorageServeContent(ctx, repo.UploadsStorage, w, r, objKey)
	})
}
