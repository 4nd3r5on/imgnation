// package upload provides functionality for uploading files to server
// don't confuse with uploads package (which is overall for managing uploads)
package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"imgnation-backend/core/process"
	"imgnation-backend/pkg/ds"
	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"
	"imgnation-backend/repository/storage"

	uploadUtils "imgnation-backend/pkg/utils/upload"

	"github.com/google/uuid"
	"github.com/safeblock-dev/werr"
)

type FileUploadStatus int

const (
	FileUploadStatusPending FileUploadStatus = iota
	FileUploadStatusSuccessful
	FileUploadStatusFailed
)

type UploadOpts struct {
	// Doesn't matter if all processing finished, publish anyway
	CreateUploadRecordAfter time.Duration
}

type UploadDeps[MetadataT any] struct {
	RepoHelpers types.RepositoryHelpers[MetadataT]
	Storage     storage.IStorage
	WorkerPool  ds.IWorkerPool
	Processors  map[string]process.ProcessFunc
	Opts        *UploadOpts
}

type UploadManager[MetadataT any] struct {
	deps     *UploadDeps[MetadataT]
	fileOpts *types.NewFileOpts[MetadataT]

	mu       sync.RWMutex
	status   FileUploadStatus
	variants []types.FileVariantInfo
	chunks   []types.ChunkInfo
	errors   []error
	progress int // Tracks processed variants
}

type VariantResult struct {
	Variants []types.FileVariantInfo
	Chunks   []types.ChunkInfo
}

func cleanUpFile(f *os.File) {
	f.Close()
	os.Remove(f.Name())
}

func InitUploadFileOpts[MetadataT any](
	authData *types.AccessClaims, metadata MetadataT, r io.Reader,
) (*types.NewFileOpts[MetadataT], *os.File, error) {
	uploadID := uuid.New()

	tempFile, err := os.CreateTemp("", uploadID.String())
	if err != nil {
		return nil, nil, werr.Wrapf(err, "failed to create temporary file")
	}

	analyzed, err := uploadUtils.WriteAndAnalyze(r, tempFile)
	if err != nil {
		return nil, tempFile, werr.Wrapf(err, "failed to write and analyze upload file")
	}
	return &types.NewFileOpts[MetadataT]{
		UploadID:            uploadID,
		AuthPayload:         authData,
		FileMetadata:        metadata,
		DetectedContentType: analyzed.ContentType,
		SizeBytes:           analyzed.WrittenBytes,
		Blake3Sum:           analyzed.Blake3Sum,
		FilePath:            tempFile.Name(),
		Reader:              tempFile,
	}, tempFile, nil
}

func Handle[MetadataT any](
	ctx context.Context,
	deps *UploadDeps[MetadataT],
	processFunc process.ProcessFunc,
	fileOpts *types.NewFileOpts[MetadataT],
	processFileOpts *types.ProcessFileOpts,
) (uuid.UUID, error) {
	// Deduplication check
	existingUploadID, err := deps.RepoHelpers.CheckUploaded(ctx, fileOpts)
	if err != nil && !errors.Is(err, xerr.ErrEntityNotFound) {
		return uuid.Nil, werr.Wrapf(err, "failed while executing check upload hash function")
	}
	if existingUploadID != uuid.Nil {
		// file exists
		return existingUploadID, nil
	}

	return ProcessUpload(ctx, deps, processFunc, fileOpts, processFileOpts)
}

func ProcessUpload[MetadataT any](
	ctx context.Context,
	deps *UploadDeps[MetadataT],
	processFunc process.ProcessFunc,
	fileOpts *types.NewFileOpts[MetadataT],
	processFileOpts *types.ProcessFileOpts,
) (uuid.UUID, error) {
	variantFuncs, err := processFunc(ctx, processFileOpts)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get variant functions: %w", err)
	}
	if len(variantFuncs) == 0 {
		return fileOpts.UploadID, werr.Wrapf(xerr.ErrInvalidAction, "cannot process upload, no process functions for file upload found")
	}

	results, errChan := make(chan VariantResult), make(chan error)
	var variantsCreated []types.FileVariantInfo
	var chunksCreated []types.ChunkInfo
	var fileRecordCreated bool
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(ctx)

	// Process variants concurrently
	wg.Add(len(variantFuncs))
	for _, variantFunc := range variantFuncs {
		deps.WorkerPool.Exec(ctx, func(ctx context.Context) {
			defer wg.Done()
			if variants, chunks, err := variantFunc(ctx, deps.Storage); err != nil && err != context.Canceled {
				select {
				case errChan <- err:
				case <-ctx.Done():
				}
			} else {
				results <- VariantResult{Variants: variants, Chunks: chunks}
			}
		})
	}
	go func() { wg.Wait(); close(results) }()

	cleanup := func() {
		cancel() // firstly we stop all the workers
		CleanUp(ctx, deps.Storage, fileOpts.UploadID, variantsCreated, chunksCreated, []string{})
	}

	for {
		select {
		case result, ok := <-results:
			if !ok {
				if err := deps.RepoHelpers.RecordNewFile(ctx, fileOpts, variantsCreated, chunksCreated); err != nil {
					cancel()
					CleanUp(ctx, deps.Storage, fileOpts.UploadID, variantsCreated, chunksCreated, []string{})
					return uuid.Nil, err
				}
				cancel()
				return fileOpts.UploadID, nil
			}
			variantsCreated = append(variantsCreated, result.Variants...)
			chunksCreated = append(chunksCreated, result.Chunks...)
			if fileRecordCreated {
				if err := deps.RepoHelpers.AddVariants(ctx, fileOpts.UploadID, result.Variants, result.Chunks); err != nil {
					cleanup()
					return uuid.Nil, err
				}
			}
		case <-time.After(deps.Opts.CreateUploadRecordAfter):
			if err := deps.RepoHelpers.RecordNewFile(ctx, fileOpts, variantsCreated, chunksCreated); err != nil {
				cleanup()
				return uuid.Nil, err
			}
			fileRecordCreated = true
		case err := <-errChan:
			cleanup()
			return uuid.Nil, err
		}
	}
}
