package handler

import (
	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/models"
	"github.com/nu-kotov/gophermart/internal/storage/postgres"
)

type Handler struct {
	Config              *config.Config
	Storage             *postgres.DBStorage
	SaveAccrualPointsCh chan models.OrderData
}

func NewHandler(config *config.Config, storage *postgres.DBStorage) *Handler {
	var hnd Handler

	hnd.Config = config
	hnd.Storage = storage
	hnd.SaveAccrualPointsCh = make(chan models.OrderData, 1024)

	go hnd.GetAccrualPoints()
	go hnd.SaveOrdersPoints()

	return &hnd
}
