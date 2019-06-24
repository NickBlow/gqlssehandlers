package subscriptions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
	"github.com/graphql-go/graphql"
	gonanoid "github.com/matoous/go-nanoid"
)

type subscriptionStorageAdapter interface {
	NotifyNewSubscription(subscriber *orchestration.SubscriptionData) error
	NotifyUnsubscribe(subscriptionID string) error
}

// SubscribeHandler handles the endpoint for processing new subscriptions and contains a reference to the Broker
type SubscribeHandler struct {
	Broker         *orchestration.Broker
	StorageAdapter subscriptionStorageAdapter
}

type gqlRequest struct {
	query         string
	operationName string
	variables     map[string]interface{}
}

func decodeGQLQuery(r *http.Request) (*orchestration.SubscriptionData, error) {
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
	subID, err := gonanoid.Nanoid()
	subData := &orchestration.SubscriptionData{
		ID:            subID,
		GraphQLParams: params,
	}
	return subData, err
}

func (s *SubscribeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := decodeGQLQuery(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Something went wrong"}`))
		return
	}
	s.StorageAdapter.NotifyNewSubscription(req)
	fmt.Fprint(w, `{"message":"OK"}`)
}

// UnsubscribeHandler handles the endpoint for deleting subscriptions and contains a reference to the Broker
type UnsubscribeHandler struct {
	Broker         *orchestration.Broker
	StorageAdapter subscriptionStorageAdapter
}

func (s *UnsubscribeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello World")
}
