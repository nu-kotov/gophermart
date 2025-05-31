package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/handler"
	"github.com/nu-kotov/gophermart/internal/logger"
	"github.com/nu-kotov/gophermart/internal/storage"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := logger.NewLogger("info"); err != nil {
		logger.Log.Info(fmt.Sprintf("Error initialize zap logger: %s", err.Error()))
		return err
	}

	config, err := config.NewConfig()
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Error initialize config: %s", err.Error()))
		return err
	}

	pgStor, err := storage.NewPgStorage(config)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Error pg connection: %s", err.Error()))
		return err
	}

	balanceStorage := storage.NewBalanceStorage(pgStor)
	ordersStorage := storage.NewOrdersStorage(pgStor)
	usersStorage := storage.NewUsersStorage(pgStor)
	withdrawalsStorage := storage.NewWithdrawalsStorage(pgStor)

	router := mux.NewRouter()

	handler.NewBalancesHandler(router, config, balanceStorage)
	handler.NewOrdersHandler(router, config, ordersStorage)
	handler.NewUsersHandler(router, config, usersStorage)
	handler.NewWithdrawalsHandler(router, config, withdrawalsStorage)

	defer pgStor.Close()

	err = http.ListenAndServe(config.RunAddr, router)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Error starting server: %s", err.Error()))
		return err
	}
	return nil
}
