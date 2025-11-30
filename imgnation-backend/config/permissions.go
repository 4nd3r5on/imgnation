package config

import (
	"log"
	"strconv"

	"imgnation-backend/pkg/utils"
)

type RolePermissionsCfg struct {
	MaxUploadSize     int64
	AnyUploadsAllowed bool
	UploadsAllowed    map[string]struct{}
}

type PermissionsCfg struct {
	DefaultRole string
	Roles       map[string]RolePermissionsCfg
}

type RolePermissionsFileCfg struct {
	MaxUploadSize  string   `yaml:"max_upload_size"`
	UploadsAllowed []string `yaml:"uploads_allowed"`
}

type PermissionsFileCfg struct {
	DefaultRole string                            `yaml:"default_role"`
	Roles       map[string]RolePermissionsFileCfg `yaml:"roles"`
}

func (cfg *PermissionsFileCfg) Process() *PermissionsCfg {
	out := &PermissionsCfg{
		DefaultRole: utils.Default(cfg.DefaultRole, DefaultRole),
		Roles:       make(map[string]RolePermissionsCfg, len(cfg.Roles)),
	}

	for roleName, roleCfg := range cfg.Roles {
		uploadsAllowed := make(map[string]struct{})
		for _, upload := range roleCfg.UploadsAllowed {
			uploadsAllowed[upload] = struct{}{}
		}

		var maxUploadSize int64
		maxUploadSize, err := strconv.ParseInt(roleCfg.MaxUploadSize, 10, 64)
		if err != nil {
			maxUploadSize = utils.ParseStrSize(roleCfg.MaxUploadSize)
			if maxUploadSize == -1 {
				log.Fatalf("Failed to parse max_upload_size in config for role %s", roleName)
			}
		}

		out.Roles[roleName] = RolePermissionsCfg{
			MaxUploadSize:     maxUploadSize,
			AnyUploadsAllowed: len(uploadsAllowed) == 0,
			UploadsAllowed:    uploadsAllowed,
		}
	}
	return out
}
