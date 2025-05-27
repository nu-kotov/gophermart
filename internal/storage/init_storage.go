package storage

import (
	"github.com/nu-kotov/gophermart/internal/config"
)

func NewStorage(c config.Config) (*DBStorage, error) {

	DBStorage, err := NewConnect(c.DatabaseConnection)
	if err != nil {
		return nil, err
	}

	return DBStorage, nil
}
