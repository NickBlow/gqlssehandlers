package subscriptionhandlers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"encoding/json"

	"github.com/NickBlow/gqlssehandlers/clientid"
	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
	"github.com/NickBlow/gqlssehandlers/protocol"
	"github.com/NickBlow/gqlssehandlers/subscriptions"
)

type subscriptionStorageAdapter interface {
	NotifyNewSubscription(ctx context.Context, subscriberData subscriptions.Data, queryData subscriptions.Query) error
	NotifyUnsubscribe(ctx context.Context, subscriberData subscriptions.Data) error
}

// Handler handles the endpoint for processing new subscriptions and contains a reference to the Broker
type Handler struct {
	Broker         *orchestration.Broker
	StorageAdapter subscriptionStorageAdapter
}

func (s *Handler) handleGQLStart(ctx context.Context, req *protocol.GQLOverWebsocketProtocol, clientID string) *protocol.Response {
	if req.Payload == nil {
		return protocol.BadRequestResponse()
	}
	var gqlPayload protocol.GQLStartPayload
	err := json.Unmarshal(req.Payload.Bytes, &gqlPayload)
	if err != nil {
		fmt.Println(err)
		return protocol.BadRequestResponse()
	}
	validationResponse := protocol.ValidatePayload(gqlPayload, s.Broker.Schema)
	if validationResponse != nil {
		return validationResponse
	}
	s.StorageAdapter.NotifyNewSubscription(ctx, subscriptions.Data{
		SubscriptionID: req.ID,
		ClientID:       clientID,
	}, subscriptions.Query{
		RequestString:  gqlPayload.Query,
		VariableValues: gqlPayload.Variables,
	})
	return protocol.OKResponse()
}

func (s *Handler) handlePayload(r *http.Request, clientID string) *protocol.Response {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return protocol.ServerErrorResponse()
	}
	req, err := protocol.DecodePayload(body)
	if err != nil {
		fmt.Println(err)
		return protocol.BadRequestResponse()
	}
	switch req.Type {
	case "GQL_START":
		response := s.handleGQLStart(r.Context(), req, clientID)
		return response
	case "GQL_STOP":
		s.StorageAdapter.NotifyUnsubscribe(r.Context(), subscriptions.Data{
			SubscriptionID: req.ID,
			ClientID:       clientID,
		})
		return protocol.OKResponse()
	case "GQL_CONNECTION_TERMINATE":
		s.Broker.ClosingClients <- clientID
		return protocol.OKResponse()
	case "GQL_INIT":
		baseResponse := protocol.OKResponse()
		baseResponse.ExtraHeaders[clientid.ClientIDHeader] = r.Context().Value(clientid.ClientIDKey).(string)
		return baseResponse
	default:
		return protocol.BadRequestResponse()
	}
}

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientID := clientid.GetClientIDFromRequest(r)
	res := s.handlePayload(r, clientID)
	for k, v := range res.ExtraHeaders {
		w.Header().Set(k, v)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.StatusCode)
	w.Write(res.Message)
}
