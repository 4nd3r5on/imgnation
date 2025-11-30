package uploadUtils

import (
	"io"

	"imgnation-backend/pkg/xerr"

	"github.com/safeblock-dev/werr"
)

func CreateLimitedReader(reader io.Reader, limitUpload bool, maxSize int64) (io.Reader, *io.LimitedReader) {
	if !limitUpload {
		return reader, nil
	}

	lr := &io.LimitedReader{
		R: reader,
		N: maxSize + 1, // +1 to detect overflow
	}
	return lr, lr
}

func ValidateSizeConstraints(lr *io.LimitedReader, limitUploadSize bool, maxUploadSize int64) error {
	if limitUploadSize && lr != nil && lr.N <= 0 {
		return werr.Wrapf(xerr.ErrInvalidArgument,
			"exceeds maximum allowed size of %d bytes", maxUploadSize)
	}
	return nil
}
