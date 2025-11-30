package uploadUtils

import (
	"io"
	"net/http"

	"github.com/safeblock-dev/werr"
	"lukechampine.com/blake3"
)

type AnalyzedFileData struct {
	ContentType  string
	WrittenBytes int64
	Blake3Sum    []byte
}

// Writes everything from reader, detects content type and calculates blake3 hash sum
func WriteAndAnalyze(r io.Reader, f io.WriteSeeker) (data *AnalyzedFileData, err error) {
	hasher := blake3.New(64, nil)
	multiWriter := io.MultiWriter(f, hasher)

	// --- Detect content type by the first 512 bytes ---
	const sniffLen = 512
	header := make([]byte, sniffLen)

	sniffBytesWritten, err := io.ReadFull(r, header)
	sniffBytes := header[:sniffBytesWritten]
	if err == io.EOF {
		return nil, werr.Wrapf(err, "0 bytes were read. Is reader empty?")
	} else if err == io.ErrUnexpectedEOF {
		if _, err := multiWriter.Write(sniffBytes); err != nil {
			return nil, werr.Wrapf(err, "failed to write header to writer or hasher")
		}
		return &AnalyzedFileData{
			WrittenBytes: int64(sniffBytesWritten),
			ContentType:  http.DetectContentType(sniffBytes),
			Blake3Sum:    hasher.Sum(nil),
		}, nil
	} else if err != nil {
		return nil, werr.Wrapf(err, "failed to read for content sniff")
	}

	if _, err := multiWriter.Write(sniffBytes); err != nil {
		return nil, werr.Wrapf(err, "failed to write header to writer or hasher")
	}

	writtenBytes, err := io.Copy(multiWriter, r)
	if err != nil {
		return nil, werr.Wrapf(err, "failed to copy to writer or hasher")
	}
	writtenBytes += int64(sniffBytesWritten)

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return nil, werr.Wrapf(err, "failed to seek back to beginning of file")
	}
	return &AnalyzedFileData{
		WrittenBytes: writtenBytes,
		ContentType:  http.DetectContentType(sniffBytes),
		Blake3Sum:    hasher.Sum(nil),
	}, nil
}
