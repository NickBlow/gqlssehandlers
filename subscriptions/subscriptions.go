package subscriptions

// Data encompasses a particular subscription and the parameters for the GraphQL Query that should be performed.
// It will use the schema passed into the handler config.
// RequestString is the raw GraphQL request string
// VariableValues are the values of the variables in the query
// This must be capable of being stored in and retreived from a database
type Data struct {
	SubscriptionID string
	ClientID       string
	RequestString  string
	VariableValues map[string]interface{}
}

// WrappedEvent contains the information needed to be sent to clients
// The event should be the results of the graphql query, and finished should be a boolean
// detailing whether more events of this type should be expected
type WrappedEvent struct {
	SubscriptionID string
	ClientID       string
	QueryResult    interface{}
	Finished       bool
}
