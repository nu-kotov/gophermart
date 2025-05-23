package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/nu-kotov/gophermart/internal/auth"
	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/logger"
	"github.com/nu-kotov/gophermart/internal/models"
	"github.com/nu-kotov/gophermart/internal/storage"
	"github.com/phedde/luhn-algorithm"
)

type Service struct {
	Config              config.Config
	Storage             storage.Storage
	SaveAccrualPointsCh chan models.Orders
}

func NewService(config config.Config, storage storage.Storage) *Service {
	var srv Service

	srv.Config = config
	srv.Storage = storage
	srv.SaveAccrualPointsCh = make(chan models.Orders, 1024)

	go srv.GetAccrualPoints()

	return &srv
}

func (srv *Service) RegisterUser(res http.ResponseWriter, req *http.Request) {

	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var jsonBody models.UserData
	if err = json.Unmarshal(body, &jsonBody); err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	passwordHash, err := argon2id.CreateHash(jsonBody.Password, argon2id.DefaultParams)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonBody.Password = passwordHash
	jsonBody.UserID = uuid.New().String()

	err = srv.Storage.InsertUserData(req.Context(), &jsonBody)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, "Register user error", http.StatusInternalServerError)
		return
	}

	value, err := auth.BuildJWTString(jsonBody.UserID, jsonBody.Login)
	if err != nil {
		logger.Log.Info(err.Error())
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
	io.WriteString(res, fmt.Sprintf("User %s registered", jsonBody.Login))
}

func (srv *Service) LoginUser(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var jsonBody models.UserData
	if err = json.Unmarshal(body, &jsonBody); err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	userData, err := srv.Storage.SelectUserData(req.Context(), &jsonBody)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, "Get user password error", http.StatusInternalServerError)
		return
	}

	match, err := argon2id.ComparePasswordAndHash(jsonBody.Password, userData.Password)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, "Password comparing error", http.StatusInternalServerError)
		return
	}
	if !match {
		http.Error(res, "Uncorrect passwort or login", http.StatusUnauthorized)
		return
	}

	value, err := auth.BuildJWTString(userData.UserID, userData.Login)
	if err != nil {
		logger.Log.Info(err.Error())
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
		logger.Log.Info(err.Error())
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := auth.GetUserID(token.Value)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}

	intBody, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}

	if isValid := luhn.IsValid(intBody); !isValid {
		res.Header().Set("Content-Type", "text/plain")
		logger.Log.Info("Invalid order number: not comply with the Luhn algorithm")
		res.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	orderData := models.OrderData{
		Number:     intBody,
		UserID:     userID,
		Status:     "NEW",
		UploadedAt: time.Now(),
	}
	err = srv.Storage.InsertOrderData(req.Context(), &orderData)
	res.Header().Set("Content-Type", "text/plain")
	if err != nil {
		if errors.Is(err, storage.ErrUserOrderDuplicate) {
			logger.Log.Info(fmt.Sprintf("Order %d has already been placed by the user %s", intBody, userID))
			res.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, storage.ErrOrderDuplicate) {
			logger.Log.Info(fmt.Sprintf("Order %d has already been placed by the user %s", intBody, userID))
			res.WriteHeader(http.StatusConflict)
			return
		}
		logger.Log.Info(err.Error())
		http.Error(res, "Create order error", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

func (srv *Service) GetUserOrders(res http.ResponseWriter, req *http.Request) {
	token, err := req.Cookie("token")

	if err != nil {
		logger.Log.Info("User unauthorized")
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := auth.GetUserID(token.Value)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	if data, err := srv.Storage.SelectOrdersByUserID(req.Context(), userID); data != nil {

		if err != nil {
			logger.Log.Info(err.Error())
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		resp, err := json.Marshal(data)
		if err != nil {
			logger.Log.Info(err.Error())
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		_, err = res.Write(resp)

		if err != nil {
			logger.Log.Info(err.Error())
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

	} else {
		res.WriteHeader(http.StatusNoContent)
	}
}

func (srv *Service) GetUserBalance(res http.ResponseWriter, req *http.Request) {
	token, err := req.Cookie("token")

	if err != nil {
		logger.Log.Info("User unauthorized")
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := auth.GetUserID(token.Value)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := srv.Storage.SelectUserBalance(req.Context(), userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNoBalance) {
			resp, err := json.Marshal(models.GetUserBalanceResponse{
				Current:   0.0,
				Withdrawn: 0.0,
			})
			if err != nil {
				logger.Log.Info(err.Error())
				http.Error(res, err.Error(), http.StatusBadRequest)
			}

			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusOK)
			_, err = res.Write(resp)

			if err != nil {
				logger.Log.Info(err.Error())
				http.Error(res, err.Error(), http.StatusBadRequest)
			}
			return
		}
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	resp, err := json.Marshal(
		models.GetUserBalanceResponse{
			Current:   data.Balance,
			Withdrawn: data.Withdrawn,
		})
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, err = res.Write(resp)

	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
	}
}

func (srv *Service) WithdrawPoints(res http.ResponseWriter, req *http.Request) {
	token, err := req.Cookie("token")

	if err != nil {
		logger.Log.Info("User unauthorized")
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := auth.GetUserID(token.Value)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}

	var jsonBody models.WithdrawnInfo
	if err = json.Unmarshal(body, &jsonBody); err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	intNumber, err := strconv.ParseInt(jsonBody.Number, 10, 64)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}

	if isValid := luhn.IsValid(intNumber); !isValid {
		logger.Log.Info("Invalid order number: not comply with the Luhn algorithm")
		http.Error(res, "Invalid order number", http.StatusUnprocessableEntity)
		return
	}

	data, err := srv.Storage.SelectUserBalance(req.Context(), userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNoBalance) {
			res.WriteHeader(http.StatusOK)
			return
		}
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if data.Balance < jsonBody.Sum {
		http.Error(res, "Insufficient funds", http.StatusPaymentRequired)
		return
	}
	data.Balance = data.Balance - jsonBody.Sum
	data.Withdrawn = data.Withdrawn + jsonBody.Sum

	withdraw := models.Withdraw{
		Number:      intNumber,
		UserID:      userID,
		Sum:         jsonBody.Sum,
		WithdrawnAt: time.Now(),
	}
	err = srv.Storage.UpdateUserBalance(req.Context(), data, &withdraw)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, "User update error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusOK)
}

func (srv *Service) GetUserWithdrawals(res http.ResponseWriter, req *http.Request) {
	token, err := req.Cookie("token")

	if err != nil {
		logger.Log.Info("User unauthorized")
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := auth.GetUserID(token.Value)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	res.Header().Set("Content-Type", "application/json")
	if data, err := srv.Storage.SelectUserWithdrawals(req.Context(), userID); len(data) > 0 {

		if err != nil {
			logger.Log.Info(err.Error())
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		resp, err := json.Marshal(data)
		if err != nil {
			logger.Log.Info(err.Error())
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		res.WriteHeader(http.StatusOK)
		_, err = res.Write(resp)

		if err != nil {
			logger.Log.Info(err.Error())
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

	} else {
		res.WriteHeader(http.StatusOK)
	}
}

func (srv *Service) GetAccrualPoints() {
	ticker := time.NewTicker(2 * time.Second)

	for range ticker.C {
		unprocessedOrders, err := srv.Storage.SelectUnprocessedOrders(context.Background())
		if err != nil {
			logger.Log.Info(err.Error())
			continue
		}
		if len(unprocessedOrders) == 0 {
			continue
		}

		client := resty.New()
		for _, order := range unprocessedOrders {
			strNum := strconv.FormatInt(order.Number, 10)

			resp, err := client.R().Get(srv.Config.AccrualAddr + "/api/orders/" + strNum)
			if err != nil {
				logger.Log.Info(err.Error())
				continue
			}
			if resp.StatusCode() == http.StatusNoContent || resp.StatusCode() == http.StatusTooManyRequests {
				continue
			}

			var accrualData models.AccrualResponse
			err = json.Unmarshal(resp.Body(), &accrualData)
			if err != nil {
				logger.Log.Info(err.Error())
				continue
			}
			if accrualData.Status == "PROCESSING" || accrualData.Status == "REGISTERED" || accrualData.Status == "PROCESSED" || accrualData.Status == "INVALID" {
				order.Accrual = accrualData.Accrual
				order.Status = accrualData.Status

				srv.SaveAccrualPointsCh <- order
			}
		}
	}
}

func (srv *Service) SaveOrdersPoints() {
	ticker := time.NewTicker(5 * time.Second)

	var OrdersForUpdate []models.Orders

	for {
		select {

		case msg := <-srv.SaveAccrualPointsCh:
			OrdersForUpdate = append(OrdersForUpdate, msg)

		case <-ticker.C:
			if len(OrdersForUpdate) == 0 {
				continue
			}

			for _, order := range OrdersForUpdate {
				err := srv.Storage.UpdateOrder(context.Background(), &order)
				if err != nil {
					logger.Log.Info(err.Error())
					continue
				}
			}

			OrdersForUpdate = nil
		}
	}
}
