package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/nu-kotov/gophermart/internal/auth"
	"github.com/nu-kotov/gophermart/internal/logger"
	"github.com/nu-kotov/gophermart/internal/models"
)

func (hnd *Handler) RegisterUser(res http.ResponseWriter, req *http.Request) {

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

	err = hnd.Storage.InsertUserData(req.Context(), &jsonBody)
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(res, "Register user error", http.StatusInternalServerError)
		return
	}

	value, err := auth.BuildJWTString(jsonBody.UserID, jsonBody.Login, hnd.Config.TokenExp, hnd.Config.SecretKey)
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
