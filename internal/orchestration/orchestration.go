package orchestration

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/NickBlow/gqlssehandlers/protocol"
	"github.com/NickBlow/gqlssehandlers/subscriptions"
	"github.com/graphql-go/graphql"
)

// ClientInfo contains information about a connected client
type ClientInfo struct {
	ClientID             string
	CommunicationChannel chan interface{}
	LastSeenEventID      string
	CloseChannel         chan bool
}

// Broker contains all the details to manage state of connected clients.
type Broker struct {
	NewClients     chan ClientInfo
	ClosedClients  chan string
	ClosingClients chan string
	Schema         *graphql.Schema
	newEvents      chan subscriptions.WrappedEvent
	bufferedEvents []interface{} // TODO implement
	clients        map[string]ClientInfo
}

// InitializeBroker creates a broker and starts listening to events
func InitializeBroker(schema *graphql.Schema) *Broker {
	b := &Broker{
		Schema:         schema,
		NewClients:     make(chan ClientInfo),
		ClosedClients:  make(chan string),
		ClosingClients: make(chan string),
		newEvents:      make(chan subscriptions.WrappedEvent),
		bufferedEvents: make([]interface{}, 0),
		clients:        map[string]ClientInfo{},
	}
	go b.listen()
	return b
}

// NewEventCallback is a function that should be called every time an event happens.
// The event should be the results of the graphql query, and finished should be a boolean
// detailing whether more events of this type should be expected
type NewEventCallback func(subscriptions.WrappedEvent)

// PushDataToClient sends the event payload to the specified clients
func (b *Broker) PushDataToClient(event subscriptions.WrappedEvent) {
	b.newEvents <- event
}

func (b *Broker) listen() {
	for {
		select {
		case client := <-b.NewClients:
			b.clients[client.ClientID] = client
		case client := <-b.ClosingClients:
			b.clients[client].CloseChannel <- true
		case client := <-b.ClosedClients:
			delete(b.clients, client)
		case event := <-b.newEvents:
			client := b.clients[event.ClientID]
			resultType := protocol.GQLData
			if event.Finished {
				resultType = protocol.GQLComplete
			}
			data, err := json.Marshal(
				&protocol.GQLOverWebsocketProtocol{
					Type:    resultType,
					Payload: &protocol.PayloadBytes{Value: event.QueryResult},
					ID:      event.SubscriptionID,
				},
			)
			if err != nil {
				fmt.Println(err)
				fmt.Println("Could not marshall data")
			} else {
				client.CommunicationChannel <- data
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
