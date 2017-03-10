Stack
=====

Stack is a collection of middleware, handlers, and models that help
facilitate the creation of golang web services.

### Dependencies
- Go v1.7 (For the use of `context`)
- [golang.org/x/crypto/bcrypt](https://golang.org/x/crypto/bcrypt)
- [github.com/bradialabs/shortid](https://github.com/bradialabs/shortid)
- [github.com/dgrijalva/jwt-go](https://github.com/dgrijalva/jwt-go)
- [gopkg.in/mgo.v2](https://gopkg.in/mgo.v2)

User Model
----------

The User model is a simple struct for working with and
storing users in MongoDB. It has the following structure:

```go
type User struct {
	ID        string      `json:"id" bson:"_id,omitempty"`
	Email     string      `json:"email"`
	Password  string      `json:"-"`
	Created   time.Time   `json:"created"`
	FirstName string      `json:"first"`
	LastName  string      `json:"last"`
	Data      interface{} `json:"data"`
}
```

The `Data` property is provided for custom data that your application
will store with the user. You can store any struct and it will be stored
as an embedded document in MongoDB.

### Functions

#### NewUser

NewUser creates a new User and saves it in the database

```go
func NewUser(email string, pass string,
  firstName string, lastName string, db *mgo.Database) (*User, error)
```

#### FindUserByID

FindUserByID searches for an existing user with the passed ID

```go
func FindUserByID(id string, db *mgo.Database) (*User, error)
```
#### FindUserByEmail

FindUserByEmail searches for an existing user with the passed Email

```go
func FindUserByEmail(email string, db *mgo.Database) (*User, error)
```

#### Save

Save Upserts the user into the database

```go
func (user *User) Save(db *mgo.Database) error
```

#### CheckPassword

CheckPassword will check a passed password string with the stored hash

```go
func (user *User) CheckPassword(password string) error
```

Middleware
----------

The middleware now uses the built in context functionality of go. It sets
values on the request context. This allows the middleware to maintain the
original handler interface.

### MongoMiddleware

Creates a connection to MongoDB using the [mgo](https://labix.org/mgo) library.

### BasicMiddleware

Adds basic authentication to routes. Uses the User model. The route also
needs to use the MongoMiddleware.

### JwtAuthMiddleware

Adds JWT Bearer Authentication to routes. Uses the User model. The route also
needs to use the MongoMiddleware. In order to use JWT you must set the secret
key using the `stack.SetJwtSecret([]byte)` function.

Handlers
--------

### SignUpHandler

Handler for signing up new users from a Post request. The body of the request
must have the following JSON format.

```javascript
{
  first:"First Name"
  last:"Last Name"
  email:"user@email.org"
  pass:"pa$$word"
}
```

The handler will return JSON data back to the caller indicating success

```javascript
{
  "result":"success"
}
```

### SignInHandler

Handler for signing in users. This handler requires the BasicMiddleware and
MongoMiddleware. It will verify the email and password sent via basic auth and
return a signed, base64 encoded JWT token.

```javascript
{
  "token":"aksdfkjasdf.ajksdfkajsldf.akjlsdfhajklsdf"
}
```

Context
-------

This library uses the net.Context package for passing data through
middleware to the handlers. There are a few helper functions provided
for getting information out of the context.

### GetDb

GetDb grabs the mgo database from the context

```go
func GetDb(ctx context.Context) *mgo.Database
```

### GetUser

GetUser grabs the current user from the context

```go
func GetUser(ctx context.Context) *User
```

Example
-------

Here is an example of how to use stack with the [Chi](https://github.com/pressly/chi) router.

```go
package main

import (
	"io/ioutil"
	"net/http"
	"encoding/json"
	"log"

	"github.com/bradialabs/stack"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2/bson"
)

type UserData struct {
	Stuff string
}

func main() {
	stack.SetJwtSecret([]byte("secret"))

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Post("/api/1/signup", SignupRouter())
	r.Mount("/api/1", ApiRouter())
	r.Mount("/api/1/login", LoginRouter())

	http.ListenAndServe(":3333", r)
}

func SignupRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(stack.MongoMiddleware)

	r.Get("/", stack.SignUpHandler)
	return r
}

func LoginRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(stack.MongoMiddleware)
	r.Use(stack.BasicMiddleware)

	r.Get("/", stack.SignInHandler)
	return r
}

func ApiRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(stack.MongoMiddleware("chitest", ""))
	r.Use(stack.JwtAuthMiddleware)

	r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
		user := stack.GetUser(r.Context())
		j, er := json.Marshal(&user)
		if er != nil {
			log.Fatal(er)
		}
		w.Write(j)
	})

	r.Put("/me", func(w http.ResponseWriter, r *http.Request) {
		db := stack.GetDb(r.Context())
		user := stack.GetUser(r.Context())

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		userData := UserData{}
		err = json.Unmarshal(body, &userData)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		user.Data = userData
		user.Save(db)

		j, er := json.Marshal(&user)
		if er != nil {
			log.Fatal(er)
		}
		w.Write(j)
	})

	return r
}

```
