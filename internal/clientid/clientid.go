package clientid

import (
	"context"
	"fmt"
	"net/http"

	gonanoid "github.com/matoous/go-nanoid"
)

type contextKeyType string

// ClientIDKey is the key for a value on the request context it will pick it up and use instead of the default cookie
// The value should be a string representing the client id
const ClientIDKey contextKeyType = "GQLSSE_CLIENT_ID"

const defaultCookieName = "GQLSSE_CLIENT_ID"

// GetClientIDFromRequest returns the client id from the request
func GetClientIDFromRequest(r *http.Request) string {
	clientIDFromContext := r.Context().Value(ClientIDKey)
	if val, ok := clientIDFromContext.(string); clientIDFromContext != nil && ok {
		return val
	}
	return getDefaultCookie(r)
}

func getDefaultCookie(r *http.Request) string {
	cookie, err := r.Cookie(defaultCookieName)
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
		Name:  defaultCookieName,
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
