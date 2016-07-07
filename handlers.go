package stack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
)

type userdata struct {
	FirstName string `json:"first"`
	LastName  string `json:"last"`
	Email     string `json:"email"`
	Password  string `json:"pass"`
}

// SignUpHandler is a Handler function for handling a user
// user signup route
func SignUpHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	db := GetDb(ctx)
	if db == nil {
		http.Error(w, "No database context", 500)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	userInfo := userdata{}
	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	_, err = NewUser(userInfo.Email, userInfo.Password,
		userInfo.FirstName, userInfo.LastName, db)
	if err != nil {
		http.Error(w, err.Error(), 409)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"result":"success"}`))
}

// SignInHandler will return a JWT token for the user that signed in.
// This route must use the BasicMiddleware for authentication
func SignInHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	db := GetDb(ctx)
	if db == nil {
		http.Error(w, "No database context", 500)
		return
	}

	user := GetUser(ctx)
	if user == nil {
		http.Error(w, "No user context", 401)
		return
	}

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims["id"] = user.ID
	token.Claims["iat"] = time.Now().Unix()
	token.Claims["exp"] = time.Now().Add(time.Second * 3600 * 24).Unix()

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Failed to create token", 401)
		return
	}

	fmt.Fprintf(w, "{\"token\": \"%s\"}", tokenString)
}
