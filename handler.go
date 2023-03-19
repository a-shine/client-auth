package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Claim describes the structure of a JWT claim (this is the same payload as in the
// https://github.com/a-shine/api-gateway repo which is what makes them compatible)
type Claim struct {
	Id     string   `json:"id"`
	Groups []string `json:"groups"` // TODO: Use this to check if user is admin, allows groups to be verified by the API gateway
	jwt.RegisteredClaims
}

// RegisterForm describes the expected json payload when a user registers
type RegisterForm struct {
	Password  string `json:"password" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

// LoginForm describes the expected json payload when a user logs in
type LoginForm struct {
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

// SuspendForm describes the expected json payload when a user suspension request is made
type SuspendForm struct {
	Id string `json:"id"`
}

// hashAndSalt hashes a provided password and returns the hashed password as a string
func hashAndSalt(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Unable to hash password: ", err)
		return "", err
	}
	return string(hash), nil
}

// verifyPassword compares a hashed password with the raw password
func verifyPassword(hashedPwd string, plainPwd string) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(plainPwd))
	return err == nil // true if the err is nil and false otherwise
}

func processClaim(r *http.Request) (int, *Claim) {
	claim := &Claim{}

	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			return http.StatusUnauthorized, nil
		}
		return http.StatusBadRequest, nil
	}
	tknStr := c.Value

	tkn, err := jwt.ParseWithClaims(tknStr, claim, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return http.StatusUnauthorized, nil
		}
		return http.StatusBadRequest, nil
	}
	if !tkn.Valid {
		return http.StatusUnauthorized, nil
	}

	return http.StatusOK, claim
}

func authenticate(users *mongo.Collection, claim *Claim) (int, *Client) {
	user := &Client{}

	// Get user by the ID in the token claim payload
	objID, _ := primitive.ObjectIDFromHex(claim.Id)
	value := users.FindOne(context.Background(), bson.M{"_id": objID}).Decode(user)

	// Check user if user can be authenticated
	if value == mongo.ErrNoDocuments {
		return http.StatusUnauthorized, user
	} else {
		return http.StatusOK, user
	}
}

func authAndAuthorised(users *mongo.Collection, claim *Claim) (int, *Client) {
	code, user := authenticate(users, claim)
	switch code {
	case http.StatusOK:
		if user.Suspended {
			return http.StatusUnauthorized, nil
		} else {
			return http.StatusOK, user
		}
	default:
		return code, nil
	}
}

func authAndAuthorisedAdmin(users *mongo.Collection, claim *Claim) (int, *Client) {
	code, user := authAndAuthorised(users, claim)
	switch code {
	case http.StatusOK:
		// If user group contains "admin" then they are authorised
		for _, group := range user.Groups {
			if group == "admin" {
				return http.StatusOK, user
			}
		}
		return http.StatusUnauthorized, nil
	default:
		return code, nil
	}
}

// makeRegisterHandler registers handler function for user registration endpoint
func makeRegisterHandler(users *mongo.Collection, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds RegisterForm

		// BUG: Does not check for missing fields
		// Get the JSON body and decode into credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			// If the structure of the body is wrong, return an HTTP error
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"Invalid JSON payload"}`))
			return
		}

		// validate json schema
		err = validate.Struct(creds)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"Invalid JSON payload"}`))
			return
		}

		// Check if user already exists
		filter := bson.D{{Key: "email", Value: creds.Email}}
		mongoErr := users.FindOne(context.Background(), filter).Err()

		if mongoErr != mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusConflict)
			log.Println(w.Write([]byte(`{"message":"An account with that email address is already registered"}`)))
			return
		}

		// Hash user password before storing in database
		hashedPassword, err := hashAndSalt(creds.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Create user
		user := &Client{
			Id:             primitive.NewObjectID(),
			Email:          creds.Email,
			FirstName:      creds.FirstName,
			LastName:       creds.LastName,
			HashedPassword: hashedPassword,
			Suspended:      false,
			Groups:         []string{},
		}

		// Insert user into database
		_, err = users.InsertOne(context.Background(), user)
		if err != nil {
			log.Println("Unable to insert user into database: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(w.Write([]byte(`{"message":"Unable to register user"}`)))
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message":"User registered successfully"}`))
	}
}

// login handler for user login
func makeLoginHandler(users *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds LoginForm

		// Get the JSON body and decode into credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			// If the structure of the body is wrong, return an HTTP error
			w.WriteHeader(http.StatusBadRequest)
			log.Println(w.Write([]byte(`{"message":"Invalid request payload"}`)))
			return
		}

		// Get the user details from the database
		user := &Client{}
		notFoundErr := users.FindOne(context.Background(), bson.D{{Key: "email", Value: creds.Email}}).Decode(user)
		if notFoundErr == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusUnauthorized)
			log.Println(w.Write([]byte(`{"message":"No account registered with this email"}`)))
			return
		}

		// Compare the provided password with the stored hashed password, if they do not match return an "Unauthorized"
		// status and an incorrect password message
		validPass := verifyPassword(user.HashedPassword, creds.Password)
		if !validPass {
			w.WriteHeader(http.StatusUnauthorized)
			log.Println(w.Write([]byte(`{"message":"Incorrect password"}`)))
			return
		}

		// Check if user is suspended. If suspended, do not issue a token
		if user.Suspended {
			w.WriteHeader(http.StatusUnauthorized)
			log.Println(w.Write([]byte(`{"message":"Account has been suspended"}`)))
			return
		}

		// Declare the expiration time of the token as determined by the maxJwtTokenExpiration variable
		expirationTime := time.Now().Add(maxJwtTokenExpiration)

		// Create the JWT claims, which includes the authenticated user ID and expiry time
		claims := &Claim{
			Id:     user.Id.Hex(),
			Groups: user.Groups,
			RegisteredClaims: jwt.RegisteredClaims{
				// In JWT, the expiry time is expressed as unix milliseconds
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}

		// Create the token with the HS256 algorithm used for signing, and the created claim
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Create the JWT string
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			// If there is an error in creating the JWT return an internal server error
			log.Println("Unable to sign token: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Finally, we set the client cookie for "token" as the JWT we just generated we also set an expiry time which is
		// the same as the token itself
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
			Path:    "/",
		})
	}
}

// Refresh is called by frontend when 401 is received it is a way of authenticating user and extending their session without the need for them to input there credentials again but also checks that they are still authorised to use the system
// refresh handler enabling authenticated non-suspended users to apply for new token lengthening their session. // If
// unable to authenticate user or is not authorised (e.g. suspended) then do not refresh token
func makeRefreshHandler(users *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code, claim := processClaim(r)
		if code != http.StatusOK {
			w.WriteHeader(code)
			log.Println(w.Write([]byte(`{"message": "Unable to process JWT token"}`)))
			return
		}

		code, _ = authAndAuthorised(users, claim)
		if code != http.StatusOK {
			w.WriteHeader(code)
			log.Println(w.Write([]byte(`{"message": "Unable to refresh token"}`)))
			return
		}

		// We ensure that a new token is not issued until enough time has elapsed. In this case, a new token will only be
		// issued if the old token is within 30 seconds of expiry. Otherwise, return a bad request status.
		if time.Until(claim.ExpiresAt.Time) > 30*time.Second {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(w.Write([]byte(`{"message": "Token still valid"}`)))
			return
		}

		// Now, create a new token for the current use, with a renewed expiration time
		expirationTime := time.Now().Add(5 * time.Minute)
		claim.ExpiresAt = jwt.NewNumericDate(expirationTime)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Set the new token as the users `token` cookie
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
			Path:    "/",
		})
	}
}

// logout handler for user logout
func makeLogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// Immediately clear the token cookie by setting cookie expiry to now
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Expires: time.Now(),
			Path:    "/",
		})
	}
}

// me handler for user details
func makeMeHandler(users *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code, claim := processClaim(r)
		if code != http.StatusOK {
			w.WriteHeader(code)
			log.Println(w.Write([]byte(`{"message": "Unable to process JWT token"}`)))
			return
		}

		status, user := authAndAuthorised(users, claim)
		if status == http.StatusOK {
			err := json.NewEncoder(w).Encode(user)
			if err != nil {
				log.Println("Unable to encode user: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			// w.WriteHeader(http.StatusUnauthorized)
			w.WriteHeader(status)
			return
		}
	}
}

// DO NOT DELETE USERS DIRECTLY FROM THE DATABASE. INSTEAD, USE THE DELETE USER HANDLER (NEEDS TO TRICKLE DOWN TO ALL SERVICES)
// deleteUser handler enables users to request for their data to be deleted. This communicates with the API Gateway
// through a pubsub 'user-delete' channel. The API Gateway will then communicate with each of the services that
// required authentication, so they can handle deletion of the user data they contain.
func makeDeleteUserHandler(users *mongo.Collection, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code, claim := processClaim(r)
		if code != http.StatusOK {
			w.WriteHeader(code)
			log.Println(w.Write([]byte(`{"message": "Unable to process JWT token"}`)))
			return
		}

		status, user := authAndAuthorised(users, claim)

		// TODO: Have some feedback system to deal with failed user deletion job requests
		// Make user deletion request to API Gateway requiring each authenticated service to delete user data
		if status == http.StatusOK {
			// Send delete signal to other Gateway (each authenticated service will deal with delete it its own way)
			// Gateway will listen to this channel and ask each service to delete user data
			if err := rdb.Publish(context.Background(), "user-delete", user.Id.Hex()).Err(); err != nil {
				log.Println("Unable to publish to Redis: ", err)
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			log.Println(w.Write([]byte(`{"message": "Unable to authenticate user"}`)))
			return
		}

		// Add ID to blacklist until token expires
		if err := rdb.Set(context.Background(), claim.Id, claim.Id, time.Until(claim.ExpiresAt.Time)).Err(); err != nil {
			log.Println("Unable to set blacklist: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		objID, _ := primitive.ObjectIDFromHex(claim.Id)
		_, err := users.DeleteOne(context.Background(), bson.M{"_id": objID})
		if err != nil {
			log.Println("Unable to delete user: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

// suspendUser is an only admin accessible handler for suspending a user
func makeSuspendUser(users *mongo.Collection, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add ID to blacklist until token expires
		code, claim := processClaim(r)
		if code != http.StatusOK {
			w.WriteHeader(code)
			log.Println(w.Write([]byte(`{"message": "Unable to process JWT token"}`)))
			return
		}

		status, _ := authAndAuthorisedAdmin(users, claim)

		if status == http.StatusOK {
			var suspendUser SuspendForm
			// Get the JSON body and decode into credentials
			err := json.NewDecoder(r.Body).Decode(&suspendUser)
			if err != nil {
				// If the structure of the body is wrong, return an HTTP error
				w.WriteHeader(http.StatusBadRequest)
				log.Println(w.Write([]byte(`{"message": "Invalid request body"}`)))
				return
			}
			rdb.Set(context.Background(), suspendUser.Id, suspendUser.Id, maxJwtTokenExpiration)
		}
	}
}
