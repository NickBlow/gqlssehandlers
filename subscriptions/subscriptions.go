package subscriptions

// Data encompasses a particular subscription and the client who requested it
type Data struct {
	SubscriptionID string
	ClientID       string
}

// Query represents the parameters for the GraphQL Query that should be performed.
// It will use the schema passed into the handler config.
// RequestString is the raw GraphQL request string
// VariableValues are the values of the variables in the query
// This must be capable of being stored in and retreived from a database
type Query struct {
	RequestString  string
	VariableValues map[string]interface{}
}

// WrappedEvent contains the information needed to be sent to clients
// The QueryResult should be the results of the graphql query, and finished should be a boolean
// detailing whether more events of this type should be expected
type WrappedEvent struct {
	SubscriptionID string
	ClientID       string
	QueryResult    interface{}
	Finished       bool
}
