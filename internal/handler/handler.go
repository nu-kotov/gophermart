package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/models"
	"github.com/nu-kotov/gophermart/internal/storage"
	"github.com/phedde/luhn-algorithm"
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

func (srv *Service) CreateOrder(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}

	intBody, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}

	if isValid := luhn.IsValid(intBody); !isValid {
		http.Error(res, "Invalid order number", http.StatusBadRequest)
		return
	}

	orderData := models.OrderData{
		Number:     intBody,
		UserID:     uuid.New().String(),
		Status:     "NEW",
		UploadedAt: time.Now(),
	}
	err = srv.Storage.InsertOrderData(req.Context(), &orderData)
	if err != nil {
		http.Error(res, "Create order error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusAccepted)
}
