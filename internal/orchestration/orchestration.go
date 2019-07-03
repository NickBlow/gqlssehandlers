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
	CommunicationChannel chan []byte
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
func InitializeBroker(schema *graphql.Schema, newClientCb func(string) error, clientDisconnectCb func(string) error) *Broker {
	b := &Broker{
		Schema:         schema,
		NewClients:     make(chan ClientInfo),
		ClosedClients:  make(chan string),
		ClosingClients: make(chan string),
		newEvents:      make(chan subscriptions.WrappedEvent),
		bufferedEvents: make([]interface{}, 0),
		clients:        map[string]ClientInfo{},
	}
	go b.listen(newClientCb, clientDisconnectCb)
	return b
}

// PushDataToClient sends the event payload to the specified clients
func (b *Broker) PushDataToClient(event subscriptions.WrappedEvent) error {
	b.newEvents <- event
	return nil
}

func (b *Broker) listen(newClientCb func(string) error, clientDisconnectCb func(string) error) {
	for {
		select {
		case client := <-b.NewClients:
			b.clients[client.ClientID] = client
			newClientCb(client.ClientID)
		case client := <-b.ClosingClients:
			b.clients[client].CloseChannel <- true
		case client := <-b.ClosedClients:
			delete(b.clients, client)
			clientDisconnectCb(client)
		case event := <-b.newEvents:
			client := b.clients[event.ClientID]
			if client.CommunicationChannel == nil {
				break
			}
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
