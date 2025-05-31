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

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"github.com/nu-kotov/gophermart/internal/auth"
	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/logger"
	"github.com/nu-kotov/gophermart/internal/middleware"
	"github.com/nu-kotov/gophermart/internal/models"
	"github.com/nu-kotov/gophermart/internal/storage/dberrors"
	"github.com/phedde/luhn-algorithm"
)

type OrdersStorage interface {
	InsertOrderData(context.Context, *models.OrderData) error
	SelectOrdersByUserID(context.Context, string) ([]models.GetUserOrdersResponse, error)
	SelectUnprocessedOrders(ctx context.Context, limit int) ([]models.OrderData, error)
	UpdateOrder(context.Context, *models.OrderData) error
}

type OrdersHandler struct {
	Config              *config.Config
	Storage             OrdersStorage
	SaveAccrualPointsCh chan models.OrderData
	UnprocessedOrdersCh chan models.OrderData
}

func NewOrdersHandler(router *mux.Router, cfg *config.Config, storage OrdersStorage) {

	handler := &OrdersHandler{
		Config:              cfg,
		Storage:             storage,
		SaveAccrualPointsCh: make(chan models.OrderData, 1024),
		UnprocessedOrdersCh: make(chan models.OrderData, 1024),
	}

	middlewareStack := middleware.Chain(
		middleware.RequestLogger,
	)

	router.HandleFunc(`/api/user/orders`, middlewareStack(handler.CreateOrder())).Methods("POST")
	router.HandleFunc(`/api/user/orders`, middlewareStack(handler.GetUserOrders())).Methods("GET")

	go handler.GetAccrualPoints()
	go handler.SaveOrdersPoints()
}

func (handler *OrdersHandler) CreateOrder() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		token, err := req.Cookie("token")

		if err != nil {
			logger.Log.Info(err.Error())
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, err := auth.GetUserID(token.Value, handler.Config.SecretKey)
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
		err = handler.Storage.InsertOrderData(req.Context(), &orderData)
		res.Header().Set("Content-Type", "text/plain")
		if err != nil {
			if errors.Is(err, dberrors.ErrUserOrderDuplicate) {
				logger.Log.Info(fmt.Sprintf("Order %d has already been placed by the user %s", intBody, userID))
				res.WriteHeader(http.StatusOK)
				return
			}
			if errors.Is(err, dberrors.ErrOrderDuplicate) {
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
}

func (handler *OrdersHandler) GetUserOrders() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		token, err := req.Cookie("token")

		if err != nil {
			logger.Log.Info("User unauthorized")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, err := auth.GetUserID(token.Value, handler.Config.SecretKey)
		if err != nil {
			logger.Log.Info(err.Error())
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		if data, err := handler.Storage.SelectOrdersByUserID(req.Context(), userID); data != nil {

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
}

func (handler *OrdersHandler) SaveOrdersPoints() {
	ticker := time.NewTicker(handler.Config.TickerPeriod)

	var OrdersForUpdate []models.OrderData

	for {
		select {

		case msg := <-handler.SaveAccrualPointsCh:
			OrdersForUpdate = append(OrdersForUpdate, msg)

		case <-ticker.C:
			if len(OrdersForUpdate) == 0 {
				continue
			}

			for _, order := range OrdersForUpdate {
				err := handler.Storage.UpdateOrder(context.Background(), &order)
				if err != nil {
					logger.Log.Info(err.Error())
					continue
				}
			}

			OrdersForUpdate = nil
		}
	}
}

func (handler *OrdersHandler) GetAccrualPoints() {
	numWorkers := handler.Config.WorkersNum

	for w := 1; w <= numWorkers; w++ {
		go handler.worker()
	}

	ticker := time.NewTicker(handler.Config.TickerPeriod)

	for range ticker.C {
		unprocessedOrders, err := handler.Storage.SelectUnprocessedOrders(context.Background(), numWorkers)
		if err != nil {
			logger.Log.Info(err.Error())
			continue
		}
		if len(unprocessedOrders) == 0 {
			continue
		}

		for _, order := range unprocessedOrders {
			handler.UnprocessedOrdersCh <- order
		}
	}
}

func (handler *OrdersHandler) worker() {

	client := resty.New()
	for order := range handler.UnprocessedOrdersCh {

		strNum := strconv.FormatInt(order.Number, 10)

		resp, err := client.R().Get(handler.Config.AccrualAddr + "/api/orders/" + strNum)
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

			handler.SaveAccrualPointsCh <- order
		}
	}
}
