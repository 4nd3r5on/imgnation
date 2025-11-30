package storage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"imgnation-backend/pkg/xerr"

	"github.com/minio/minio-go/v7"
	"github.com/safeblock-dev/werr"
)

type StoreObj struct {
	Key         string
	Size        int64
	Content     io.Reader
	ContentType string
}

type RetrievedObj struct {
	Key         string
	Size        int64
	ModTime     time.Time
	ContentType string
	Content     io.ReadSeekCloser
}

type IStorage interface {
	Exists(ctx context.Context, name string) (bool, error)
	BatchExists(ctx context.Context, names ...string) (map[string]bool, error)
	SetupBucket(ctx context.Context) error
	Store(ctx context.Context, obj StoreObj) error
	BatchStore(ctx context.Context, contentType string, objsChanIn chan StoreObj) error
	Get(ctx context.Context, name string) (RetrievedObj, error)
	Delete(ctx context.Context, ignore404 bool, keys ...string) []error
}

type Storage struct {
	bucket string
	client *minio.Client
}

func NewStorage(client *minio.Client, bucket string) *Storage {
	return &Storage{
		bucket: bucket,
		client: client,
	}
}

func (s *Storage) Exists(ctx context.Context, name string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, name, minio.StatObjectOptions{})
	if err != nil {
		var e minio.ErrorResponse
		if errors.As(err, &e) && e.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Storage) BatchExists(ctx context.Context, names ...string) (map[string]bool, error) {
	results := make(map[string]bool)
	for _, name := range names {
		exists, err := s.Exists(ctx, name)
		if err != nil {
			return nil, err
		}
		results[name] = exists
	}
	return results, nil
}

func (s *Storage) SetupBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil || exists {
		return err
	}
	return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
}

func (s *Storage) Store(ctx context.Context, obj StoreObj) error {
	_, err := s.client.PutObject(
		ctx, s.bucket,
		obj.Key, obj.Content, obj.Size,
		minio.PutObjectOptions{ContentType: obj.ContentType},
	)
	return err
}

func (s *Storage) BatchStore(ctx context.Context, contentType string, objsChanIn chan StoreObj) error {
	objsChan := make(chan minio.SnowballObject, 1)

	go func() {
		for {
			obj, ok := <-objsChanIn
			if !ok {
				close(objsChan)
				break
			}
			headers := http.Header{}
			if obj.ContentType != "" {
				headers.Set("Content-Type", obj.ContentType)
			}
			objsChan <- minio.SnowballObject{
				Key:     obj.Key,
				Size:    obj.Size,
				Content: obj.Content,
				Headers: headers,
			}
		}
	}()

	err := s.client.PutObjectsSnowball(ctx, s.bucket, minio.SnowballOptions{
		Opts: minio.PutObjectOptions{
			ContentType: contentType,
		},
	}, objsChan)
	return err
}

func (s *Storage) Get(ctx context.Context, name string) (RetrievedObj, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return RetrievedObj{}, err
	}
	attr, err := obj.Stat()
	if err != nil {
		errResp, ok := err.(minio.ErrorResponse)
		if ok && errResp.StatusCode == http.StatusNotFound {
			return RetrievedObj{}, werr.Wrapf(xerr.ErrEntityNotFound, "object wasn't found")
		}
		obj.Close()
		return RetrievedObj{}, err
	}

	return RetrievedObj{
		Key:         attr.Key,
		ContentType: attr.ContentType,
		ModTime:     attr.LastModified,
		Size:        attr.Size,
		Content:     obj,
	}, nil
}

func (s *Storage) Delete(ctx context.Context, ignore404 bool, keys ...string) []error {
	if len(keys) == 0 {
		return nil
	}

	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for _, name := range keys {
			objectsCh <- minio.ObjectInfo{Key: name}
		}
	}()

	errorCh := s.client.RemoveObjects(ctx, s.bucket, objectsCh, minio.RemoveObjectsOptions{})

	errs := []error{}
	for e := range errorCh {
		if e.Err != nil {
			if ignore404 {
				err, ok := e.Err.(minio.ErrorResponse)
				if ok && err.StatusCode == http.StatusNotFound {
					continue
				}
			}
			errs = append(errs, e.Err)
		}
	}
	return errs
}
