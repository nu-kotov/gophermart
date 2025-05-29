package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/nu-kotov/gophermart/internal/auth"
	"github.com/nu-kotov/gophermart/internal/logger"
	"github.com/nu-kotov/gophermart/internal/models"
)

func (hnd *Handler) LoginUser(res http.ResponseWriter, req *http.Request) {
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

	userData, err := hnd.Storage.SelectUserData(req.Context(), &jsonBody)
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

	value, err := auth.BuildJWTString(userData.UserID, userData.Login, hnd.Config.TokenExp, hnd.Config.SecretKey)
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
