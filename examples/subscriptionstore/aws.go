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

// DDBStore wraps a DynamoDB database
type DDBStore struct {
	TableName string
}

// StoreSubscriptionInDDB stores the subscription data in DDB with a TTL.
// Table schema should be primary key ClientID with a sort key of SubscriptionID, and the field 'TTL' as the ttl
func (d *DDBStore) StoreSubscriptionInDDB(subscriberData subscriptions.Data, queryData subscriptions.Query) error {
	ttl := getSubscriptionTTL()
	variables, err := dynamodbattribute.MarshalMap(queryData.VariableValues)
	if err != nil {
		return err
	}
	_, err = ddbSvc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(d.TableName),
		Item: map[string]*dynamodb.AttributeValue{
			"TTL": {
				N: aws.String(strconv.FormatInt(ttl, 10)),
			},
			"ClientID": {
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
func (d *DDBStore) GetSubscriptionsFromDDB(ClientID string) []FullSubscriptionData {
	return []FullSubscriptionData{}
}

// RemoveSubscriptionFromDDB removes the subscription data in DDB.
func (d *DDBStore) RemoveSubscriptionFromDDB(subscriberData subscriptions.Data) error {
	ddbSvc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(d.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ClientID": {
				S: aws.String(subscriberData.ClientID),
			},
			"SubscriptionID": {
				S: aws.String(subscriberData.SubscriptionID),
			},
		},
	})
	return nil
}

// TTL is 48 hours
func getSubscriptionTTL() int64 {
	expiresAt := time.Now().Add(time.Hour * 24 * 2)
	return expiresAt.UnixNano() / 1e6
}
