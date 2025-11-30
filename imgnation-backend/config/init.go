package config

import (
	"imgnation-backend/pkg/utils"
	usersRepo "imgnation-backend/repository/users"
)

type InitAdminCfg struct {
	Name     string   `yaml:"name"`
	Username string   `yaml:"username"`
	Email    string   `yaml:"email"`
	Password string   `yaml:"password"`
	Roles    []string `yaml:"roles"`
}

func (cfg *InitAdminCfg) Process() *usersRepo.NewUserReqData {
	return &usersRepo.NewUserReqData{
		Name:     utils.Default(cfg.Name, "admin"),
		Username: utils.Default(cfg.Username, "admin"),
		Email:    utils.Default(cfg.Email, "admin@example.com"),
		Password: utils.Default(cfg.Email, "password"),
		Roles:    append(cfg.Roles, AdminRole),
	}
}
