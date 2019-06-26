package adapters

import (
	"context"
	"time"

	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
	"github.com/graphql-go/graphql"
)

type subscriptionData = orchestration.SubscriptionData
type subcriptionWithContext = orchestration.SubscriptionWithContext

var subscribersMap = make(map[string]subscriptionData)

// InMemoryAdapter stores subscribers in memory, and triggers events at random intervals
type InMemoryAdapter struct{}

var fields = graphql.Fields{
	"hello": &graphql.Field{
		Type: graphql.String,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return "world", nil
		},
	},
}
var rootQuery = graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
var schemaConfig = graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
var schema, err = graphql.NewSchema(schemaConfig)

//StartListening calls the callback at random intervals
func (a *InMemoryAdapter) StartListening(cb orchestration.NewEventCallback) {
	go func() {
		for {
			<-time.After(time.Second * 10)
			for _, val := range subscribersMap {
				valWithContext := subcriptionWithContext{SubscriptionData: val}
				cb([]subcriptionWithContext{valWithContext})
			}
		}
	}()
}

// NotifyNewSubscription adds a subscriber to the map
func (a *InMemoryAdapter) NotifyNewSubscription(ctx context.Context, clientID string, subscriptionID string, subscriberData orchestration.SubscriptionData) error {
	subscribersMap[subscriptionID] = subscriberData
	return nil
}

// NotifyUnsubscribe removes a subscriber from the map
func (a *InMemoryAdapter) NotifyUnsubscribe(ctx context.Context, clientID string, subscriptionID string) error {
	delete(subscribersMap, subscriptionID)
	return nil
}

// GetSubscriptionData returns data on a subscription by ID
func (a *InMemoryAdapter) GetSubscriptionData(subscriptionID string) (orchestration.SubscriptionData, error) {
	return subscribersMap[subscriptionID], nil
}
