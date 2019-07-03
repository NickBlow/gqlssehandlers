package adapters

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/NickBlow/gqlssehandlers/examples/schema"
	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
	"github.com/NickBlow/gqlssehandlers/subscriptions"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
)

type wrappedSubscriptionData struct {
	subscriptions.Data
	subscriptions.Query
	communicationChannel chan schema.SampleEvent
}

var subscribersMap = make(map[string]wrappedSubscriptionData)

// InMemoryAdapter stores subscribers in memory, and triggers events at random intervals
type InMemoryAdapter struct {
	resultChannel chan subscriptions.WrappedEvent
	mux           sync.Mutex
}

func hasChannelClosedError(errors []gqlerrors.FormattedError) bool {
	for _, val := range errors {
		gqlError, ok := val.OriginalError().(*gqlerrors.Error) // gql library returns errors in a wrapped struct
		if _, isClosedErr := gqlError.OriginalError.(*schema.ChannelClosedError); ok && isClosedErr {
			return true
		}
	}
	return false
}

func (a *InMemoryAdapter) doSubscription(ctx context.Context, data subscriptions.Data, query subscriptions.Query) {
	go func() {
	Loop:
		for {
			res := graphql.Do(graphql.Params{
				Schema:         schema.HelloReactiveSchema,
				Context:        ctx,
				RequestString:  query.RequestString,
				VariableValues: query.VariableValues,
			})
			if hasChannelClosedError(res.Errors) {
				break Loop
			}
			a.resultChannel <- subscriptions.WrappedEvent{
				SubscriptionID: data.SubscriptionID,
				ClientID:       data.ClientID,
				QueryResult:    res,
			}
		}
		fmt.Println("Stopped executing subscription")
	}()
}

// StartListening sends data to each subscriber's source channel at a random interval.
// It must be called before NotifyNewSubscription or NotifyUnsubscribe as it initialises the result channel
func (a *InMemoryAdapter) StartListening(cb orchestration.NewEventCallback) {
	exampleNames := []string{"graphql", "gophers", "world"}
	a.resultChannel = make(chan subscriptions.WrappedEvent)
	go func() {
		for {
			select {
			case <-time.After(time.Second * time.Duration(rand.Intn(10))):
				for _, val := range subscribersMap {
					val.communicationChannel <- schema.SampleEvent{Name: exampleNames[rand.Intn(len(exampleNames))]}
				}
			case event := <-a.resultChannel:
				cb(event)
			}
		}
	}()
}

// cleanUpSubscription removes any existing subscriptions for that clientID/subscriptionID combo
func (a *InMemoryAdapter) cleanUpSubscription(subscriberData subscriptions.Data) {
	compoundKey := fmt.Sprintf("%v_%v", subscriberData.ClientID, subscriberData.SubscriptionID)
	a.mux.Lock() // Just so we don't have concurrent callbacks trying to close the same channel
	if client, ok := subscribersMap[compoundKey]; ok {
		close(client.communicationChannel) // IMPORTANT! This will terminate the goroutine that is executing the subscription
	}
	delete(subscribersMap, compoundKey)
	a.mux.Unlock()
}

// NotifyNewSubscription adds a subscriber to the map
func (a *InMemoryAdapter) NotifyNewSubscription(ctx context.Context, subscriberData subscriptions.Data, queryData subscriptions.Query) error {
	compoundKey := fmt.Sprintf("%v_%v", subscriberData.ClientID, subscriberData.SubscriptionID)
	communicationChannel := make(chan schema.SampleEvent)
	a.cleanUpSubscription(subscriberData) // To stop potential leaks of goroutines
	subscribersMap[compoundKey] = wrappedSubscriptionData{
		Data:                 subscriberData,
		Query:                queryData,
		communicationChannel: communicationChannel,
	}
	subscriptionContext := schema.AddChannelToContext(context.Background(), communicationChannel)
	a.doSubscription(subscriptionContext, subscriberData, queryData)
	return nil
}

// NotifyUnsubscribe removes a subscriber from the map
func (a *InMemoryAdapter) NotifyUnsubscribe(ctx context.Context, subscriberData subscriptions.Data) error {
	a.cleanUpSubscription(subscriberData)
	return nil
}
