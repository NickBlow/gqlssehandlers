package orchestration

import (
	"encoding/json"
	"log"
	"time"

	"github.com/graphql-go/graphql"
)

// SubscriptionData encompasses a particular subscription and the parameters for the GraphQL Query that should be performed.
type SubscriptionData struct {
	SubscriptionID string
	ClientID       string
	GraphQLParams  *graphql.Params
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
	executeSubscriptions        chan SubscriptionData
	bufferedEvents              []interface{}
	clients                     map[string]ClientInfo
	recentlyDisconnectedClients map[string]disconnectedClient
}

// InitializeBroker creates a broker and starts listening to events
func InitializeBroker() *Broker {
	b := &Broker{
		NewClients:                  make(chan ClientInfo),
		ClosedClients:               make(chan string),
		executeSubscriptions:        make(chan SubscriptionData),
		bufferedEvents:              make([]interface{}, 0),
		clients:                     map[string]ClientInfo{},
		recentlyDisconnectedClients: map[string]disconnectedClient{},
	}
	go b.listen()
	return b
}

// NewEventCallback is a function that should be called every time an event happens,
// and contain details of the subscriptions that should be called in response to that event
type NewEventCallback func(subscriptions []SubscriptionData)

// ExecuteQueriesAndPublish triggers the broker to perform the GraphQL Query specified in the SubscriptionData, and propagate it to the clients
func (b *Broker) ExecuteQueriesAndPublish(subscriptions []SubscriptionData) {
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
				result := graphql.Do(*event.GraphQLParams)
				json, err := marshallGQLResult(result)
				if err != nil {
					log.Printf("failed to marshal message: %v", err)
				} else {
					client.CommunicationChannel <- json
				}
			}
		}
	}
}

func marshallGQLResult(result *graphql.Result) (string, error) {
	message, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(message), nil
}
