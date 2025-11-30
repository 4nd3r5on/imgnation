package config

import "imgnation-backend/pkg/utils"

type CorsCfg struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

type ApiCfg struct {
	Host   string   `yaml:"host"`
	Port   int      `yaml:"port"`
	Prefix string   `yaml:"prefix"`
	Cors   *CorsCfg `yaml:"cors"`
}

func (cfg *CorsCfg) Process() *CorsCfg {
	// Apply defaults
	if len(cfg.AllowedOrigins) == 0 {
		cfg.AllowedOrigins = DefaultApiCorsAllowedOrigins
	}
	if cfg.AllowedMethods == nil {
		cfg.AllowedMethods = []string{}
	}
	if cfg.AllowedHeaders == nil {
		cfg.AllowedHeaders = []string{}
	}
	return cfg
}

func (cfg *ApiCfg) Process() *ApiCfg {
	// Apply defaults
	cfg.Host = utils.Default(cfg.Host, DefaultAPIHost)
	cfg.Port = utils.Default(cfg.Port, DefaultAPIPort)
	cfg.Prefix = utils.Default(cfg.Prefix, DefaultAPIPrefix)
	cfg.Cors = cfg.Cors.Process()
	return cfg
}
