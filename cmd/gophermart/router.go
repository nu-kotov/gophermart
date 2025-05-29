package main

import (
	"github.com/gorilla/mux"
	"github.com/nu-kotov/gophermart/internal/handler"
	"github.com/nu-kotov/gophermart/internal/middleware"
)

func NewRouter(handler handler.Handler) *mux.Router {
	router := mux.NewRouter()

	middlewareStack := middleware.Chain(
		middleware.RequestLogger,
	)

	router.HandleFunc(`/api/user/register`, middlewareStack(handler.RegisterUser)).Methods("POST")
	router.HandleFunc(`/api/user/login`, middlewareStack(handler.LoginUser)).Methods("POST")
	router.HandleFunc(`/api/user/orders`, middlewareStack(handler.CreateOrder)).Methods("POST")
	router.HandleFunc(`/api/user/orders`, middlewareStack(handler.GetUserOrders)).Methods("GET")
	router.HandleFunc(`/api/user/balance`, middlewareStack(handler.GetUserBalance)).Methods("GET")
	router.HandleFunc(`/api/user/balance/withdraw`, middlewareStack(handler.WithdrawPoints)).Methods("POST")
	router.HandleFunc(`/api/user/withdrawals`, middlewareStack(handler.GetUserWithdrawals)).Methods("GET")

	return router
}
