package storage

import (
	"context"

	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/models"
)

type Storage interface {
	InsertUserData(context.Context, *models.UserData) error
	InsertOrderData(context.Context, *models.OrderData) error
	SelectUserData(context.Context, *models.UserData) (*models.UserData, error)
}

func NewStorage(c config.Config) (Storage, error) {

	DBStorage, err := NewConnect(c.DatabaseConnection)
	if err != nil {
		return nil, err
	}

	return DBStorage, nil
}
