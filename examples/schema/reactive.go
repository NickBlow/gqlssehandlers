package schema

import (
	"context"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
)

// HelloReactiveSchema is a schema that resolves based on messages coming from a channel passed in via the context
// It is intended to be run in a loop in a goroutine - the resolve function will block until something comes into the channel
// Query is simply `{hello}`
var HelloReactiveSchema graphql.Schema

type channelKeyType string

const channelKey channelKeyType = "source_stream"

// SampleEvent is an example struct to demonstrate the schema - your real events will be more complex!
type SampleEvent struct {
	Name string
}

type channelClosedError struct{}

func (e *channelClosedError) Error() string {
	return "Channel closed"
}

// HasChannelClosedError checks whether the error was due to the communication channel being closed
func HasChannelClosedError(errors []gqlerrors.FormattedError) bool {
	for _, val := range errors {
		gqlError, ok := val.OriginalError().(*gqlerrors.Error) // gql library returns errors in a wrapped struct
		if _, isClosedErr := gqlError.OriginalError.(*channelClosedError); ok && isClosedErr {
			return true
		}
	}
	return false
}

// AddChannelToContext adds the sampleEvent channel to the provided context and returns a new context with that channel set
func AddChannelToContext(ctx context.Context, channel chan SampleEvent) context.Context {
	return context.WithValue(ctx, channelKey, channel)
}

func getChannelFromContext(ctx context.Context) (chan SampleEvent, bool) {
	u, ok := ctx.Value(channelKey).(chan SampleEvent)
	return u, ok
}

func init() {
	var fields = graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				channel, ok := getChannelFromContext(ctx)
				if !ok {
					panic("No channel in context") // you can probably handle this more elegantly
				}
				value, more := <-channel // Block until a value is sent down the channel, or it is closed
				if !more {
					return nil, &channelClosedError{}
				}
				return value.Name, nil
			},
		},
	}
	var rootQuery = graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	var schemaConfig = graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	var schema, err = graphql.NewSchema(schemaConfig)
	if err != nil {
		fmt.Println(err)
		panic("Couldn't create schema")
	}
	HelloReactiveSchema = schema
}
