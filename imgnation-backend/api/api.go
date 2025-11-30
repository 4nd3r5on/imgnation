package api

import (
	"fmt"
	"net/http"
	"time"

	"imgnation-backend/api/middleware"
	"imgnation-backend/config"
	"imgnation-backend/core/upload"
	"imgnation-backend/core/uploads"
	"imgnation-backend/pkg/ds"
	"imgnation-backend/repository"

	authApi "imgnation-backend/api/auth"

	uploadsApi "imgnation-backend/api/uploads"
	usersApi "imgnation-backend/api/users"
)

func NewApiHandler(cfg *config.ServerCfg, wp ds.IWorkerPool, aceessCache uploads.AccessCache, repo *repository.Repo) *http.ServeMux {
	prefix := cfg.Api.Prefix
	r := http.NewServeMux()
	cors := middleware.CorsMiddleware{Cfg: cfg.Api.Cors}
	fileProcessors := upload.InitProcessors()
	uploadDeps := &upload.UploadDeps[*upload.UploadMetadata]{
		RepoHelpers: upload.NewRegistryHelpers(repo),
		Storage:     repo.UploadsStorage,
		WorkerPool:  wp,
		Processors:  fileProcessors,
		Opts: &upload.UploadOpts{ // give 15 seconds to respond
			CreateUploadRecordAfter: 15 * time.Second,
		},
	}

	// Authenticantion ================================

	r.Handle(fmt.Sprintf("%s/auth/login", prefix),
		cors.Middleware(authApi.LoginHandler(cfg, repo)))
	r.Handle(fmt.Sprintf("%s/auth/signup", prefix),
		cors.Middleware(authApi.SignupHandler(cfg, repo)))
	r.Handle(fmt.Sprintf("%s/auth/refresh", prefix),
		cors.Middleware(authApi.RefreshHandler(cfg, repo)))
	r.Handle(fmt.Sprintf("%s/auth/checkusername", prefix),
		cors.Middleware(authApi.CheckUsernameHandler(cfg, repo)))

	// Users ==========================================

	r.Handle(prefix+"%s/users/count",
		cors.Middleware(usersApi.GetUsersCountHandler(cfg, repo)))
	r.Handle(prefix+"/users",
		cors.Middleware(usersApi.UsersHandler(cfg, repo)))
	r.Handle(prefix+"/users/{id}",
		cors.Middleware(usersApi.UsersIdHandler(cfg, repo)))
	r.Handle(prefix+"/users/username/{username}",
		cors.Middleware(usersApi.UsersUsernameHandler(cfg, repo)))

	// Uploads  =======================================

	// Endpoint for uploading and searching uploads
	// (pagination+search params should be used for get request)
	r.Handle(prefix+"/uploads", // POST/GET
		cors.Middleware(uploadsApi.UploadsHandler(cfg, repo, uploadDeps)))
	// Endpoint for deleteing uploads and getting upload files
	// For GET request variant is specified through url query params (default is "" for original)
	r.Handle(prefix+"/uploads/{id}", // DELETE/GET
		cors.Middleware(uploadsApi.UploadsIdHandler(cfg, repo, aceessCache)))
	r.Handle(prefix+"/uploads/data/{id}", // GET
		cors.Middleware(uploadsApi.UploadsDataIdHandler(cfg, repo, aceessCache)))
	// Endpoint for reacting on uploads (like/dislike)
	r.Handle(fmt.Sprintf("%s/uploads/react/{id}", prefix),
		cors.Middleware(uploadsApi.ReactUploadHandler(cfg, repo)))

	// Streaming ======================================

	r.Handle(fmt.Sprintf("GET %s/uploads/{id}/{variant}/manifest", prefix),
		cors.Middleware(uploadsApi.GetManifest(cfg, repo, aceessCache)))
	r.Handle(fmt.Sprintf("GET %s/uploads/{id}/{variant}/chunks/segment/{chunk_idx}", prefix),
		cors.Middleware(uploadsApi.GetSegmentChunk(cfg, repo, aceessCache)))
	r.Handle(fmt.Sprintf("GET %s/uploads/{id}/{variant}/chunks/init", prefix),
		cors.Middleware(uploadsApi.GetInitChunk(cfg, repo, aceessCache)))

	return r
}
