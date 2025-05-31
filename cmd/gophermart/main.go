package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/handler"
	"github.com/nu-kotov/gophermart/internal/logger"
	"github.com/nu-kotov/gophermart/internal/storage"
)

func main() {
	if err := logger.NewLogger("info"); err != nil {
		log.Fatal("Error initialize zap logger: ", err)
	}

	config, err := config.NewConfig()
	if err != nil {
		log.Fatal("Error initialize config: ", err)
	}

	pgStor, err := storage.NewPgStorage(config)
	if err != nil {
		log.Fatal("Error pg connection: ", err)
	}

	balanceStorage := storage.NewBalanceStorage(pgStor)
	if err != nil {
		log.Fatal("Error initialize BalanceStorage: ", err)
	}
	ordersStorage := storage.NewOrdersStorage(pgStor)
	if err != nil {
		log.Fatal("Error initialize OrdersStorage: ", err)
	}
	usersStorage := storage.NewUsersStorage(pgStor)
	if err != nil {
		log.Fatal("Error initialize UsersStorage: ", err)
	}
	withdrawalsStorage := storage.NewWithdrawalsStorage(pgStor)
	if err != nil {
		log.Fatal("Error initialize WithdrawalsStorage: ", err)
	}

	router := mux.NewRouter()

	handler.NewBalancesHandler(router, config, balanceStorage)
	handler.NewOrdersHandler(router, config, ordersStorage)
	handler.NewUsersHandler(router, config, usersStorage)
	handler.NewWithdrawalsHandler(router, config, withdrawalsStorage)

	defer pgStor.Close()

	log.Fatal(http.ListenAndServe(config.RunAddr, router))
}
