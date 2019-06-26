package orchestration

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/graphql-go/graphql"
)

// SubscriptionData encompasses a particular subscription and the parameters for the GraphQL Query that should be performed.
// It will use the schema passed into the handler config.
// RequestString is the raw GraphQL request string
// VariableValues are the values of the variables in the query
// This must be capable of being stored in and retreived from a database
type SubscriptionData struct {
	SubscriptionID string
	ClientID       string
	RequestString  string
	VariableValues map[string]interface{}
}

// SubscriptionWithContext encapsulates a subscription plus any per-request context that the resolvers might need access to.
// RootObject is a mutable map that is passed to all the resolvers (https://github.com/graphql-go/graphql/issues/345)
// Context is per request context
// See https://github.com/graphql-go/graphql for more info on the parameters
type SubscriptionWithContext struct {
	SubscriptionData
	RootObject map[string]interface{}
	Context    context.Context
}

// ClientInfo contains information about a connected client
type ClientInfo struct {
	ClientID             string
	CommunicationChannel chan string
	LastSeenEventID      string
}

// disconnectedClient contains details about a disconnected client, these will be periodically removed
type disconnectedClient struct {
	ClientID       string
	DisconnectedAt int64
}

// Broker contains all the details to manage state of connected clients.
type Broker struct {
	NewClients                  chan ClientInfo
	ClosedClients               chan string
	Schema                      *graphql.Schema
	executeSubscriptions        chan SubscriptionWithContext
	bufferedEvents              []interface{}
	clients                     map[string]ClientInfo
	recentlyDisconnectedClients map[string]disconnectedClient
}

// InitializeBroker creates a broker and starts listening to events
func InitializeBroker(schema *graphql.Schema) *Broker {
	b := &Broker{
		Schema:                      schema,
		NewClients:                  make(chan ClientInfo),
		ClosedClients:               make(chan string),
		executeSubscriptions:        make(chan SubscriptionWithContext),
		bufferedEvents:              make([]interface{}, 0),
		clients:                     map[string]ClientInfo{},
		recentlyDisconnectedClients: map[string]disconnectedClient{},
	}
	go b.listen()
	return b
}

// NewEventCallback is a function that should be called every time an event happens,
// and contain details of the subscriptions that should be called in response to that event
type NewEventCallback func(subscriptions []SubscriptionWithContext)

// ExecuteQueriesAndPublish triggers the broker to perform the GraphQL Query specified in the SubscriptionData, and propagate it to the clients
func (b *Broker) ExecuteQueriesAndPublish(subscriptions []SubscriptionWithContext) {
	for _, val := range subscriptions {
		b.executeSubscriptions <- val
	}
}

func (b *Broker) listen() {
	for {
		select {
		case client := <-b.NewClients:
			b.clients[client.ClientID] = client
		case client := <-b.ClosedClients:
			b.recentlyDisconnectedClients[client] = disconnectedClient{
				ClientID:       client,
				DisconnectedAt: time.Now().Unix(),
			}
			delete(b.clients, client)
		case event := <-b.executeSubscriptions:
			if client, ok := b.clients[event.ClientID]; ok {
				result := graphql.Do(graphql.Params{
					Schema:         *b.Schema,
					RequestString:  event.RequestString,
					VariableValues: event.VariableValues,
					RootObject:     event.RootObject,
					Context:        event.Context,
				})
				json, err := marshallGQLResult(event.SubscriptionID, result)
				if err != nil {
					log.Printf("failed to marshal message: %v", err)
				} else {
					client.CommunicationChannel <- json
				}
			}
		}
	}
}

type gqlData struct {
	Payload *graphql.Result `json:"payload"`
	Type    string          `json:"type"`
	ID      string          `json:"id"`
}

func marshallGQLResult(SubscriptionID string, result *graphql.Result) (string, error) {
	if result == nil {
		return "", errors.New("Nil result")
	}
	response := gqlData{
		Payload: result,
		Type:    "GQL_DATA",
	}
	message, err := json.Marshal(response)
	if err != nil {
		return "", err
	}
	return string(message), nil
}
