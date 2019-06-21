package adapters

import (
	"time"

	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
)

type subscriptionData = orchestration.SubscriptionData

var subscribersMap = make(map[string]subscriptionData)

// InMemoryAdapter stores subscribers in memory, and triggers events at random intervals
type InMemoryAdapter struct{}

//StartListening calls the callback at random intervals
func (a *InMemoryAdapter) StartListening(cb orchestration.NewEventCallback) {

	go func() {
		for {
			<-time.After(time.Second * 10)
			cb([]subscriptionData{subscriptionData{ID: "Hello"}})
		}
	}()

}

// NotifyNewSubscription adds a subscriber to the map
func (a *InMemoryAdapter) NotifyNewSubscription(subscriber subscriptionData) error {
	subscribersMap[subscriber.ID] = subscriber
	return nil
}

// NotifyUnsubscribe removes a subscriber from the map
func (a *InMemoryAdapter) NotifyUnsubscribe(subscriber string) error {
	delete(subscribersMap, subscriber)
	return nil
}
