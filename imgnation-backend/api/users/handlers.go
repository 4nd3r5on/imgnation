package usersApi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"imgnation-backend/config"
	"imgnation-backend/core/auth"
	"imgnation-backend/core/users"
	"imgnation-backend/pkg/xerr"
	"imgnation-backend/repository"

	"github.com/google/uuid"
)

type GetUsersCountRespBody struct {
	Count int64 `json:"count"`
}

type PostUserRespBody struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	UserID       string `json:"user_id"`
}

type UserPublicData struct {
	ID               uuid.UUID `bson:"id" json:"id"`
	ProfilePicfileID uuid.UUID `bson:"profile_pic_file_id" json:"profile_pic_file_id"`
	Username         string    `bson:"username" json:"username"`
	Name             string    `bson:"name" json:"name"`
	Roles            []string  `bson:"roles" json:"roles"`

	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

func GetUsersCountHandler(cfg *config.ServerCfg, repo *repository.Repo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		count, err := users.GetUsersCount(ctx, repo.UsersRepo)
		if err != nil {
			xerr.HttpHandleError(ctx, w, r, err)
			return
		}
		resp, err := json.Marshal(GetUsersCountRespBody{
			Count: count,
		})
		if err != nil {
			xerr.HttpHandleError(ctx, w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	})
}

func UsersHandler(cfg *config.ServerCfg, repo *repository.Repo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = context.Background()
		_, err := auth.AuthAccess(r, cfg.Auth.JwtAccessSecret)
		if err != nil && !errors.Is(err, xerr.ErrUnauthorized) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Method {
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func UsersIdHandler(cfg *config.ServerCfg, repo *repository.Repo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		userIdStr := r.PathValue("id")
		userID, err := uuid.Parse(userIdStr)
		if err != nil {
			http.Error(w, "failed to parse upload ID. Upload ID should be in UUID format", http.StatusBadRequest)
		}
		payload, err := auth.AuthAccess(r, cfg.Auth.JwtAccessSecret)
		if err != nil && !errors.Is(err, xerr.ErrUnauthorized) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			user, err := users.GetUser(ctx, userID, repo.UsersRepo, payload)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			resp, err := json.Marshal(UserPublicData{
				ID:               user.ID,
				ProfilePicfileID: user.ProfilePicfileID,
				Username:         user.Username,
				Name:             user.Name,
				Roles:            user.Roles,
				UpdatedAt:        user.UpdatedAt,
				CreatedAt:        user.CreatedAt,
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

func UsersUsernameHandler(cfg *config.ServerCfg, repo *repository.Repo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		username := r.PathValue("username")
		payload, err := auth.AuthAccess(r, cfg.Auth.JwtAccessSecret)
		if err != nil && !errors.Is(err, xerr.ErrUnauthorized) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			user, err := users.GetUserByUsername(ctx, username, repo.UsersRepo, payload)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			resp, err := json.Marshal(UserPublicData{
				ID:               user.ID,
				ProfilePicfileID: user.ProfilePicfileID,
				Username:         user.Username,
				Name:             user.Name,
				Roles:            user.Roles,
				UpdatedAt:        user.UpdatedAt,
				CreatedAt:        user.CreatedAt,
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
