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

	flag.StringVar(&config.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&config.DatabaseConnection, "d", "", "Database connection string")
	flag.StringVar(&config.AccrualAddr, "r", "http://localhost:8888", "default schema, host and port in compressed URL")

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		config.RunAddr = envRunAddr
	}
	if envDatabaseConnection := os.Getenv("DATABASE_URI"); envDatabaseConnection != "" {
		config.DatabaseConnection = envDatabaseConnection
	}
	if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
		config.DatabaseConnection = envAccrualAddr
	}

	flag.Parse()

	return config
}
