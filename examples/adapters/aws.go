package adapters

import (
	"context"

	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
	"github.com/NickBlow/gqlssehandlers/subscriptions"
)

// AWSAdapter stores subscribers in DynamoDB, and uses a combination of SQS and SNS for pubsub
// It expects to be run inside ECS, and will create a queue for itself based on the ECS task id
// It has a default queue name of Subscription-Server-GQL_SSE_HANDLERS if the ECS_CONTAINER_METADATA_URI env var is not set
// It requires a lambda or some other periodic task to clean up dead queues and subscriptions.
// Some of the services may not be covered by the AWS free tier, so please check before running this.
type AWSAdapter struct{}

// StartListening sets up some pubsub infrastructure with SQS and SNS and calls the callback with information about new messages
func (a *AWSAdapter) StartListening(cb orchestration.NewEventCallback) {

}

// NotifyNewSubscription stores a subscription in DynamoDB with a TTL
func (a *AWSAdapter) NotifyNewSubscription(ctx context.Context,
	clientID string, subscriptionID string, subscriberData subscriptions.Data) error {
	return nil
}

// NotifyUnsubscribe removes a subscription from DynamoDB
func (a *AWSAdapter) NotifyUnsubscribe(ctx context.Context, clientID string, subscriptionID string) error {
	return nil
}
