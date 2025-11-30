package uploadsApi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"imgnation-backend/config"
	"imgnation-backend/core/auth"
	"imgnation-backend/core/uploads"
	"imgnation-backend/pkg/xerr"
	"imgnation-backend/repository"

	"github.com/google/uuid"
)

const (
	ReactionLike    = "like"
	ReactionDislike = "dislike"
)

var AllowedReaction = map[string]struct{}{
	ReactionLike:    {},
	ReactionDislike: {},
}

type ReactHandlerReqBody struct {
	Reaction string `json:"reaction"`
}
type ReactHandlerRespBody struct {
	IsSet bool `json:"is_set"`
}

func ReactUploadHandler(cfg *config.ServerCfg, repo *repository.Repo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		switch r.Method {
		case http.MethodPost:
			payload, err := auth.AuthAccess(r, cfg.Auth.JwtAccessSecret)
			if err != nil && !errors.Is(err, xerr.ErrUnauthorized) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			uploadIdStr := r.PathValue("id")
			uploadID, err := uuid.Parse(uploadIdStr)
			if err != nil {
				http.Error(w, "failed to parse upload ID. Upload ID should be in UUID format", http.StatusBadRequest)
			}

			var req ReactHandlerReqBody
			err = json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			_, allowedReaction := AllowedReaction[req.Reaction]
			if !allowedReaction {
				errMsg := fmt.Sprintf("reaction %s is not supported", req.Reaction)
				http.Error(w, errMsg, http.StatusBadRequest)
				return
			}
			isSet, err := uploads.ToggleReaction(ctx, repo, uploadID, payload, req.Reaction)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			resp, err := json.Marshal(ReactHandlerRespBody{IsSet: isSet})
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
