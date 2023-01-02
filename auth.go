package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var collection *mongo.Collection
var ctx = context.TODO()

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func initDb() {
	mongoUri := "mongodb://" + os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT")
	clientOptions := options.Client().ApplyURI(mongoUri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database(os.Getenv("DB_NAME")).Collection("users")
}

// Hash and salt password
func hashAndSalt(pwd string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err)
	}
	return string(hash)
}

// Compare password with hash
func verifyPassword(hashedPwd string, plainPwd string) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(plainPwd))
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// Create a struct that models the structure of a user, both in the request body, and in the DB
type LoginForm struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegisterForm struct {
	Password  string `json:"password"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Change claim to be associated with Id instead of username
type Claims struct {
	Id string `json:"id"`
	jwt.RegisteredClaims
}

func Signup(w http.ResponseWriter, r *http.Request) {
	var creds RegisterForm
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if user already exists
	filter := bson.D{{"email", creds.Email}}
	mongoErr := collection.FindOne(ctx, filter).Err()

	if mongoErr != mongo.ErrNoDocuments {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"message":"An account with that email address is already registered"}`))
		return
	}

	// Create user
	hashedPassword := hashAndSalt(creds.Password)
	user := &User{
		Id:             primitive.NewObjectID(),
		Email:          creds.Email,
		FirstName:      creds.FirstName,
		LastName:       creds.LastName,
		HashedPassword: hashedPassword,
		Active:         true,
	}
	collection.InsertOne(ctx, user)
}

func Signin(w http.ResponseWriter, r *http.Request) {
	var creds LoginForm
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the expected password from our in memory map
	// expectedPassword, ok := users[creds.Username]
	user := &User{}
	notFoundErr := collection.FindOne(ctx, bson.D{{"email", creds.Email}}).Decode(user)

	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	validPass := verifyPassword(user.HashedPassword, creds.Password)
	if notFoundErr == mongo.ErrNoDocuments || !validPass {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Declare the expiration time of the token
	// here, we have kept it as 5 minutes
	expMin, _ := strconv.Atoi(os.Getenv("JWT_TOKEN_EXP_MIN"))
	mins := time.Duration(expMin) * time.Minute
	expirationTime := time.Now().Add(mins)
	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		// TODO: Change to Id
		Id: user.Id.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Finally, we set the client cookie for "token" as the JWT we just generated
	// we also set an expiry time which is the same as the token itself
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
}

// TODO: Check with database
func Refresh(w http.ResponseWriter, r *http.Request) {
	// (BEGIN) The code uptil this point is the same as the first part of the `Welcome` route
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tknStr := c.Value
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// (END) The code uptil this point is the same as the first part of the `Welcome` route

	// We ensure that a new token is not issued until enough time has elapsed
	// In this case, a new token will only be issued if the old token is within
	// 30 seconds of expiry. Otherwise, return a bad request status
	if time.Until(claims.ExpiresAt.Time) > 30*time.Second {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Now, create a new token for the current use, with a renewed expiration time
	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
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
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// immediately clear the token cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Expires: time.Now(),
	})
}

// only admin can do this
func deactivateUser() {
	// add id to blacklist until token expires
}

func deleteUser() {
	// send delete signal to other services (each service will deal with delete it its own way)
	// add id to blacklist until token expires
}
