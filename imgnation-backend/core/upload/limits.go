package upload

import (
	"maps"

	"imgnation-backend/config"
	"imgnation-backend/pkg/types"
)

type UploadLimits struct {
	AllowUpload        bool
	IsDefaultRole      bool
	LimitUploadSize    bool
	MaxUploadSize      int64
	AllowAnyUploadType bool
	AllowUploadTypes   map[string]struct{}
}

func GetUploadLimits(
	cfg *config.ServerCfg, claims *types.AccessClaims,
) (uploadLimits *UploadLimits, err error) {
	var roles []string
	var maxSize int64
	var allowAnyUploadType bool

	limitUploadSize := true
	isDefaultRole := true
	allowUploadTypes := map[string]struct{}{}

	if claims == nil {
		roles = []string{cfg.Permissions.DefaultRole}
	} else {
		roles = claims.Roles
	}

	for _, role := range roles {
		roleCfg, exists := cfg.Permissions.Roles[role]
		if !exists {
			continue
		}
		isDefaultRole = false
		// forbid upload
		if roleCfg.MaxUploadSize < 0 {
			return &UploadLimits{AllowUpload: false}, nil
		}
		// unlimited upload
		if roleCfg.MaxUploadSize == 0 {
			limitUploadSize = false
		}
		// set the biggest role privelage
		allowAnyUploadType = allowAnyUploadType || roleCfg.AnyUploadsAllowed
		maxSize = max(roleCfg.MaxUploadSize, maxSize)
		maps.Copy(allowUploadTypes, roleCfg.UploadsAllowed)
	}
	// Allow upload with limited size
	return &UploadLimits{
		AllowUpload:        true,
		IsDefaultRole:      isDefaultRole,
		LimitUploadSize:    limitUploadSize,
		MaxUploadSize:      maxSize,
		AllowAnyUploadType: allowAnyUploadType,
		AllowUploadTypes:   allowUploadTypes,
	}, nil
}
