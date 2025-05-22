package main

import (
	"github.com/gorilla/mux"
	"github.com/nu-kotov/gophermart/internal/handler"
	"github.com/nu-kotov/gophermart/internal/middleware"
)

func NewRouter(service handler.Service) *mux.Router {
	router := mux.NewRouter()

	middlewareStack := middleware.Chain(
		middleware.RequestLogger,
	)

	router.HandleFunc(`/api/user/register`, middlewareStack(service.RegisterUser)).Methods("POST")
	router.HandleFunc(`/api/user/login`, middlewareStack(service.LoginUser)).Methods("POST")
	router.HandleFunc(`/api/user/orders`, middlewareStack(service.CreateOrder)).Methods("POST")
	router.HandleFunc(`/api/user/orders`, middlewareStack(service.GetUserOrders)).Methods("GET")
	router.HandleFunc(`/api/user/balance`, middlewareStack(service.GetUserBalance)).Methods("GET")
	router.HandleFunc(`/api/user/balance/withdraw`, middlewareStack(service.WithdrawPoints)).Methods("POST")
	router.HandleFunc(`/api/user/withdrawals`, middlewareStack(service.GetUserWithdrawals)).Methods("GET")

	return router
}
