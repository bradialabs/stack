package stack

import (
	"encoding/base64"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"

	"golang.org/x/net/context"
	"gopkg.in/mgo.v2"
)

// MongoMiddleware adds mgo MongoDB to context
func MongoMiddleware(dbName string, connectURI string, next Handler) HandlerFunc {
	// setup the mgo connection
	session, err := mgo.Dial(connectURI)

	if err != nil {
		panic(err)
	}

	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		reqSession := session.Clone()
		defer reqSession.Close()
		db := reqSession.DB(dbName)
		ctx = context.WithValue(ctx, dbKey, db)
		next.ServeHTTPC(ctx, w, r)
	}
}

// BasicMiddleware adds Basic Auth to routes. This assumes
// that the route is using the MongoMiddleware
func BasicMiddleware(next Handler) HandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		db := GetDb(ctx)
		if db == nil {
			log.Print("No database context")
			http.Error(w, "Not authorized", 401)
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(s) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}

		b, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		pair := strings.SplitN(string(b), ":", 2)
		if len(pair) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}

		//Find the user in the database
		user, err := FindUserByEmail(pair[0], db)
		if err != nil || user == nil {
			log.Printf("User %+v not found.", pair[0])
			http.Error(w, "Not authorized", 401)
			return
		}

		//Check their password
		err = user.CheckPassword(pair[1])
		if err != nil {
			log.Printf("Invalid password for User: %+v.", pair[0])
			http.Error(w, "Not authorized", 401)
			return
		}

		//clear password
		user.Password = ""

		//Set the logged in user in the context
		ctx = context.WithValue(ctx, userKey, user)

		next.ServeHTTPC(ctx, w, r)
	}
}

// JwtAuthMiddleware JWT Bearer Authentication to routes. This assumes
// that the route is using the MongoMiddleware
func JwtAuthMiddleware(next Handler) HandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		db := GetDb(ctx)
		if db == nil {
			log.Print("No database context")
			http.Error(w, "Not authorized", 401)
		}

		token, err := jwt.ParseFromRequest(r, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil {
			http.Error(w, "Invalid Token", 401)
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid Token", 401)
			return
		}

		// Check signing method to avoid vulnerabilities
		if token.Method != jwt.SigningMethodHS256 {
			http.Error(w, "Invalid Token", 401)
			return
		}

		//Find the user in the database
		user, err := FindUserByID(token.Claims["id"].(string), db)
		if err != nil || user == nil {
			log.Printf("User %s not found.", token.Claims["id"].(string))
			http.Error(w, "Not authorized", 401)
			return
		}

		//clear password
		user.Password = ""

		//Set the logged in user in the context
		ctx = context.WithValue(ctx, userKey, user)

		next.ServeHTTPC(ctx, w, r)
	}
}
