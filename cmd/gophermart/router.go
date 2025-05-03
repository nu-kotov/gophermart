package main

import (
	"github.com/gorilla/mux"
	"github.com/nu-kotov/gophermart/internal/handler"
)

func NewRouter(service handler.Service) *mux.Router {
	router := mux.NewRouter()

	//middlewareStack := middleware.Chain()

	router.HandleFunc(`/api/user/register`, service.RegisterUser).Methods("POST")
	router.HandleFunc(`/api/user/login`, service.LoginUser).Methods("POST")
	router.HandleFunc(`/api/user/orders`, service.CreateOrder).Methods("POST")
	//router.HandleFunc(`/api/user/orders`, service.RegisterUser).Methods("GET")
	//router.HandleFunc(`/api/user/balance`, service.RegisterUser).Methods("GET")
	//router.HandleFunc(`/api/user/balance/withdraw`, service.RegisterUser).Methods("POST")
	//router.HandleFunc(`/api/user/withdrawals`, service.RegisterUser).Methods("GET")

	return router
}
