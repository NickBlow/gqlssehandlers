// Package gqlssehandlers is a GraphQL Subscriptions over Server Sent Events library for Go.
// It will create two handler endpoints that react to a subset of messages from the graphql-over-websocket protocol
// (see subscriptions package)
package gqlssehandlers

import (
	"context"
	"net/http"

	"github.com/NickBlow/gqlssehandlers/clientid"
	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
	"github.com/NickBlow/gqlssehandlers/internal/streaming"
	"github.com/NickBlow/gqlssehandlers/internal/subscriptionhandlers"
	"github.com/NickBlow/gqlssehandlers/subscriptions"
	"github.com/graphql-go/graphql"
)

// SubscriptionAdapter represents an interface that begins listening to a stream of arbitrary events (e.g. a queue or a pub/sub service),
// and calls the callback with an array of interested clients whenever it receives an event.
// The NotifyNewSubscription and NotifyUnsubscribe functions take the request context,
// which can be set by middleware on the incoming requests for the subscription handler.
// SubscriptionIDs are set by the user, and are therefore *NOT* unique.
// The combination of clientID and subscriptionID is unique
// ClientID and subscriptions SHOULD be created with a TTL to clean them up after a while, and clients SHOULD be aware of this TTL,
// so they can automatically recreate subscriptions they've created and have expired.
type SubscriptionAdapter interface {
	StartListening(cb orchestration.NewEventCallback)
	NotifyNewSubscription(ctx context.Context, clientID string, subscriptionID string, subscriberData subscriptions.Data) error
	NotifyUnsubscribe(ctx context.Context, clientID string, subscriptionID string) error
}

// Handlers is a struct containing the generated handlers.
type Handlers struct {
	SubscribeHandler     http.Handler
	PublishStreamHandler http.Handler
}

// HandlerConfig represesents the configuration options for GQLSSEHandlers
// You should pass in your graphql Schema here
type HandlerConfig struct {
	Adapter SubscriptionAdapter
	Schema  *graphql.Schema
}

// GetHandlers returns all the handlers required to set up the GraphQL subscription.
// The handlers have a concept of client ID, and will by default set a cookie with a client id and use that.
// This default is not safe across multiple browser windows/tabs,
// and while the id uses a strong random number generator, it is not signed.
// You can write middleware to set the ClientIDKey in the context to overwrite this default behaviour
// See the clientid package for more information.
func GetHandlers(config *HandlerConfig) *Handlers {
	subscriptionBroker := orchestration.InitializeBroker(config.Schema)
	config.Adapter.StartListening(subscriptionBroker.PushDataToClient)

	subscribeHandler := &subscriptionhandlers.Handler{
		Broker:         subscriptionBroker,
		StorageAdapter: config.Adapter,
	}

	publishStreamHandler := &streaming.Handler{
		Broker: subscriptionBroker,
	}
	return &Handlers{
		SubscribeHandler:     clientid.Middleware(subscribeHandler),
		PublishStreamHandler: clientid.Middleware(publishStreamHandler),
	}
}
