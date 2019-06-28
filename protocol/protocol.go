// Package protocol is an implementation of https://github.com/apollographql/subscriptions-transport-ws/blob/master/PROTOCOL.md
// based on the work by eientei on https://github.com/eientei/wsgraphql
// GQL_INIT, GQL_START, GQL_STOP, GQL_CONNECTION_TERMINATE will be sent to the subscription endpoint,
// and GQL_ERROR will be returned synchronously in case of an error, otherwise a 200 with {"type":"GQL_CONNECTION_ACK"} will be returned.
// GQL_COMPLETE, GQL_KEEPALIVE and GQL_DATA will be sent over the streaming endpoint.
// GQL_INIT will respond with the ClientIDHeader, defined in the clientid package, as well as a cookie.
package protocol

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/parser"
)

// OperationType type represents a GQL operation type
type OperationType string

// The different operation types it will handle
const (
	// GQLUnknown is the default
	GQLUnknown OperationType = ""

	// Client to Server types

	GQLConnectionInit      = "GQL_INIT"
	GQLStart               = "GQL_START"
	GQLStop                = "GQL_STOP"
	GQLConnectionTerminate = "GQL_CONNECTION_TERMINATE"

	// Server to Client

	GQLConnectionAck       = "GQL_CONNECTION_ACK"
	GQLData                = "GQL_DATA"
	GQLError               = "GQL_ERROR"
	GQLComplete            = "GQL_COMPLETE"
	GQLConnectionKeepAlive = "GQL_KEEPALIVE"
)

// GQLOverWebsocketProtocol is the wrapper for the protocol
type GQLOverWebsocketProtocol struct {
	Payload *PayloadBytes `json:"payload,omitempty"`
	Type    string        `json:"type"`
	ID      string        `json:"id,omitempty"`
}

// PayloadBytes is a utility combining functionality of interface{} and json.RawMessage
type PayloadBytes struct {
	Bytes []byte
	Value interface{}
}

// UnmarshalJSON saves bytes for later deserialization, just as json.RawMessage
func (payload *PayloadBytes) UnmarshalJSON(b []byte) error {
	payload.Bytes = b
	return nil
}

// MarshalJSON serializes interface value
func (payload *PayloadBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(payload.Value)
}

// GQLStartPayload represents the data sent on a GQL start message
type GQLStartPayload struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

type gqlValidationError struct {
	Errors []gqlerrors.FormattedError `json:"errors"`
}

// Response is a thin wrapper around HTTP status code & body
type Response struct {
	ExtraHeaders map[string]string
	Message      []byte
	StatusCode   int
}

// OKResponse returns a default ACK response, using GQLConnectionAck
func OKResponse() *Response {
	return &Response{
		ExtraHeaders: map[string]string{},
		Message:      []byte(`{"type":"` + GQLConnectionAck + `"}`),
		StatusCode:   http.StatusOK,
	}
}

// ServerErrorResponse returns a default Server error response
func ServerErrorResponse() *Response {
	return &Response{
		Message:    []byte(`{"type":"` + GQLError + `", "errors":[{"message": "Something went wrong"}]}`),
		StatusCode: http.StatusInternalServerError,
	}
}

// BadRequestResponse returns a default Bad Error response
func BadRequestResponse() *Response {
	return &Response{
		Message:    []byte(`{"type":"` + GQLError + `", "errors":[{"message": "Please send a valid payload"}]}`),
		StatusCode: http.StatusBadRequest,
	}
}

// DecodePayload decodes the request body as GQLOverWebsocketProtocol
func DecodePayload(body []byte) (*GQLOverWebsocketProtocol, error) {
	var req GQLOverWebsocketProtocol
	err := json.Unmarshal(body, &req)
	fmt.Println(string(body))
	if err != nil {
		return nil, err
	}
	return &req, nil
}

//KeepAlivePayload is a pre-marshalled JSON string representing the keepalive
const KeepAlivePayload = `{"type":"` + GQLConnectionKeepAlive + `"}`

func validationErrorResponse(errors []gqlerrors.FormattedError) *Response {
	val, err := json.Marshal(GQLOverWebsocketProtocol{
		Type:    GQLError,
		Payload: &PayloadBytes{Value: gqlValidationError{Errors: errors}},
	})
	if err != nil {
		fmt.Println("error while marshalling validation errors")
		fmt.Println(err)
		return ServerErrorResponse()
	}
	return &Response{
		StatusCode: 400,
		Message:    val,
	}
}

// ValidatePayload validates a grapqhl payload without executing it
func ValidatePayload(gqlPayload GQLStartPayload, schema *graphql.Schema) *Response {
	// validate without executing - ignoring extensions for now
	AST, err := parser.Parse(parser.ParseParams{Source: gqlPayload.Query})

	if err != nil {
		formatted := gqlerrors.FormatErrors(err)
		return validationErrorResponse(formatted)
	}
	validationResult := graphql.ValidateDocument(schema, AST, nil)
	if !validationResult.IsValid {
		return validationErrorResponse(validationResult.Errors)
	}
	return nil
}
