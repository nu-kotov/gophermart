package storage

import (
	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/storage/postgres"
)

func NewPgStorage(c *config.Config) (*postgres.DBStorage, error) {

	DBStorage, err := postgres.NewConnect(c.DatabaseConnection)
	if err != nil {
		return nil, err
	}

	return DBStorage, nil
}

func NewBalanceStorage(pg *postgres.DBStorage) *postgres.BalanceStorage {
	return &postgres.BalanceStorage{Stor: pg}
}

func NewOrdersStorage(pg *postgres.DBStorage) *postgres.OrdersStorage {
	return &postgres.OrdersStorage{Stor: pg}
}

func NewUsersStorage(pg *postgres.DBStorage) *postgres.UsersStorage {
	return &postgres.UsersStorage{Stor: pg}
}

func NewWithdrawalsStorage(pg *postgres.DBStorage) *postgres.WithdrawalsStorage {
	return &postgres.WithdrawalsStorage{Stor: pg}
}
