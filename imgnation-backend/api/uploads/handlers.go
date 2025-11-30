package uploadsApi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"imgnation-backend/config"
	"imgnation-backend/core/auth"
	"imgnation-backend/core/common"
	"imgnation-backend/core/process"
	"imgnation-backend/core/upload"
	"imgnation-backend/core/uploads"
	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"
	"imgnation-backend/repository"

	"github.com/google/uuid"
)

type GetUploadsDataIdRespBody struct {
	Tags          []string                `json:"tags"`
	FileSizeBytes int64                   `json:"file_size_bytes"`
	Variants      []types.FileVariantInfo `json:"variants"`
}

type PostUploadsRespBody struct {
	Results []upload.UploadResult `json:"results"`
}

func UploadsDataIdHandler(
	cfg *config.ServerCfg,
	repo *repository.Repo,
	cache uploads.AccessCache,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		uploadIdStr := r.PathValue("id")
		uploadID, err := uuid.Parse(uploadIdStr)
		if err != nil {
			http.Error(w, "failed to parse upload ID. Upload ID should be in UUID format", http.StatusBadRequest)
		}

		switch r.Method {
		case http.MethodGet:
			access, payload, ok := getFileAccess(ctx, cfg, repo, cache, w, r, uploadID)
			if !ok {
				return
			}
			if !access.Read {
				http.Error(w, "file access is not allowed", http.StatusUnauthorized)
				return
			}

			fileMetadata, err := uploads.GetFileMetadata(ctx, repo, cfg, payload, uploadID)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}

			resp, err := json.Marshal(fileMetadata)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(resp)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func UploadsIdHandler(
	cfg *config.ServerCfg,
	repo *repository.Repo,
	cache uploads.AccessCache,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		idStr := r.PathValue("id")
		uploadID, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "failed to parse image ID. Image ID should be in UUID format", http.StatusBadRequest)
		}

		switch r.Method {
		case http.MethodGet:
			access, _, ok := getFileAccess(ctx, cfg, repo, cache, w, r, uploadID)
			if !ok {
				return
			}
			if !access.Read {
				http.Error(w, "file access is not allowed", http.StatusUnauthorized)
				return
			}

			variant := r.URL.Query().Get("variant")
			objKey := process.MakeKeyFileVariant(uploadID, variant)
			obj, err := repo.UploadsStorage.Get(ctx, objKey)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			w.Header().Set("Content-Type", obj.ContentType)
			http.ServeContent(w, r, obj.Key, obj.ModTime, obj.Content)

		case http.MethodDelete:
			access, _, ok := getFileAccess(ctx, cfg, repo, cache, w, r, uploadID)
			if !ok {
				return
			}
			if !access.Delete {
				http.Error(w, "file access is not allowed", http.StatusUnauthorized)
				return
			}

			// TODO: Implement
			w.WriteHeader(http.StatusNotImplemented)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func UploadsHandler(cfg *config.ServerCfg, repo *repository.Repo, deps *upload.UploadDeps[*upload.UploadMetadata]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		payload, err := auth.AuthAccess(r, cfg.Auth.JwtAccessSecret)
		if err != nil && !errors.Is(err, xerr.ErrUnauthorized) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			q := r.URL.Query()
			searchParams, err := uploads.ParseSearchUploadsParams(q)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			paginationParams, err := common.ParsePaginationParams(q)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			res, err := uploads.SearchUploads(ctx, repo, payload, searchParams, paginationParams)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			resp, err := json.Marshal(res)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(resp)

		case http.MethodPost:
			mr, err := r.MultipartReader()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			results, err := upload.HandleUpload(ctx, deps, repo, cfg, payload, mr)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			resp, err := json.Marshal(PostUploadsRespBody{
				Results: results,
			})
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(resp)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
