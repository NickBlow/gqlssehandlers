package callbacks

import "github.com/NickBlow/gqlssehandlers/subscriptions"

// NewEventCallback is a function that should be called every time an event happens.
// The event should be the results of the graphql query, and finished should be a boolean
// detailing whether more events of this type should be expected
// It will return an error if something went wrong when executing the callback
type NewEventCallback func(subscriptions.WrappedEvent) error
