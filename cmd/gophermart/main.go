package main

import (
	"log"
	"net/http"

	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/handler"
	"github.com/nu-kotov/gophermart/internal/logger"
	"github.com/nu-kotov/gophermart/internal/storage"
)

func main() {
	if err := logger.NewLogger("info"); err != nil {
		log.Fatal("Error initialize zap logger: ", err)
	}
	config := config.NewConfig()
	store, err := storage.NewConnect(config.DatabaseConnection)
	if err != nil {
		log.Fatal("Error initialize storage: ", err)
	}

	service := handler.NewService(config, store)
	router := NewRouter(*service)

	defer service.Storage.Close()

	log.Fatal(http.ListenAndServe(config.RunAddr, router))
}
