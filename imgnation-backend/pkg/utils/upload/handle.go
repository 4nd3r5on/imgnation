package uploadUtils

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"time"

	"imgnation-backend/pkg/xerr"

	"github.com/safeblock-dev/werr"
)

type MetadataPartJson struct {
	Metadata []json.RawMessage `json:"metadata"`
}

func HandleMetadataPart(mr *multipart.Reader, limitSize bool, maxSize int64) (*MetadataPartJson, error) {
	part, err := mr.NextPart()
	switch {
	case errors.Is(err, io.EOF):
		return nil, werr.Wrapf(xerr.ErrInvalidArgument, "no parts provided")
	case err != nil:
		return nil, werr.Wrapf(err, "error reading first part")
	}

	reader, lr := CreateLimitedReader(part, limitSize, maxSize)

	if part.FormName() != "metadata" {
		return nil, werr.Wrapf(xerr.ErrInvalidArgument,
			"first part must be metadata (got %s)", part.FormName())
	}
	var metadata MetadataPartJson
	if err := json.NewDecoder(reader).Decode(&metadata); err != nil {
		return nil, werr.Wrapf(err, "failed to decode metadata")
	}
	if err := part.Close(); err != nil {
		return nil, werr.Wrapf(err, "failed to close metadata part")
	}
	err = ValidateSizeConstraints(lr, limitSize, maxSize)
	if err != nil {
		return nil, werr.Wrapf(err, "failed to parse metadata part")
	}
	return &metadata, nil
}

type HandleFilePart func(
	ctx context.Context,
	part *multipart.Part,
	// metadata will be parsed as raw message
	// so handler can decide how to parse it for itself
	// based on content type or any other things
	metadata json.RawMessage,
	filePartIdx int,
) error

func HandleFileParts(
	ctx context.Context,
	mr *multipart.Reader,
	metadata []json.RawMessage,
	handleTimeout time.Duration,
	handlePart HandleFilePart,
) (err error) {
	for i, m := range metadata {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return werr.Wrapf(err, "error reading next part (%d)", i)
		}
		ctx, cancel := NewTimeoutCtx(ctx, handleTimeout)
		err = handlePart(ctx, part, m, i)
		cancel()
		if err != nil {
			return werr.Wrapf(err, "failed to handle part (%d)", i)
		}
	}
	return nil
}

func NewTimeoutCtx(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	}
	return context.WithCancel(ctx)
}
