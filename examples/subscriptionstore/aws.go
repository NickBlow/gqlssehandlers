package subscriptionstore

import (
	"os"
	"strconv"
	"time"

	"github.com/NickBlow/gqlssehandlers/subscriptions"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var ddbSvc *dynamodb.DynamoDB

func init() {
	sess := getSession()
	ddbSvc = dynamodb.New(sess)
}

// returns a correctly configured AWS session with correct region etc.
// This is a C&P of the one in the event stream, potentially fix this...
func getSession() *session.Session {
	// Set some defaults if env vars not set
	region := os.Getenv("REGION")
	if region == "" {
		region = "eu-west-1"
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region)},
	))
	return sess
}

// StoreSubscriptionInDDB stores the subscription data in DDB with a TTL.
// Table schema should be primary key ClientID with a sort key of SubscriptionID, and the field 'TTL' as the ttl
func StoreSubscriptionInDDB(subscriberData subscriptions.Data, queryData subscriptions.Query, tableName string) error {
	ttl := getSubscriptionTTL()
	variables, err := dynamodbattribute.MarshalMap(queryData.VariableValues)
	if err != nil {
		return err
	}
	_, err = ddbSvc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]*dynamodb.AttributeValue{
			"TTL": {
				N: aws.String(strconv.FormatInt(ttl, 10)),
			},
			"ClientId": {
				S: aws.String(subscriberData.ClientID),
			},
			"SubscriptionID": {
				S: aws.String(subscriberData.SubscriptionID),
			},
			"Variables": {
				M: variables,
			},
			"QueryString": {
				S: aws.String(queryData.RequestString),
			},
		},
	})
	return err
}

// FullSubscriptionData contains the details of the client and subscription including all query data
type FullSubscriptionData struct {
	subscriptions.Data
	subscriptions.Query
}

// GetSubscriptionsFromDDB gets the subscription data from DDB - this is not required by the interface, but on reconnect of a client,
// we should check that we're processing all their active subscriptions.
func GetSubscriptionsFromDDB(ClientID string) []FullSubscriptionData {
	return []FullSubscriptionData{}
}

// RemoveSubscriptionFromDDB removes the subscription data in DDB.
func RemoveSubscriptionFromDDB(subscriberData subscriptions.Data, tableName string) error {
	return nil
}

// TTL is 48 hours
func getSubscriptionTTL() int64 {
	expiresAt := time.Now().Add(time.Hour * 24 * 2)
	return expiresAt.UnixNano() / 1e6
}
