// Package stack is a collection of middleware, handlers, and models that help facilitate the creation of golang web services.
package stack

import (
	"context"

	"gopkg.in/mgo.v2"
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

// SetJwtSecret sets the secret that will be used to sign and verify JWT tokens
func SetJwtSecret(secret []byte) {
	jwtSecret = secret
}
