package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

var jwtSecret []byte

func init() {
	// The secret is b64 encoded
	jwtSecret64 := os.Getenv("JWT_SECRET")
	if jwtSecret64 == "" {
		panic("JWT_SECRET env var not set")
	}
	jwtSecretBytes, err := base64.StdEncoding.DecodeString(jwtSecret64)
	if err != nil {
		panic("Couldn't decode the JWT secret. Make sure JWT_SECRET env variable is set to something b64 encoded")
	}
	jwtSecret = jwtSecretBytes
}

func decodeJWT(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
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

// SymmetricJWTMiddleware is an example authorization middleware which authorizes & authenticates
// the user from a JWT in the Authorization header. The JWT is signed with a symmetric algorithm
func SymmetricJWTMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(401)
			w.Write([]byte("{}"))
			return
		}
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims, err := decodeJWT(tokenString)
		fmt.Println(claims)
		fmt.Println(err)
		next.ServeHTTP(w, r)
	}
}
