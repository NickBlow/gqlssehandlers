package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type userIDKeyType string

// UserIDKey is the key for the value of the user id in the request context
const UserIDKey userIDKeyType = "user_id"

func decodeJWT(tokenString string, jwtSecret []byte) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

func decodeHeader(authHeader string, jwtSecret []byte) (map[string]interface{}, error) {
	if authHeader == "" {
		return nil, errors.New("Empty Authorization header")
	}
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	claims, err := decodeJWT(tokenString, jwtSecret)
	return claims, err
}

// CheckClaimsFunc should return a string representing the user id,
// or an error if the user is not allowed to access the resource
type CheckClaimsFunc func(map[string]interface{}) (string, error)

// GetSecretFunc should return the JWT signing secret
type GetSecretFunc func() []byte

// SymmetricJWTMiddleware is an example authorization middleware which authorizes & authenticates
// the user from a JWT in the Authorization header. The JWT is signed with a symmetric algorithm
func SymmetricJWTMiddleware(next http.Handler, checkClaimsFunc CheckClaimsFunc, getSecretFunc GetSecretFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		claims, err := decodeHeader(authHeader, getSecretFunc())
		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte("{}"))
			return
		}
		userID, err := checkClaimsFunc(claims)
		if err != nil {
			w.WriteHeader(403)
			w.Write([]byte("{}"))
			return
		}
		ctx := r.Context()
		newContext := context.WithValue(ctx, UserIDKey, userID)
		newRequest := r.WithContext(newContext)
		next.ServeHTTP(w, newRequest)
	}
}
