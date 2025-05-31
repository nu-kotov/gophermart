package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nu-kotov/gophermart/internal/auth"
	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/logger"
	"github.com/nu-kotov/gophermart/internal/middleware"
	"github.com/nu-kotov/gophermart/internal/models"
)

type WithdrawalsStorage interface {
	SelectUserWithdrawals(context.Context, string) ([]models.WithdrawnInfo, error)
}

type WithdrawalsHandler struct {
	Config  *config.Config
	Storage WithdrawalsStorage
}

func NewWithdrawalsHandler(router *mux.Router, cfg *config.Config, storage WithdrawalsStorage) {

	handler := &WithdrawalsHandler{
		Config:  cfg,
		Storage: storage,
	}

	middlewareStack := middleware.Chain(
		middleware.RequestLogger,
	)

	router.HandleFunc(`/api/user/withdrawals`, middlewareStack(handler.GetUserWithdrawals())).Methods("GET")
}

func (handler *WithdrawalsHandler) GetUserWithdrawals() http.HandlerFunc {
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

		res.Header().Set("Content-Type", "application/json")
		if data, err := handler.Storage.SelectUserWithdrawals(req.Context(), userID); len(data) > 0 {

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
}
