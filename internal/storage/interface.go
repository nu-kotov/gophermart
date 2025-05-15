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
	SelectOrdersByUserID(context.Context, string) ([]models.GetUserOrdersResponse, error)
	SelectUserBalance(context.Context, string) (*models.UserBalance, error)
	UpdateUserBalance(context.Context, *models.UserBalance, *models.Withdraw) error
	SelectUserWithdrawals(context.Context, string) ([]models.WithdrawnInfo, error)
	SelectUnprocessedOrders(ctx context.Context) ([]string, error)
}

func NewStorage(c config.Config) (Storage, error) {

	DBStorage, err := NewConnect(c.DatabaseConnection)
	if err != nil {
		return nil, err
	}

	return DBStorage, nil
}
