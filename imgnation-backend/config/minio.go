package config

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/safeblock-dev/werr"
)

type MinioCredentialsFileData struct {
	URL       string `json:"url,omitempty"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	API       string `json:"api,omitempty"`
	Path      string `json:"path,omitempty"`
}

func ReadMinioCredentialsJSON(filePath string) (*MinioCredentialsFileData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, werr.Wrapf(err, "error while trying to open credentials file")
	}

	out := &MinioCredentialsFileData{}
	if err := json.NewDecoder(file).Decode(&out); err != nil {
		return nil, werr.Wrapf(err, "error while trying to decode json file")
	}
	if err := file.Close(); err != nil {
		return nil, werr.Wrapf(err, "error while trying to close credentials file")
	}
	return out, nil
}

type MinioRawCfg struct {
	UseCredentialsFile  bool
	CredentialsFilePath string

	AccessKey string
	SecretKey string

	Endpoint string
	UseSSL   bool
}

func NewMinioRawCfgFromEnv() *MinioRawCfg {
	credsPath := os.Getenv("MINIO_CREDENTIALS_FILE_PATH")

	useSSLStr := os.Getenv("MINIO_USE_SSL")
	useSSL := true // default to true
	if useSSLStr != "" {
		parsed, err := strconv.ParseBool(useSSLStr)
		if err != nil {
			log.Fatalf("Invalid boolean value for MINIO_USE_SSL: %v", err)
		}
		useSSL = parsed
	}

	return &MinioRawCfg{
		UseCredentialsFile:  credsPath != "",
		CredentialsFilePath: credsPath,
		AccessKey:           os.Getenv("ACCESS_KEY"),
		SecretKey:           os.Getenv("SECRET_KEY"),
		Endpoint:            MustEnv("MINIO_ENDPOINT"),
		UseSSL:              useSSL,
	}
}

func (cfg *MinioRawCfg) Process() (*MinioCfg, error) {
	var accessKey, secretKey string

	if cfg.UseCredentialsFile {
		creds, err := ReadMinioCredentialsJSON(cfg.CredentialsFilePath)
		if err != nil {
			return nil, werr.Wrapf(err, "failed to read minio credentials file (%s)", cfg.CredentialsFilePath)
		}
		accessKey, secretKey = creds.AccessKey, creds.SecretKey
	} else {
		accessKey, secretKey = cfg.AccessKey, cfg.SecretKey
	}
	return &MinioCfg{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Endpoint:  cfg.Endpoint,
		UseSSL:    cfg.UseSSL,
	}, nil
}

type MinioCfg struct {
	AccessKey string
	SecretKey string

	Endpoint string
	UseSSL   bool
}
