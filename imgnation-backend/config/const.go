package config

const DefaultFileCfgPath = "./config.yml"

const (
	DefaultAPIHost   = "127.0.0.1"
	DefaultAPIPort   = 80
	DefaultAPIPrefix = ""
)

const (
	DefaultRole              = "unauthorized"
	DefaultMaxUploadSize     = -1
	DefaultEncryptionAllowed = false
)

const (
	DefaultMongoHost = "localhost"
	DefaultMongoPort = 27017
	DefaultMongoUser = "mongo"
	DefaultMongoDB   = "default"
)

const (
	DefaultMinioHost = "localhost"
	DefaultMinioPort = 9000
	DefaultMinioUser = "minio"
)

var (
	DefaultUploadsAllowed        = []string{"image"}
	DefaultApiCorsAllowedOrigins = []string{"*"}
)

const (
	AdminRole = "admin"
)
