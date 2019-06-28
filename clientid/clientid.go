package clientid

import (
	"context"
	"fmt"
	"net/http"

	gonanoid "github.com/matoous/go-nanoid"
)

type contextKeyType string

// ClientIDKey is the key for a value on the request context it will pick it up and use instead of the default cookie
// The value should be a string representing the client id.
// Malformed (non-string) values will automatically fall back to other methods
const ClientIDKey contextKeyType = "gql_sse_client_id"

// ClientIDQueryString is the query string value that can optionally be set to associate a request with a particular client id 'out of the box'
// Any custom middleware setting the ClientIDKey on the context will take priority. The query string will take priority over the header.
// To use this correctly, it's recommended you do a GQL_INIT call to the subscription endpoint, and use the cookie provided there as the value
// for this query string on the subscription endpoint. This is the only default method guaranteed to work across multiple browser tabs.
const ClientIDQueryString = "gql_sse_client_id"

// ClientIDHeader is the header value that can optionally be set to associate a request with a particular client id 'out of the box'
// Any custom middleware setting the ClientIDKey on the context will take priority. If multiple headers are set with this value,
// it will take the first one.
// To use this correctly, it's recommended you connect to the streaming endpoint, save the cookie in a value and set that value as the header
// on requests to the subscription endpoint.
// As you cannot set headers on an EventSource, this may not be safe to use across multiple browser tabs.
// This is not case sensitive in Go, but may be in other languages. It will be returned from the server exacly as below.
const ClientIDHeader = "X-Gql-Sse-Client-Id"

// DefaultCookieName is the name of the cookie that will contain the client ID. The ClientIDHeader and ClientIDQueryString will take priority over the cookie.
// By default, this cookie is set when it does not exist.
// As the server cannot handle multiple connected clients with the same id (it will replace the latest one),
// this is not safe to use across multiple browser tabs if they all need to be connected.
const DefaultCookieName = "gql_sse_client_id"

// GetClientIDFromRequest returns the client id from the request, starting with the context,
// and then checking the query string, header and finally cookie in that order
func GetClientIDFromRequest(r *http.Request) string {
	clientIDFromContext := r.Context().Value(ClientIDKey)
	if val, ok := clientIDFromContext.(string); clientIDFromContext != nil && ok {
		return val
	}
	return getClientIDFromDefaults(r)
}

func getClientIDFromDefaults(r *http.Request) string {
	if qs := r.URL.Query().Get(ClientIDQueryString); qs != "" {
		return qs
	}
	if headerArray, ok := r.Header[ClientIDHeader]; ok && len(headerArray) > 1 {
		return headerArray[0]
	}
	cookie, err := r.Cookie(DefaultCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func setDefaultCookie(w http.ResponseWriter) (string, error) {
	newClientID, err := gonanoid.Nanoid()
	if err != nil {
		return "", err
	}
	http.SetCookie(w, &http.Cookie{
		Name:  DefaultCookieName,
		Value: newClientID,
	})
	return newClientID, nil
}

// Middleware adds the clientID to the request, either by using the existing value in the context
// or by using a cookie fallback
func Middleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := GetClientIDFromRequest(r)
		var err error
		if clientID == "" {
			clientID, err = setDefaultCookie(w)
			if err != nil {
				fmt.Println("Couldn't generate client ID")
				http.Error(w, "Error", http.StatusInternalServerError)
				return
			}
		}
		ctx := r.Context()
		newContext := context.WithValue(ctx, ClientIDKey, clientID)
		newRequest := r.WithContext(newContext)
		next.ServeHTTP(w, newRequest)
	}
}
