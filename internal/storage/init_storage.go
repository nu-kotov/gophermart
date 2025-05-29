package storage

import (
	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/storage/postgres"
)

func NewStorage(c *config.Config) (*postgres.DBStorage, error) {

	DBStorage, err := postgres.NewConnect(c.DatabaseConnection)
	if err != nil {
		return nil, err
	}

	return DBStorage, nil
}
