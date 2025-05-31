package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/nu-kotov/gophermart/internal/auth"

	"github.com/gorilla/mux"
	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/logger"
	"github.com/nu-kotov/gophermart/internal/middleware"
	"github.com/nu-kotov/gophermart/internal/models"
)

type UsersStorage interface {
	InsertUserData(context.Context, *models.UserData) error
	SelectUserData(context.Context, *models.UserData) (*models.UserData, error)
}

type UsersHandler struct {
	Config  *config.Config
	Storage UsersStorage
}

func NewUsersHandler(router *mux.Router, cfg *config.Config, storage UsersStorage) {

	handler := &UsersHandler{
		Config:  cfg,
		Storage: storage,
	}

	middlewareStack := middleware.Chain(
		middleware.RequestLogger,
	)

	router.HandleFunc(`/api/user/register`, middlewareStack(handler.RegisterUser())).Methods("POST")
	router.HandleFunc(`/api/user/login`, middlewareStack(handler.LoginUser())).Methods("POST")

}

func (handler *UsersHandler) RegisterUser() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
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

		err = handler.Storage.InsertUserData(req.Context(), &jsonBody)
		if err != nil {
			logger.Log.Info(err.Error())
			http.Error(res, "Register user error", http.StatusInternalServerError)
			return
		}

		value, err := auth.BuildJWTString(
			jsonBody.UserID,
			jsonBody.Login,
			handler.Config.TokenExp,
			handler.Config.SecretKey,
		)
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
}

func (handler *UsersHandler) LoginUser() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
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

		userData, err := handler.Storage.SelectUserData(req.Context(), &jsonBody)
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

		value, err := auth.BuildJWTString(
			userData.UserID,
			userData.Login,
			handler.Config.TokenExp,
			handler.Config.SecretKey,
		)
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
}
