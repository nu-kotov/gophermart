package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nu-kotov/gophermart/internal/auth"
	"github.com/nu-kotov/gophermart/internal/logger"
	"github.com/nu-kotov/gophermart/internal/models"
	"github.com/nu-kotov/gophermart/internal/storage/db_errors"
)

func (hnd *Handler) GetUserBalance(res http.ResponseWriter, req *http.Request) {
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
		return
	}

	data, err := hnd.Storage.SelectUserBalance(req.Context(), userID)
	if err != nil {
		if errors.Is(err, db_errors.ErrUserNoBalance) {
			resp, err := json.Marshal(models.UserBalance{
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
}
