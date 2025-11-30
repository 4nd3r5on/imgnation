package upload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strings"

	"imgnation-backend/config"
	"imgnation-backend/pkg/types"
	"imgnation-backend/pkg/xerr"
	"imgnation-backend/repository"

	usersRepo "imgnation-backend/repository/users"

	uploadUtils "imgnation-backend/pkg/utils/upload"

	"github.com/safeblock-dev/werr"
)

type AccessControl struct {
	Level             string   `bson:"level" json:"level"`
	AllowedUserEmails []string `bson:"allowed_user_emails,omitempty" json:"allowed_user_emails,omitempty"`
}

var AccessLevelMap = map[string]types.AccessLevel{
	"anyone":     types.AccessLevelAnyone,
	"authorized": types.AccessLevelAuthorized,
	"private":    types.AccessLevelPrivate,
	"admin":      types.AccessLevelAdmin,
}

type ReqUploadMetadata struct {
	Tags        []string      `json:"tags"`
	Public      bool          `json:"public"`
	Access      AccessControl `json:"access"`
	Description string        `json:"description,omitempty"`
	/*
	  To be added later:
	  Encryption & permissions features later:
	    EncryptionUsed      bool   `json:"encryption_used"`
	    EncryptionAlgorithm string `json:"encryption_algorithm"`
	    PubKey              string `json:"pub_key"`  // For EC25519 or RSA
	    KeyHash             string `json:"key_hash"` // For AES
	*/
}

type UploadMetadata struct {
	Tags              []string
	FileName          string
	HeaderContentType string
	Access            types.AccessControl
	Description       string
}

type UploadResult struct {
	Successful bool   `json:"successful,omitempty"`
	UploadID   string `json:"upload_id,omitempty"`
	FileName   string `json:"filename,omitempty"`
	Error      string `json:"error,omitempty"`
}

func DiscardFilePart(p *multipart.Part) error {
	if _, err := io.Copy(io.Discard, p); err != nil && !errors.Is(err, io.EOF) {
		return werr.Wrapf(err, "failed to discard part")
	}
	if err := p.Close(); err != nil {
		return werr.Wrapf(err, "failed to close part")
	}
	return nil
}

func DiscardFilePartMaybe(p *multipart.Part) {
	if _, err := io.Copy(io.Discard, p); err != nil && !errors.Is(err, io.EOF) {
		log.Printf("Warning: failed to discard part data: %v", err)
	}
	if err := p.Close(); err != nil {
		log.Printf("Warning: failed to close part: %v", err)
	}
}

func FileOptsToProcessOpts(opts *types.NewFileOpts[*UploadMetadata]) *types.ProcessFileOpts {
	return &types.ProcessFileOpts{
		UploadID:            opts.UploadID,
		HeaderContentType:   opts.FileMetadata.HeaderContentType,
		DetectedContentType: opts.DetectedContentType,
		SizeBytes:           opts.SizeBytes,
		Blake3Sum:           opts.Blake3Sum,
		FilePath:            opts.FilePath,
		Reader:              opts.Reader,
	}
}

func ParseMetadata(ctx context.Context, part *multipart.Part, users usersRepo.UsersRepo, authData *types.AccessClaims, rawMetadata json.RawMessage) (*UploadMetadata, error) {
	fileName := part.FileName()
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return nil, fmt.Errorf("part filename shouldn't be empty")
	}
	headerContentType := part.Header.Get("Content-Type")
	headerContentType = strings.TrimSpace(headerContentType)
	if headerContentType == "" {
		return nil, fmt.Errorf("part header Content-Type shouldn't be empty")
	}

	var reqMetadata ReqUploadMetadata
	if err := json.Unmarshal(rawMetadata, &reqMetadata); err != nil {
		return nil, werr.Wrapf(err, "failed to unmarshal metadata part")
	}
	if err := ValidateDescription(reqMetadata.Description); err != nil {
		return nil, err
	}
	access, err := ProcessAccessMetadata(ctx, users, authData, reqMetadata.Access)
	if err != nil {
		return nil, werr.Wrapf(err, "failed to process access metadata")
	}
	return &UploadMetadata{
		Tags:              ProcessTags(reqMetadata.Tags),
		FileName:          fileName,
		HeaderContentType: headerContentType,
		Access:            access,
		Description:       reqMetadata.Description,
	}, nil
}

func HandleUpload(
	ctx context.Context,
	deps *UploadDeps[*UploadMetadata],
	repo *repository.Repo,
	cfg *config.ServerCfg,
	authData *types.AccessClaims,
	mr *multipart.Reader,
) (results []UploadResult, err error) {
	limits, err := GetUploadLimits(cfg, authData)
	if err != nil {
		return nil, werr.Wrapf(err, "failed to calculate upload limits")
	}
	if !limits.AllowUpload {
		return nil, werr.Wrapf(xerr.ErrPermissionDenied, "upload is not allowed")
	}

	metadata, err := uploadUtils.HandleMetadataPart(mr, limits.LimitUploadSize, limits.MaxUploadSize)
	if err != nil {
		return nil, err
	}
	results = make([]UploadResult, len(metadata.Metadata))

	uploadUtils.HandleFileParts(ctx, mr, metadata.Metadata, 0,
		func(ctx context.Context, part *multipart.Part, rawMetadata json.RawMessage, filePartIdx int) error {
			reader, lr := uploadUtils.CreateLimitedReader(part, limits.LimitUploadSize, limits.MaxUploadSize)
			metadata, err := ParseMetadata(ctx, part, repo.UsersRepo, authData, rawMetadata)
			if err != nil {
				results[filePartIdx] = UploadResult{
					Error: fmt.Sprintf("failed to parse metadata (file part %d): %s", filePartIdx, err.Error()),
				}
				if err := DiscardFilePart(part); err != nil {
					return werr.Wrapf(err, "failed to discard file part %d", filePartIdx)
				}
				return nil
			}

			fileOpts, tempFile, err := InitUploadFileOpts(authData, metadata, reader)
			if err != nil {
				results[filePartIdx] = UploadResult{
					Error: fmt.Sprintf("failed to init upload options (file part %d): %s", filePartIdx, err.Error()),
				}
				DiscardFilePartMaybe(part)
				return nil
			}
			defer cleanUpFile(tempFile)

			processFileOpts := FileOptsToProcessOpts(fileOpts)

			processFunc, err := GetProcessFunc(
				limits.AllowAnyUploadType,
				limits.AllowUploadTypes,
				deps.Processors,
				processFileOpts)
			if err != nil {
				results[filePartIdx] = UploadResult{
					Error: fmt.Sprintf("failed to get process func file upload (file part %d): %s", filePartIdx, err.Error()),
				}
				DiscardFilePartMaybe(part)
				return nil
			}

			uploadID, err := Handle(ctx, deps, processFunc, fileOpts, processFileOpts)
			if err != nil {
				results[filePartIdx] = UploadResult{
					Error: fmt.Sprintf("failed to handle file upload (file part %d): %s", filePartIdx, err.Error()),
				}
				DiscardFilePartMaybe(part)
				return nil
			}

			err = uploadUtils.ValidateSizeConstraints(lr, limits.LimitUploadSize, limits.MaxUploadSize)
			if err != nil {
				results[filePartIdx] = UploadResult{
					FileName: metadata.FileName,
					Error:    fmt.Sprintf("failed to validate size constraints (file part %d): %s", filePartIdx, err.Error()),
				}
				DiscardFilePartMaybe(part)
				return nil
			}

			results[filePartIdx] = UploadResult{
				Successful: true,
				UploadID:   uploadID.String(),
				FileName:   metadata.FileName,
			}
			return nil
		})

	return results, nil
}
