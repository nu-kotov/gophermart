package handler

import (
	"net/http"

	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/storage"
)

type Service struct {
	Config  config.Config
	Storage storage.Storage
}

func NewService(config config.Config, storage storage.Storage) *Service {
	var srv Service

	srv.Config = config
	srv.Storage = storage

	return &srv
}

func (srv *Service) RegisterUser(res http.ResponseWriter, req *http.Request) {

}
