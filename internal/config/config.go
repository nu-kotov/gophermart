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

	var RunAddrFlag string
	var DatabaseConnectionFlag string
	var AccrualAddrFlag string

	flag.StringVar(&RunAddrFlag, "a", "localhost:8181", "address and port to run server")
	flag.StringVar(&DatabaseConnectionFlag, "d", "", "Database connection string")
	flag.StringVar(&AccrualAddrFlag, "r", "http://localhost:8888", "default schema, host and port in compressed URL")

	flag.Parse()

	if envRunAddr, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		config.RunAddr = envRunAddr
	} else {
		config.RunAddr = RunAddrFlag
	}
	if envDatabaseConnection, ok := os.LookupEnv("DATABASE_URI"); ok {
		config.DatabaseConnection = envDatabaseConnection
	} else {
		config.DatabaseConnection = DatabaseConnectionFlag
	}
	if envAccrualAddr, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		config.AccrualAddr = envAccrualAddr
	} else {
		config.AccrualAddr = AccrualAddrFlag
	}

	return config
}
