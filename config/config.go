package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/morzisorn/gofermart/internal/logger"
	"go.uber.org/zap"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string

	SecretKey             string
	RateLimit             int //Processing workers rate limit
	LoyaltyUpdateInterval int //Loyalty update interval in seconds
}

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		var err error
		instance, err = New()
		if err != nil {
			logger.Log.Error("Error getting service", zap.Error(err))
		}
	})
	return instance
}

func New() (*Config, error) {
	envPath := getEncFilePath()

	if err := loadEnvFile(envPath); err != nil {
		fmt.Printf("Load .env error: %v. Env path: %s\n", err, envPath)
	}

	c := &Config{}

	if err := c.parseEnv(); err != nil {
		return c, fmt.Errorf("error parsing env: %v", err)
	}

	return c, nil
}

func getEncFilePath() string {
	basePath, err := GetProjectRoot()
	if err != nil {
		logger.Log.Error("Error getting project root ", zap.Error(err))
		return ".env"
	}
	return filepath.Join(basePath, "config", ".env")
}

func GetProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("project root not found")
		}
		wd = parent
	}
}
