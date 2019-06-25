package gqlssehandlers

import (
	"net/http"

	"github.com/NickBlow/gqlssehandlers/auth"
	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
	"github.com/NickBlow/gqlssehandlers/internal/streaming"
	"github.com/NickBlow/gqlssehandlers/internal/subscriptions"
)

// SubscriptionAdapter represents an interface that begins listening to a stream of arbitrary events (e.g. a queue or a pub/sub service),
// and calls the callback with an array of interested clients whenever it receives an event.
type SubscriptionAdapter interface {
	StartListening(cb orchestration.NewEventCallback)
	NotifyNewSubscription(subscriber orchestration.SubscriptionData) error
	NotifyUnsubscribe(subscriptionID string, userID string) error
}

// AuthAdapter provides the methods required for authentication and authorization
// It also contains methods to store and validate short lived tokens for the streaming endpoint.
// The implementer is in charge of ensuring these tokens expire.
type AuthAdapter interface {
	StoreStreamingToken(token string, userID string) error
	GetUserIDForToken(token string) (string, error)
	GetUserIDFromRequest(req *http.Request) (string, *auth.FailedResponse)
}

// Handlers is a struct containing the generated handlers.
type Handlers struct {
	SubscribeHandler     http.Handler
	PublishStreamHandler http.Handler
}

// HandlerConfig represesents the configuration options for GQLSSEHandlers
// If CookieSigningKey is set, it will sign the session cookie it sets for the streaming endpoint with this key.
// Regardless of whether this is set, clients should handle the streaming endpoint sending them 401s,
// and refresh the streaming token (key may have rotated, token may have expired etc).
// If the EventBuffer size has been set, the server will store events in memory,
// and re-send them if a client reconnects with a Last-Event-ID Header. This may result in duplicate messages being sent,
// so ensure your client is idempotent.
// This will also only send events that the server itself received.
// If you want to ensure a client always gets buffered messages, you can either use sticky sessions,
// route based on some hash, or multicast events to all servers.
type HandlerConfig struct {
	Adapter          SubscriptionAdapter
	AuthAdapter      AuthAdapter
	EventBufferSize  int64
	CookieSigningKey string
}

// GetHandlers returns all the handlers required to set up the GraphQL subscription
func GetHandlers(config *HandlerConfig) *Handlers {
	subscriptionBroker := orchestration.InitializeBroker()
	config.Adapter.StartListening(subscriptionBroker.ExecuteQueriesAndPublish)

	return &Handlers{
		SubscribeHandler: &subscriptions.SubscribeHandler{
			Broker:         subscriptionBroker,
			StorageAdapter: config.Adapter,
			AuthAdapter:    config.AuthAdapter,
		},
		PublishStreamHandler: &streaming.Handler{
			Broker:      subscriptionBroker,
			AuthAdapter: config.AuthAdapter,
		},
	}
}
