package core

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"
	// "strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
)

// AuthResource creates the login and logout routes when bound
type AuthResource struct {
	router *gin.RouterGroup
}

// Bind creates the default user if it doesn't exist and adds the
// login and logout routes
func (auth AuthResource) Bind(db KeyValueDatabase) error {
	// connect to keyspace
	users, err := db.GetOrCreateKeyspace("users")
	if err != nil {
		log.Printf("Error creating keyspace: %v\n", err)
		return err
	}

	// create default user
	if err = createDefaultUser(users); err != nil {
		log.Printf("could not create default user: %v\n", err)
		return err
	}

	handler := AuthHandler{users}
	auth.router.POST("/login", handler.Login)
	auth.router.POST("/logout", handler.Logout)
	return nil
}

// AuthHandler handles the login and logout routes as well as session and cookie management.
type AuthHandler struct {
	users Keyspace
}

// setSession creates a secure cookie
func setSession(user BaseUser, sc *securecookie.SecureCookie, response http.ResponseWriter) {
	value := map[string]interface{}{
		"name":  user.Username,
		"reqid": user.RequestId,
	}
	if encoded, err := sc.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:     "session",
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().AddDate(0, 1, 0),
		}
		http.SetCookie(response, cookie)
	}

	// logged in
	loggedin := &http.Cookie{
		Name:     "part-of-the-club",
		Value:    "true",
		Path:     "/",
		HttpOnly: false,
	}
	http.SetCookie(response, loggedin)
}

// clearSession clears the session cookie
func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(response, cookie)

	// logged in
	loggedin := &http.Cookie{
		Name:     "part-of-the-club",
		Value:    "false",
		Path:     "/",
		HttpOnly: false,
		MaxAge:   -1,
	}
	http.SetCookie(response, loggedin)
}

// getSecureCookie is responsible for creating a secure cookie for each user
func getSecureCookie(user BaseUser, users Keyspace) (*securecookie.SecureCookie, error) {
	// create cookie handler
	var blockKey []byte

	// user doesn't have cookie block; create one
	if user.CookieBlock == "" {
		blockKey = securecookie.GenerateRandomKey(32)

		// encode block key as base64 string
		user.CookieBlock = base64.StdEncoding.EncodeToString(blockKey)

		// save updated user
		if err := SaveUser(&user, users); err != nil {
			log.Printf("could not save user: %v\n", err)
			return nil, err
		}

		// success. create secure cookie
		return securecookie.New(CookieHashKey, blockKey), nil
	}

	// User already has a cookie block
	// decode user's cookie block
	blockKey, err := base64.StdEncoding.DecodeString(user.CookieBlock)
	if err != nil {

		// could not decode user's cookie block
		log.Printf("could not decode user's cookie block: (%s) %v\n", user.Username, err)
		return nil, err
	}

	// success. create secure cookie
	return securecookie.New(CookieHashKey, blockKey), nil
}

// Login handles requests to /login
func (a *AuthHandler) Login(ctx *gin.Context) {
	ctx.Request.ParseForm()

	// get form values
	username := ctx.Request.Form.Get("username")
	password := ctx.Request.Form.Get("password")

	// authenticate user
	if username != "" && password != "" {
		var user BaseUser

		// get JSON encoded user object
		bytes, err := a.users.Get(username)
		if err != nil {

			// error reading user object

			log.Printf("could not get user from database: (%s) %v\n", username, err)
			ctx.String(500, "Error retreiving user")
			return
		} else if bytes == nil {

			// could not find username
			log.Printf("could not find user: (%s) %v\n", username, err)
			ctx.String(401, "Invalid username or password")
			return
		}

		// decode user object
		// log.Println("JSON user: ", string(bytes))
		err = json.Unmarshal(bytes, &user)
		if err != nil {

			log.Printf("could not unmarshal user: (%s) %v\n", username, err)
			ctx.String(500, "Error retreiving user")
			return
		}

		// login
		success := authenticate(user, password)
		if !success {

			log.Printf("failed login: (%s)\n", username)
			ctx.String(401, "Invalid username or password")
			return
		}

		sc, err := getSecureCookie(user, a.users)
		if err != nil {

			// failed to create secure cookie
			log.Printf("failed to create secure cookie: (%s) %v\n", username, err)
			ctx.String(500, "Internal Server Error")
			return
		}

		// create new session
		setSession(user, sc, ctx.Writer)

		// redirect home
		ctx.Redirect(302, "/")
	} else {

		// failed login due to invalid form

		log.Printf("Invalid login form: (%s : %s)\n", username, password)
		ctx.String(401, "Invalid username or password")
	}
}

// Logout handles requests to /logout
func (a *AuthHandler) Logout(ctx *gin.Context) {
	// clear cookie
	clearSession(ctx.Writer)

	// redirect home
	ctx.Redirect(302, "/")
}

// SaveUser saves a user object in the given Keyspace
func SaveUser(user *BaseUser, ks Keyspace) error {
	log.Printf("Saving user: %#v\n", user)

	// encode user as JSON
	newuser, err := json.Marshal(user)
	log.Printf("JSON user: %#v\n", string(newuser))
	if err != nil {

		// failed to encode new user
		log.Printf("could not marshal user: %v", err)
		return err
	}

	// update user with new cookie block
	err = ks.Update(user.Username, newuser)
	if err != nil {

		// failed to update user data
		log.Printf("could not save user: (%s) %v", user.Username, err)
		return err
	}

	// no errors
	return nil
}
