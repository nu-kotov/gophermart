package main

import (
	"log"
	"net/http"

	"github.com/nu-kotov/gophermart/internal/config"
	"github.com/nu-kotov/gophermart/internal/handler"
	"github.com/nu-kotov/gophermart/internal/storage"
)

func main() {

	config := config.NewConfig()
	store, err := storage.NewConnect(config.DatabaseConnection)
	if err != nil {
		log.Fatal("Error initialize storage: ", err)
	}

	service := handler.NewService(config, store)
	router := NewRouter(*service)

	//defer service.Storage.Close()

	log.Fatal(http.ListenAndServe(config.RunAddr, router))
}
