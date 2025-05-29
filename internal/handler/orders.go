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

	"github.com/nu-kotov/gophermart/internal/auth"
	"github.com/nu-kotov/gophermart/internal/logger"
	"github.com/nu-kotov/gophermart/internal/models"
	"github.com/nu-kotov/gophermart/internal/storage/db_errors"
	"github.com/phedde/luhn-algorithm"
)

func (hnd *Handler) CreateOrder(res http.ResponseWriter, req *http.Request) {

	token, err := req.Cookie("token")

	if err != nil {
		logger.Log.Info(err.Error())
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := auth.GetUserID(token.Value, hnd.Config.SecretKey)
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
	err = hnd.Storage.InsertOrderData(req.Context(), &orderData)
	res.Header().Set("Content-Type", "text/plain")
	if err != nil {
		if errors.Is(err, db_errors.ErrUserOrderDuplicate) {
			logger.Log.Info(fmt.Sprintf("Order %d has already been placed by the user %s", intBody, userID))
			res.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, db_errors.ErrOrderDuplicate) {
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

func (hnd *Handler) GetUserOrders(res http.ResponseWriter, req *http.Request) {
	token, err := req.Cookie("token")

	if err != nil {
		logger.Log.Info("User unauthorized")
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := auth.GetUserID(token.Value, hnd.Config.SecretKey)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	if data, err := hnd.Storage.SelectOrdersByUserID(req.Context(), userID); data != nil {

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

func (hnd *Handler) SaveOrdersPoints() {
	ticker := time.NewTicker(1 * time.Second)

	var OrdersForUpdate []models.OrderData

	for {
		select {

		case msg := <-hnd.SaveAccrualPointsCh:
			OrdersForUpdate = append(OrdersForUpdate, msg)

		case <-ticker.C:
			if len(OrdersForUpdate) == 0 {
				continue
			}

			for _, order := range OrdersForUpdate {
				err := hnd.Storage.UpdateOrder(context.Background(), &order)
				if err != nil {
					logger.Log.Info(err.Error())
					continue
				}
			}

			OrdersForUpdate = nil
		}
	}
}
