package config

import (
	"log"
	"os"
	"path/filepath"

	"imgnation-backend/pkg/utils"
	usersRepo "imgnation-backend/repository/users"

	"github.com/safeblock-dev/werr"
	"gopkg.in/yaml.v3"
)

type AuthCfg struct {
	JwtAccessSecret  []byte
	JwtRefreshSecret []byte
}

func NewAuthCfgFromEnv() *AuthCfg {
	jwtAccessSecret := MustEnv("AUTH_JWT_ACCESS_SECRET")
	jwtRefreshSecret := MustEnv("AUTH_JWT_REFRESH_SECRET")
	return &AuthCfg{
		JwtAccessSecret:  []byte(jwtAccessSecret),
		JwtRefreshSecret: []byte(jwtRefreshSecret),
	}
}

type ServerCfg struct {
	Api         *ApiCfg
	Mongo       *MongoCfg
	Minio       *MinioCfg
	Auth        *AuthCfg
	Permissions *PermissionsCfg
	InitAdmin   *usersRepo.NewUserReqData
}

type ServerFileCfg struct {
	Api         ApiCfg             `yaml:"api"`
	Permissions PermissionsFileCfg `yaml:"permissions"`
	InitAdmin   *InitAdminCfg      `yaml:"init_admin"`
}

func LoadServerCfg() (*ServerCfg, error) {
	serverCfg := &ServerCfg{}
	fileCfg := &ServerFileCfg{}

	fileCfgPath := utils.Default(os.Getenv("CONFIG_PATH"), DefaultFileCfgPath)
	absoluteCfgPath, _ := filepath.Abs(fileCfgPath)
	log.Println("Reading config file:", utils.Default(absoluteCfgPath, fileCfgPath))

	data, err := os.ReadFile(fileCfgPath)
	if err != nil {
		return nil, werr.Wrapf(err, "failed to read config file")
	}
	if err := yaml.Unmarshal(data, fileCfg); err != nil {
		return nil, err
	}

	serverCfg.Permissions = fileCfg.Permissions.Process()
	serverCfg.Api = fileCfg.Api.Process()
	serverCfg.Auth = NewAuthCfgFromEnv()
	serverCfg.Mongo = NewMongoCfgFromEnv()

	minioRawCfg := NewMinioRawCfgFromEnv()
	serverCfg.Minio, err = minioRawCfg.Process()
	if err != nil {
		return nil, werr.Wrapf(err, "failed to process minio cfg")
	}

	serverCfg.InitAdmin = fileCfg.InitAdmin.Process()

	return serverCfg, nil
}
