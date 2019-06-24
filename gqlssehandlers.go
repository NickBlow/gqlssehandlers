package gqlssehandlers

import (
	"net/http"

	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
	"github.com/NickBlow/gqlssehandlers/internal/streaming"
	"github.com/NickBlow/gqlssehandlers/internal/subscriptions"
)

// OTP represents the data for a one time password for the streaming endpoint.
type OTP struct {
	ID       string
	ClientID *string
}

// SubscriptionAdapter represents an interface that begins listening to a stream of arbitrary events (e.g. a queue or a pub/sub service),
// and calls the callback with an array of interested clients whenever it receives an event.
type SubscriptionAdapter interface {
	StartListening(cb orchestration.NewEventCallback)
	NotifyNewSubscription(subscriber *orchestration.SubscriptionData) error
	NotifyUnsubscribe(subscriptionID string) error
}

// Handlers is a struct containing the generated handlers.
type Handlers struct {
	SubscribeHandler     http.Handler
	UnsubscribeHandler   http.Handler
	PublishStreamHandler http.Handler
}

// HandlerConfig represesents the configuration options for GQLSSEHandlers
// If the EventBuffer size has been set, the server will store events in memory,
// and re-send them if a client reconnects with a Last-Event-ID Header. This may result in duplicate messages being sent,
// so ensure your client is idempotent.
// This will also only send events that the server itself received.
// If you want to ensure a client always gets buffered messages, you can either use sticky sessions, route based on some hash, or multicast events to all servers.
type HandlerConfig struct {
	Adapter         SubscriptionAdapter
	EventBufferSize int64
}

// GetHandlers returns all the handlers required to set up the GraphQL subscription
func GetHandlers(config *HandlerConfig) *Handlers {
	subscriptionBroker := orchestration.InitializeBroker()
	config.Adapter.StartListening(subscriptionBroker.ExecuteQueriesAndPublish)

	return &Handlers{
		SubscribeHandler:     &subscriptions.SubscribeHandler{Broker: subscriptionBroker, StorageAdapter: config.Adapter},
		UnsubscribeHandler:   &subscriptions.UnsubscribeHandler{Broker: subscriptionBroker, StorageAdapter: config.Adapter},
		PublishStreamHandler: &streaming.Handler{Broker: subscriptionBroker},
	}
}
