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
	config, err := config.NewConfig()
	if err != nil {
		log.Fatal("Error initialize config: ", err)
	}

	store, err := storage.NewStorage(config)
	if err != nil {
		log.Fatal("Error initialize storage: ", err)
	}

	handler := handler.NewHandler(config, store)
	router := NewRouter(*handler)

	defer handler.Storage.Close()

	log.Fatal(http.ListenAndServe(config.RunAddr, router))
}
