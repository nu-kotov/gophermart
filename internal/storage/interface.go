package storage

import (
	"github.com/nu-kotov/gophermart/internal/config"
)

type Storage interface {
}

func NewStorage(c config.Config) (Storage, error) {

	DBStorage, err := NewConnect(c.DatabaseConnection)
	if err != nil {
		return nil, err
	}

	return DBStorage, nil
}
