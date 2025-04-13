package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/morzisorn/gofermart/internal/logger"
	"go.uber.org/zap"
)

func loadEnvFile(envPath string) error {
	return godotenv.Load(envPath)
}

func (c *Config) parseEnv() error {
	err := c.parseFlags()
	if err != nil {
		logger.Log.Panic("Parse flags error ", zap.Error(err))
	}

	addr, err := getEnvString("RUN_ADDRESS")
	if err == nil {
		c.RunAddress = addr
	}

	dbstr, err := getEnvString("DATABASE_URI")
	if err == nil {
		c.DatabaseURI = dbstr
	}

	accrual, err := getEnvString("ACCRUAL_SYSTEM_ADDRESS")
	if err == nil {
		if strings.Contains(accrual, "http://") {
			spl := strings.Split(accrual, "//")
			c.AccrualSystemAddress = spl[1]
		}
	}

	key, err := getEnvString("SECRET_KEY")
	if err == nil {
		c.SecretKey = key
	}

	rateLimit, err := getEnvInt("RATE_LIMIT")
	if err == nil {
		c.RateLimit = int(rateLimit)
	}

	interval, err := getEnvInt("LOYALTY_UPDATE_INTERVAL")
	if err == nil {
		c.LoyaltyUpdateInterval = int(interval)
	}

	return nil
}

func getEnvString(key string) (string, error) {
	env := os.Getenv(key)
	if env != "" {
		return env, nil
	}
	return "", fmt.Errorf("env %s not found", key)
}

func getEnvInt(key string) (int64, error) {
	env := os.Getenv(key)
	if env != "" {
		return strconv.ParseInt(env, 10, 64)
	}
	return 0, fmt.Errorf("env %s not found", key)
}
