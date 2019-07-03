package adapters

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/NickBlow/gqlssehandlers/examples/schema"
	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
	"github.com/NickBlow/gqlssehandlers/subscriptions"
	"github.com/graphql-go/graphql"
)

type subscriptionData = subscriptions.Data

var subscribersMap = make(map[string]subscriptionData)

// InMemoryAdapter stores subscribers in memory, and triggers events at random intervals
type InMemoryAdapter struct{}

//StartListening calls the callback at random intervals
func (a *InMemoryAdapter) StartListening(cb orchestration.NewEventCallback) {
	go func() {
		for {
			<-time.After(time.Second * time.Duration(rand.Intn(10)))
			for _, val := range subscribersMap {
				cb(subscriptions.WrappedEvent{
					SubscriptionID: val.SubscriptionID,
					ClientID:       val.ClientID,
					QueryResult: graphql.Do(
						graphql.Params{
							Schema:         schema.HelloWorldSchema,
							RequestString:  val.RequestString,
							VariableValues: val.VariableValues,
						},
					),
				})
			}
		}
	}()
}

// NotifyNewSubscription adds a subscriber to the map
func (a *InMemoryAdapter) NotifyNewSubscription(ctx context.Context, clientID string, subscriptionID string, subscriberData subscriptionData) error {
	compoundKey := fmt.Sprintf("%v_%v", clientID, subscriptionID)
	subscribersMap[compoundKey] = subscriberData
	return nil
}

// NotifyUnsubscribe removes a subscriber from the map
func (a *InMemoryAdapter) NotifyUnsubscribe(ctx context.Context, clientID string, subscriptionID string) error {
	compoundKey := fmt.Sprintf("%v_%v", clientID, subscriptionID)
	delete(subscribersMap, compoundKey)
	return nil
}
