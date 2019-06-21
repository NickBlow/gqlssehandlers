package subscriptions

import (
	"fmt"
	"net/http"

	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
)

// SubscribeHandler handles the endpoint for processing new subscriptions and contains a reference to the Broker
type SubscribeHandler struct {
	Broker *orchestration.Broker
}

func (s *SubscribeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello World")
}

// UnsubscribeHandler handles the endpoint for deleting subscriptions and contains a reference to the Broker
type UnsubscribeHandler struct {
	Broker *orchestration.Broker
}

func (s *UnsubscribeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello World")
}
