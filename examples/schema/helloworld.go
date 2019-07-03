package schema

import (
	"fmt"

	"github.com/graphql-go/graphql"
)

// HelloWorldSchema is the example schema
var HelloWorldSchema graphql.Schema

func init() {
	var fields = graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
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
	HelloWorldSchema = schema
}
