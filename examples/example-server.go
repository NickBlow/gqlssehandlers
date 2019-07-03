package main

import (
	"log"
	"net/http"
	"time"

	"github.com/NickBlow/gqlssehandlers"
	"github.com/NickBlow/gqlssehandlers/examples/adapters"
	"github.com/NickBlow/gqlssehandlers/examples/schema"

	gorrilaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	router := mux.NewRouter()
	eventStream := &adapters.InMemoryAdapter{}

	subscriptionServerConfig := &gqlssehandlers.HandlerConfig{
		Adapter: eventStream,
		Schema:  &schema.HelloReactiveSchema,
	}

	handlers := gqlssehandlers.GetHandlers(subscriptionServerConfig)
	router.Handle("/", handlers.PublishStreamHandler).Methods("GET")
	router.Handle("/subscribe", handlers.SubscribeHandler).Methods("POST")

	originsOk := gorrilaHandlers.AllowedOrigins([]string{"https://localhost.wakelet.com"})
	methodsOk := gorrilaHandlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"})

	// Server has long write timeout because we're supporting a streaming response.
	srv := http.Server{
		Addr:         ":8080",
		Handler:      gorrilaHandlers.CORS(originsOk, methodsOk)(router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Minute,
	}

	log.Printf("Server starting on port 8080")
	log.Fatal(srv.ListenAndServe())

}
