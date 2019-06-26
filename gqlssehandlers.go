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
// The NotifyNewSubscription and NotifyUnsubscribe functions take the request context,
// which can be set by middleware on the incoming requests for the subscription handler.
type SubscriptionAdapter interface {
	StartListening(cb orchestration.NewEventCallback)
	NotifyNewSubscription(ctx context.Context, subscriptionID string, subscriberData orchestration.SubscriptionData) error
	NotifyUnsubscribe(ctx context.Context, subscriptionID string, userID string) error
}

// Handlers is a struct containing the generated handlers.
type Handlers struct {
	SubscribeHandler     http.Handler
	PublishStreamHandler http.Handler
}

// HandlerConfig represesents the configuration options for GQLSSEHandlers
// If the EventBuffer size has been set, the server will store events in memory,
// and re-send them if a client reconnects with a Last-Event-ID Header. This may result in duplicate messages being sent,
// so ensure your client is idempotent.
// This will also only send events that the server itself received.
// If you want to ensure a client always gets buffered messages, you can either use sticky sessions,
// route based on some hash, or multicast events to all servers.
type HandlerConfig struct {
	Adapter         SubscriptionAdapter
	EventBufferSize int64
}

// GetHandlers returns all the handlers required to set up the GraphQL subscription.
// The handlers have a concept of client ID, and will by default set a cookie with a client id and use that.
// This default is not safe across multiple browser windows/tabs,
// and while the id uses a strong random number generator, it is not signed.
// If you set a 'GQLSSE_CLIENT_ID' value on the request context it will pick it up and use this instead.
func GetHandlers(config *HandlerConfig) *Handlers {
	subscriptionBroker := orchestration.InitializeBroker()
	config.Adapter.StartListening(subscriptionBroker.ExecuteQueriesAndPublish)

	return &Handlers{
		SubscribeHandler: &subscriptions.SubscribeHandler{
			Broker:         subscriptionBroker,
			StorageAdapter: config.Adapter,
		},
		PublishStreamHandler: &streaming.Handler{
			Broker: subscriptionBroker,
		},
	}
}
