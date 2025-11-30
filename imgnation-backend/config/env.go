package config

import (
	"log"
	"os"

	"imgnation-backend/pkg/utils"
)

func EnvOrDefault(envName string, defaultVal string) string {
	return utils.Default(os.Getenv(envName), defaultVal)
}

func MustEnv(envName string) string {
	val := os.Getenv(envName)
	if val == "" {
		log.Fatalf("Required environment varialbe %s wasn't set", envName)
	}
	return val
}

func IntEnvOrDefault[T utils.Int](envName string, defaultVal T) T {
	val := os.Getenv(envName)
	if val == "" {
		return defaultVal
	}
	parsed, err := utils.ParseInt[T](val)
	if err != nil {
		return defaultVal
	}
	return parsed
}

func FloatEnvOrDefault[T utils.Float](envName string, defaultVal T) T {
	val := os.Getenv(envName)
	if val == "" {
		return defaultVal
	}
	parsed, err := utils.ParseFloat[T](val)
	if err != nil {
		return defaultVal
	}
	return parsed
}
