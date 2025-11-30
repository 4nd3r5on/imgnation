package authApi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"imgnation-backend/config"
	"imgnation-backend/core/auth"
	"imgnation-backend/pkg/xerr"
	"imgnation-backend/repository"

	usersRepo "imgnation-backend/repository/users"
)

type LoginRespBody struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	UserID       string `json:"user_id"`
}

type CheckUsernameTakenReqBody struct {
	Username string `json:"username"`
}
type CheckUsernameTakenRespBody struct {
	IsTaken bool `json:"is_taken"`
}

type RefreshAccessTokenRespBody struct {
	AccessToken string `json:"access_token"`
}

func LoginHandler(cfg *config.ServerCfg, repo *repository.Repo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		switch r.Method {
		case http.MethodPost:
			var reqData auth.LoginReqData
			err := json.NewDecoder(r.Body).Decode(&reqData)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			tokens, userID, err := auth.Login(
				ctx, cfg, repo.UsersRepo, &reqData,
			)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			resp, err := json.Marshal(LoginRespBody{
				RefreshToken: tokens.RefreshToken,
				AccessToken:  tokens.AccesToken,
				UserID:       userID.String(),
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

func SignupHandler(cfg *config.ServerCfg, repo *repository.Repo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		switch r.Method {
		case http.MethodPost:
			payload, err := auth.AuthAccess(r, cfg.Auth.JwtAccessSecret)
			if err != nil && !errors.Is(err, xerr.ErrUnauthorized) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var reqData usersRepo.NewUserReqData
			err = json.NewDecoder(r.Body).Decode(&reqData)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}

			tokens, userID, err := auth.Signup(
				ctx, cfg, repo.UsersRepo, &reqData, payload,
			)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			resp, err := json.Marshal(LoginRespBody{
				RefreshToken: tokens.RefreshToken,
				AccessToken:  tokens.AccesToken,
				UserID:       userID.String(),
			})
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(resp)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}

func RefreshHandler(cfg *config.ServerCfg, repo *repository.Repo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		switch r.Method {
		case http.MethodGet:
			token, err := auth.RefreshAccessToken(
				ctx, repo.UsersRepo,
				cfg.Auth.JwtRefreshSecret,
				cfg.Auth.JwtAccessSecret,
				r,
			)
			if err != nil {
				xerr.HttpHandleError(ctx, w, r, err)
				return
			}
			resp, err := json.Marshal(LoginRespBody{
				AccessToken: token,
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

func CheckUsernameHandler(cfg *config.ServerCfg, repo *repository.Repo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		var req CheckUsernameTakenReqBody
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			xerr.HttpHandleError(ctx, w, r, err)
			return
		}

		isTaken, err := repo.UsersRepo.CheckUsernameTaken(ctx, req.Username)
		if err != nil {
			xerr.HttpHandleError(ctx, w, r, err)
			return
		}
		resp, err := json.Marshal(CheckUsernameTakenRespBody{
			IsTaken: isTaken,
		})
		if err != nil {
			xerr.HttpHandleError(ctx, w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	})
}
