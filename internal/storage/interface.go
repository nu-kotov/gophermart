package storage

import (
	"context"

	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/models"
)

type Storage interface {
	InsertUserData(ctx context.Context, data *models.UserData) error
	InsertOrderData(ctx context.Context, data *models.OrderData) error
}

func NewStorage(c config.Config) (Storage, error) {

	DBStorage, err := NewConnect(c.DatabaseConnection)
	if err != nil {
		return nil, err
	}

	return DBStorage, nil
}
