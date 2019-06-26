// Package gqlssehandlers is a GraphQL Subscriptions over Server Sent Events library for Go.
package gqlssehandlers

import (
	"context"
	"net/http"

	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
	"github.com/NickBlow/gqlssehandlers/internal/streaming"
	"github.com/NickBlow/gqlssehandlers/internal/subscriptions"
)

// SubscriptionAdapter represents an interface that begins listening to a stream of arbitrary events (e.g. a queue or a pub/sub service),
// and calls the callback with an array of interested clients whenever it receives an event.
// The NotifyNewSubscription and NotifyUnsubscribe functions take a context,
// which can be set by middleware on the incoming requests for the subscription handler
type SubscriptionAdapter interface {
	StartListening(cb orchestration.NewEventCallback)
	NotifyNewSubscription(ctx context.Context, subscriptionID string, subscriberData orchestration.SubscriptionData) error
	NotifyUnsubscribe(ctx context.Context, subscriptionID string, userID string) error
}

// TokenAdapter contains methods to store and validate short lived tokens for the streaming endpoint.
// The implementer is in charge of ensuring these tokens expire.
// The StoreStreamingToken and GetClientIDForToken functions take a context,
// which can be set by middleware on the incoming requests for the subscription and streaming handler respectively.
// Please note, the streaming handler can only handle GET requests and cannot set any custom headers, so any context must be
// taken from query strings or cookies.
type TokenAdapter interface {
	StoreStreamingToken(ctx context.Context, token string, clientID string) error
	GetClientIDForToken(ctx context.Context, token string) (string, error)
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
	AuthAdapter      TokenAdapter
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
