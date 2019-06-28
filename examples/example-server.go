package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/NickBlow/gqlssehandlers"
	"github.com/NickBlow/gqlssehandlers/examples/adapters"
	"github.com/graphql-go/graphql"

	gorrilaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	router := mux.NewRouter()
	eventStream := &adapters.InMemoryAdapter{}

	var fields = graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
			},
		},
	}
	var rootQuery = graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	var schemaConfig = graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	var schema, err = graphql.NewSchema(schemaConfig)
	if err != nil {
		fmt.Println(err)
		panic("Couldn't create schema")
	}

	subscriptionServerConfig := &gqlssehandlers.HandlerConfig{
		Adapter: eventStream,
		Schema:  &schema,
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
