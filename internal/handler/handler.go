package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/models"
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

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var jsonBody models.UserData
	if err = json.Unmarshal(body, &jsonBody); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	passwordHash, err := argon2id.CreateHash(jsonBody.Password, argon2id.DefaultParams)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonBody.Password = passwordHash

	err = srv.Storage.InsertUserData(req.Context(), &jsonBody)
	if err != nil {
		http.Error(res, "Register user error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, "User registered")
}
