package stack

import (
	"net/http"

	"gopkg.in/mgo.v2"

	"golang.org/x/net/context"
)

type key int

var dbKey key = 100000
var userKey key = 200000

var jwtSecret []byte

// GetDb grabs the mgo database from the context
func GetDb(ctx context.Context) *mgo.Database {
	return ctx.Value(dbKey).(*mgo.Database)
}

// GetUser grabs the current user from the context
func GetUser(ctx context.Context) *User {
	return ctx.Value(userKey).(*User)
}

// Handler is like net/http's http.Handler, but also includes a
// mechanism for serving requests with a context.
type Handler interface {
	ServeHTTPC(context.Context, http.ResponseWriter, *http.Request)
}

// HandlerFunc is like net/http's http.HandlerFunc, but supports a context
// object.
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// SetJwtSecret sets the secret that will be used to sign and verify JWT tokens
func SetJwtSecret(secret []byte) {
	jwtSecret = secret
}
