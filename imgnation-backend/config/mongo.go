package config

import "fmt"

type MongoCfg struct {
	Host     string
	Port     int
	User     string
	Password string
	DB       string
}

func NewMongoCfgFromEnv() *MongoCfg {
	return &MongoCfg{
		Host:     EnvOrDefault("MONGO_HOST", DefaultMongoHost),
		Port:     IntEnvOrDefault("MONGO_PORT", DefaultMongoPort),
		User:     EnvOrDefault("MONGO_USER", DefaultMongoUser),
		Password: MustEnv("MONGO_PASSWORD"),
		DB:       EnvOrDefault("MONGO_DB", DefaultMongoDB),
	}
}

func (cfg *MongoCfg) URI() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d", cfg.User, cfg.Password, cfg.Host, cfg.Port)
}
