package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/nu-kotov/gophermart/internal/auth"
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
	jsonBody.UserID = uuid.New().String()

	err = srv.Storage.InsertUserData(req.Context(), &jsonBody)
	if err != nil {
		http.Error(res, "Register user error", http.StatusInternalServerError)
		return
	}

	value, err := auth.BuildJWTString(jsonBody.UserID, jsonBody.Login)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	cookie := &http.Cookie{
		Name:     "token",
		Value:    value,
		HttpOnly: true,
	}

	http.SetCookie(res, cookie)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, "User registered")
}

func (srv *Service) LoginUser(res http.ResponseWriter, req *http.Request) {
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

	userData, err := srv.Storage.SelectUserData(req.Context(), &jsonBody)
	if err != nil {
		http.Error(res, "Get user password error", http.StatusInternalServerError)
		return
	}

	match, err := argon2id.ComparePasswordAndHash(jsonBody.Password, userData.Password)
	if err != nil {
		http.Error(res, "Password comparing error", http.StatusInternalServerError)
		return
	}
	if !match {
		http.Error(res, "Uncorrect passwort or login", http.StatusUnauthorized)
		return
	}

	value, err := auth.BuildJWTString(userData.UserID, userData.Login)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	cookie := &http.Cookie{
		Name:     "token",
		Value:    value,
		HttpOnly: true,
	}

	http.SetCookie(res, cookie)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, "User authorized")
}

func (srv *Service) CreateOrder(res http.ResponseWriter, req *http.Request) {

	token, err := req.Cookie("token")

	if err != nil {
		res.WriteHeader(http.StatusUnauthorized)
	}

	userID, err := auth.GetUserID(token.Value)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

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
		UserID:     userID,
		Status:     "NEW",
		UploadedAt: time.Now(),
	}
	err = srv.Storage.InsertOrderData(req.Context(), &orderData)
	if err != nil {
		fmt.Println(err)
		http.Error(res, "Create order error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusAccepted)
}

func (srv *Service) GetUserOrders(res http.ResponseWriter, req *http.Request) {
	token, err := req.Cookie("token")

	if err != nil {
		res.WriteHeader(http.StatusUnauthorized)
	}

	userID, err := auth.GetUserID(token.Value)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}
	fmt.Println(userID)
	res.Header().Set("Content-Type", "application/json")
	if data, err := srv.Storage.SelectOrdersByUserID(req.Context(), userID); data != nil {

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		resp, err := json.Marshal(data)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		_, err = res.Write(resp)

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

	} else {
		res.WriteHeader(http.StatusNoContent)
	}
}

func (srv *Service) GetUserBalance(res http.ResponseWriter, req *http.Request) {
	token, err := req.Cookie("token")

	if err != nil {
		res.WriteHeader(http.StatusUnauthorized)
	}

	userID, err := auth.GetUserID(token.Value)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}
	res.Header().Set("Content-Type", "application/json")
	if data, err := srv.Storage.SelectUserBalance(req.Context(), userID); data != nil {

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		resp, err := json.Marshal(data)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		_, err = res.Write(resp)

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

	} else {
		res.WriteHeader(http.StatusNoContent)
	}
}

func (srv *Service) WithdrawPoints(res http.ResponseWriter, req *http.Request) {
	token, err := req.Cookie("token")

	if err != nil {
		res.WriteHeader(http.StatusUnauthorized)
	}

	userID, err := auth.GetUserID(token.Value)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}

	var jsonBody models.WithdrawnInfo
	if err = json.Unmarshal(body, &jsonBody); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(jsonBody.Number)

	intNumber, err := strconv.ParseInt(jsonBody.Number, 10, 64)
	if err != nil {
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}

	if isValid := luhn.IsValid(intNumber); !isValid {
		http.Error(res, "Invalid order number", http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	data, err := srv.Storage.SelectUserBalance(req.Context(), userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	fmt.Println(data.Balance)
	if data.Balance < jsonBody.Sum {
		http.Error(res, "Insufficient funds", http.StatusPaymentRequired)
		return
	}
	data.Balance = data.Balance - jsonBody.Sum
	data.Withdrawn = data.Withdrawn + jsonBody.Sum

	fmt.Println(data.Balance)
	withdraw := models.Withdraw{
		Number:      intNumber,
		UserID:      userID,
		Sum:         jsonBody.Sum,
		WithdrawnAt: time.Now(),
	}
	err = srv.Storage.UpdateUserBalance(req.Context(), data, &withdraw)
	if err != nil {
		http.Error(res, "User update error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusOK)
}

func (srv *Service) GetUserWithdrawals(res http.ResponseWriter, req *http.Request) {
	token, err := req.Cookie("token")

	if err != nil {
		res.WriteHeader(http.StatusUnauthorized)
	}

	userID, err := auth.GetUserID(token.Value)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	res.Header().Set("Content-Type", "application/json")
	fmt.Println(userID)
	if data, err := srv.Storage.SelectUserWithdrawals(req.Context(), userID); data != nil {

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		resp, err := json.Marshal(data)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		_, err = res.Write(resp)

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

	} else {
		res.WriteHeader(http.StatusNoContent)
	}
}
