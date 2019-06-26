// Package subscriptions defines a handler endpoint that reacts to a subset of messages from the graphql-over-websocket protocol
// https://github.com/apollographql/subscriptions-transport-ws/blob/master/PROTOCOL.md
// GQL_START, GQL_STOP will be sent to this endpoint,
// and GQL_ERROR will be returned synchronously in case of an error, otherwise a 200 with {"type":"GQL_OK"} will be returned.
// GQL_COMPLETE and GQL_DATA will be sent over the streaming endpoint
package subscriptions

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/NickBlow/gqlssehandlers/internal/clientid"
	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/parser"
)

type subscriptionStorageAdapter interface {
	NotifyNewSubscription(ctx context.Context, clientID string, subscriptionID string, subscriberData orchestration.SubscriptionData) error
	NotifyUnsubscribe(ctx context.Context, clientID string, subscriptionID string) error
}

// SubscribeHandler handles the endpoint for processing new subscriptions and contains a reference to the Broker
type SubscribeHandler struct {
	Broker         *orchestration.Broker
	StorageAdapter subscriptionStorageAdapter
}

type gqlOverWebsocketProtocol struct {
	Payload interface{} `json:"payload,omitempty"`
	Type    string      `json:"type"`
	ID      string      `json:"id,omitempty"`
}

type gqlRequest struct {
	query         string
	operationName string
	variables     map[string]interface{}
}

type gqlValidationError struct {
	Errors []gqlerrors.FormattedError `json:"errors"`
}

func decodePayload(r *http.Request) (*gqlOverWebsocketProtocol, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var req gqlOverWebsocketProtocol
	err = json.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func decodeGQLQuery(r *http.Request, clientID string, operationID string) (*orchestration.SubscriptionData, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var req gqlRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	subData := &orchestration.SubscriptionData{
		SubscriptionID: fmt.Sprintf("%s_%s", clientID, operationID),
		VariableValues: req.variables,
		RequestString:  req.query,
	}
	return subData, err
}

type response struct {
	Message    []byte
	StatusCode int
}

func validationErrorResponse(errors []gqlerrors.FormattedError) *response {
	val, err := json.Marshal(gqlOverWebsocketProtocol{
		Type:    "GQL_ERROR",
		Payload: errors,
	})
	if err != nil {
		return &response{
			Message:    []byte(`{"message":"Something went wrong"}`),
			StatusCode: http.StatusInternalServerError,
		}
	}
	return &response{
		StatusCode: 400,
		Message:    val,
	}

}

func (s *SubscribeHandler) validatePayload(gqlPayload gqlRequest) *response {
	// validate without executing - ignoring extensions for now
	AST, err := parser.Parse(parser.ParseParams{Source: gqlPayload.query})

	if err != nil {
		formatted := gqlerrors.FormatErrors(err)
		return validationErrorResponse(formatted)
	}
	validationResult := graphql.ValidateDocument(s.Broker.Schema, AST, nil)
	if !validationResult.IsValid {
		return validationErrorResponse(validationResult.Errors)
	}
	return nil
}

func (s *SubscribeHandler) handlePayload(r *http.Request, clientID string) *response {
	req, err := decodePayload(r)
	if err != nil {
		return &response{
			Message:    []byte(`{"message":"Something went wrong"}`),
			StatusCode: http.StatusInternalServerError,
		}
	}
	switch req.Type {
	case "GQL_START":
		gqlPayload := req.Payload.(gqlRequest)
		validationResponse := s.validatePayload(gqlPayload)
		if validationResponse != nil {
			return validationResponse
		}
		s.StorageAdapter.NotifyNewSubscription(r.Context(), clientID, req.ID, orchestration.SubscriptionData{
			SubscriptionID: req.ID,
			ClientID:       clientID,
			RequestString:  gqlPayload.query,
			VariableValues: gqlPayload.variables,
		})
		return &response{
			Message:    []byte(`{"type":"GQL_OK"}`),
			StatusCode: http.StatusOK,
		}
	case "GQL_STOP":
		s.StorageAdapter.NotifyUnsubscribe(r.Context(), clientID, req.ID)
		return &response{
			Message:    []byte(`{"type":"GQL_OK"}`),
			StatusCode: http.StatusOK,
		}
	default:
		return &response{
			Message:    []byte(`{"message":"Invalid GQL type, supported GQL_START, GQL_STOP"}`),
			StatusCode: http.StatusBadRequest,
		}
	}
}

func (s *SubscribeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientID := clientid.GetClientIDFromRequest(r)
	res := s.handlePayload(r, clientID)
	w.WriteHeader(res.StatusCode)
	w.Write(res.Message)
}
