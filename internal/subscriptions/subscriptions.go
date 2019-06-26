package subscriptions

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
	"github.com/graphql-go/graphql"
)

// The subscription endpoint reacts to a subset of messages from the graphql-over-websocket protocol
// https://github.com/apollographql/subscriptions-transport-ws/blob/master/PROTOCOL.md
// which has been slightly altered - see below.
// GQL_CONNECTION_INIT will be sent to this handler, the returned GQL_CONNECTION_ACK message will contain an object in the payload: {cid: string}
// This result MUST be sent with all subsequent requests to this and the streaming endpoint as a query string 'cid',
// in order to ensure the library will keep working across multiple open browser tabs etc.
// GQL_START, GQL_STOP will be sent to this endpoint,
// and GQL_ERROR will be returned synchronously in case of an error, otherwise a 200 with {"type":"GQL_OK"} will be returned.
// GQL_COMPLETE and GQL_DATA will be sent over the streaming endpoint

type subscriptionStorageAdapter interface {
	NotifyNewSubscription(ctx context.Context, subscriptionID string, subscriberData orchestration.SubscriptionData) error
	NotifyUnsubscribe(ctx context.Context, subscriptionID string, userID string) error
}

type tokenAdapter interface {
	StoreStreamingToken(ctx context.Context, token string, clientID string) error
}

// SubscribeHandler handles the endpoint for processing new subscriptions and contains a reference to the Broker
type SubscribeHandler struct {
	Broker         *orchestration.Broker
	StorageAdapter subscriptionStorageAdapter
	TokenAdapter   tokenAdapter
}

type gqlOverWebsocketProtocol struct {
	Payload interface{}
	Type    string
	ID      string
}

type gqlRequest struct {
	query         string
	operationName string
	variables     map[string]interface{}
}

func decodeGQLQuery(r *http.Request, streamingContext string, operationID string) (*orchestration.SubscriptionData, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var req gqlRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	params := &graphql.Params{
		VariableValues: req.variables,
		RequestString:  req.query,
		OperationName:  req.operationName,
	}
	subData := &orchestration.SubscriptionData{
		SubscriptionID: fmt.Sprintf("%s_%s", streamingContext, operationID),
		GraphQLParams:  params,
	}
	return subData, err
}

func (s *SubscribeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := "123"
	req, err := decodeGQLQuery(r, userID, "foo")
	req.UserID = userID
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Something went wrong"}`))
		return
	}
	s.StorageAdapter.NotifyNewSubscription(r.Context(), req.SubscriptionID, *req)
	fmt.Fprint(w, `{"message":"OK"}`)
}
