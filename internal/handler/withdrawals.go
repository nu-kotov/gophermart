package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/nu-kotov/gophermart/internal/auth"
	"github.com/nu-kotov/gophermart/internal/logger"
	"github.com/nu-kotov/gophermart/internal/models"
	"github.com/nu-kotov/gophermart/internal/storage/dberrors"
	"github.com/phedde/luhn-algorithm"
)

func (hnd *Handler) WithdrawPoints(res http.ResponseWriter, req *http.Request) {
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

	data, err := hnd.Storage.SelectUserBalance(req.Context(), userID)
	if err != nil {
		if errors.Is(err, dberrors.ErrUserNoBalance) {
			res.WriteHeader(http.StatusOK)
			return
		}
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if data.Current < jsonBody.Sum {
		http.Error(res, "Insufficient funds", http.StatusPaymentRequired)
		return
	}
	data.Current = data.Current - jsonBody.Sum
	data.Withdrawn = data.Withdrawn + jsonBody.Sum

	withdraw := models.Withdraw{
		Number:      intNumber,
		UserID:      userID,
		Sum:         jsonBody.Sum,
		WithdrawnAt: time.Now(),
	}
	err = hnd.Storage.UpdateUserBalance(req.Context(), data, &withdraw)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, "User update error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusOK)
}

func (hnd *Handler) GetUserWithdrawals(res http.ResponseWriter, req *http.Request) {
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

	res.Header().Set("Content-Type", "application/json")
	if data, err := hnd.Storage.SelectUserWithdrawals(req.Context(), userID); len(data) > 0 {

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
