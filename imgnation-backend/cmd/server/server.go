package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"imgnation-backend/api"
	"imgnation-backend/config"
	"imgnation-backend/core/auth"
	"imgnation-backend/core/uploads"
	"imgnation-backend/pkg/ds"
	"imgnation-backend/repository"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/safeblock-dev/werr"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Server struct {
	HttpServer  *http.Server
	HttpApiMux  *http.ServeMux
	WorkerPool  *ds.WorkerPool
	AccessCache uploads.AccessCache

	Cfg     *config.ServerCfg
	MongoDB *mongo.Client
	Minio   *minio.Client
	Repo    *repository.Repo
}

func NewServer(cfg *config.ServerCfg) *Server {
	var wpMaxWorkers int32 = int32(max(runtime.NumCPU()-1, 1))
	return &Server{
		Cfg:         cfg,
		WorkerPool:  ds.NewWorkerPool(wpMaxWorkers, 124),
		AccessCache: uploads.NewAccessCache(),
	}
}

func InitMongoDB(ctx context.Context, mongoCfg *config.MongoCfg) (client *mongo.Client, err error) {
	const maxRetries = 3
	clientOpts := options.Client().ApplyURI(mongoCfg.URI()).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second)
	client, err = mongo.Connect(clientOpts)
	if err != nil {
		return nil, werr.Wrapf(err, "failed to connect")
	}
	err = client.Ping(ctx, nil)
	for i := 0; err != nil && i < maxRetries; i++ {
		log.Printf("Failed to ping MongoDB")
		time.Sleep(500 * time.Millisecond)
		err = client.Ping(ctx, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("mongo ping failed: %w", err)
	}
	return client, nil
}

func InitMinio(cfg *config.MinioCfg) (*minio.Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, werr.Wrapf(err, "failed to create minio client")
	}
	return client, nil
}

func (s *Server) InitHttpServer(ctx context.Context) (err error) {
	httpServerAddr := fmt.Sprintf("%s:%d", s.Cfg.Api.Host, s.Cfg.Api.Port)
	s.HttpServer = &http.Server{
		Addr:    httpServerAddr,
		Handler: api.NewApiHandler(s.Cfg, s.WorkerPool, s.AccessCache, s.Repo),
	}
	return nil
}

func (s *Server) Init(ctx context.Context) (err error) {
	log.Println("Initializing MongoDB")
	if s.MongoDB, err = InitMongoDB(ctx, s.Cfg.Mongo); err != nil {
		return err
	}
	log.Println("Initializing MinIO")
	if s.Minio, err = InitMinio(s.Cfg.Minio); err != nil {
		return err
	}
	log.Println("Initializing Repository layer")
	if s.Repo, err = repository.NewRepo(ctx, s.MongoDB.Database(s.Cfg.Mongo.DB), s.Minio); err != nil {
		return err
	}
	log.Println("Initializing HTTP API")
	if err = s.InitHttpServer(ctx); err != nil {
		return err
	}
	usersCount, err := s.Repo.UsersRepo.GetUsersCount(ctx)
	if err != nil {
		return err
	}
	if usersCount == 0 {
		log.Println("Creating init admin")
		err := auth.InitAdmin(ctx, s.Cfg, s.Repo.UsersRepo, s.Cfg.InitAdmin)
		if err != nil {
			return err
		}
	}
	return nil
}

// Helper func to shutdown the server
func shutdownServ(server *http.Server) {
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v\n", err)
	}
	log.Println("Graceful shutdown complete.")
}

func (s *Server) stop() {
	shutdownServ(s.HttpServer)
	s.WorkerPool.Stop()
}

func (s *Server) Run(ctx context.Context) error {
	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		ctx := context.Background()
		s.WorkerPool.Run(ctx)
	}()

	go func() {
		ctx := context.Background()
		var err error
		for err == nil {
			err = s.AccessCache.WaitRemoveExpired(ctx)
		}
		errChan <- err
	}()

	go func() {
		log.Printf("Running server %s", s.HttpServer.Addr)
		errChan <- s.HttpServer.ListenAndServe()
	}()

	go func() {
		for {
			select {
			case <-time.Tick(3 * time.Second):
				err := uploads.ApplyReactions(context.Background(), s.Repo)
				if err != nil {
					errChan <- err
					return
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			s.stop()
			return ctx.Err()
		case err := <-errChan:
			s.stop()
			return err
		case <-sigChan:
			s.stop()
			return nil
		}
	}
}
