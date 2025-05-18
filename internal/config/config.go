package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddr            string
	DatabaseConnection string
	AccrualAddr        string
}

func NewConfig() Config {
	var config Config

	flag.StringVar(&config.RunAddr, "a", "localhost:8181", "address and port to run server")
	flag.StringVar(&config.DatabaseConnection, "d", "", "Database connection string")
	flag.StringVar(&config.AccrualAddr, "r", "http://localhost:8888", "default schema, host and port in compressed URL")

	flag.Parse()

	if envRunAddr, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		config.RunAddr = envRunAddr
	}
	if envDatabaseConnection, ok := os.LookupEnv("DATABASE_URI"); ok {
		config.DatabaseConnection = envDatabaseConnection
	}
	if envAccrualAddr, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		config.DatabaseConnection = envAccrualAddr
	}

	return config
}
