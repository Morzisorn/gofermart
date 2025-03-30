package config

import (
	"os"

	"github.com/spf13/pflag"
)

func (c *Config) parseFlags() error {
	pflag.StringVarP(&c.RunAddress, "addr", "a", "localhost:8081", "address and port to run server")
	pflag.StringVarP(&c.DatabaseURI, "dbstr", "d", "", "db connection string")
	pflag.StringVarP(&c.AccrualSystemAddress, "accrual", "r", "", "accrual system address")

	pflag.StringVarP(&c.AccrualSystemAddress, "key", "k", "VERY_SECRET", "secret key")
	pflag.IntVarP(&c.RateLimit, "limit", "l", 5, "loyalty updater rate limit")
	pflag.IntVarP(&c.LoyaltyUpdateInterval, "interval", "i", 5, "loyalty update interval in seconds")

	return pflag.CommandLine.Parse(os.Args[1:])
}
