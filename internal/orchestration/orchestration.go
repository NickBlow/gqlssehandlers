package orchestration

import "github.com/graphql-go/graphql"

// SubscriptionData encompasses a particular subscription and the parameters for the GraphQL Query that should be performed.
type SubscriptionData struct {
	ID            string
	GraphQLParams *graphql.Params
}

// ClientInfo contains information about a connected client
type ClientInfo struct {
	ClientID             string
	CommunicationChannel chan string
	LastSeenEventID      string
	subscriptions        []string
}

var subscriptionLookupTable = make(map[string]ClientInfo)

// Broker contains all the details to manage state of connected clients. In order to start working, you must call StartListening
type Broker struct {
	NewClients           chan ClientInfo
	ClosedClients        chan string
	NewSubscriptions     chan SubscriptionData
	executeSubscriptions chan SubscriptionData
	bufferedEvents       []interface{}
	clients              map[string]ClientInfo
}

// InitializeBroker creates a broker and starts listening to events
func InitializeBroker() *Broker {
	b := &Broker{
		NewClients:           make(chan ClientInfo),
		ClosedClients:        make(chan string),
		NewSubscriptions:     make(chan SubscriptionData),
		executeSubscriptions: make(chan SubscriptionData),
		bufferedEvents:       make([]interface{}, 0),
		clients:              map[string]ClientInfo{},
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
			delete(b.clients, client)
		case event := <-b.executeSubscriptions:
			for _, client := range b.clients {
				client.CommunicationChannel <- event.ID
			}
		}
	}
}
