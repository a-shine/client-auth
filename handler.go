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

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var collection *mongo.Collection
var ctx = context.TODO()

var rdb *redis.Client
var maxExperiation time.Duration

var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func initCache() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0, // use default DB
	})
}

func initMaxExperation() {
	mins, err := strconv.Atoi(os.Getenv("JWT_TOKEN_EXP_MIN"))
	if err != nil {
		panic(err)
	}
	maxExperiation = time.Duration(mins) * time.Minute
}

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

type SuspendForm struct {
	Id string `json:"id"`
}

type DeleteForm struct {
	Id string `json:"id"`
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
		Suspended:      false,
		Admin:          false,
	}
	collection.InsertOne(ctx, user)
}

// BUG: Get unauthorized error when trying to login with correct credentials
func Signin(w http.ResponseWriter, r *http.Request) {
	var creds LoginForm
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"Invalid request payload"}`))
		return
	}

	// Get the expected password from our in memory map
	// expectedPassword, ok := users[creds.Username]
	user := &User{}
	notFoundErr := collection.FindOne(ctx, bson.D{{"email", creds.Email}}).Decode(user)
	if notFoundErr == mongo.ErrNoDocuments {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"No account registered with this email"}`))
		return
	}

	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	validPass := verifyPassword(user.HashedPassword, creds.Password)
	if !validPass {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Incorrect password"}`))
		return
	}

	// if user is suspended do not issue token
	if user.Suspended {
		fmt.Println("User is suspended")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"User is suspended"}`))
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
		Id: user.Id.Hex(),
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

	// If user is not authorised then do not refresh token
	_, user := authenticate(r)
	if user.Suspended {
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

func isAuthorised(user *User) bool {
	return !user.Suspended
}

func authenticate(r *http.Request) (int, *User) {
	user := &User{}
	// (BEGIN) The code uptil this point is the same as the first part of the `Welcome` route
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			return http.StatusUnauthorized, user
		}
		return http.StatusBadRequest, user
	}
	tknStr := c.Value
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return http.StatusUnauthorized, user
		}
		return http.StatusBadRequest, user
	}
	if !tkn.Valid {
		return http.StatusUnauthorized, user
	}

	// Check user is auth and active - get by id
	// get user by id

	objID, _ := primitive.ObjectIDFromHex(claims.Id)
	value := collection.FindOne(ctx, bson.M{"_id": objID}).Decode(user)

	if value == mongo.ErrNoDocuments || user.Suspended {
		return http.StatusUnauthorized, user
	} else {
		return http.StatusOK, user
	}

}

// only admin can do this
func SuspendUser(w http.ResponseWriter, r *http.Request) {
	// add id to blacklist until token expires
	status, user := authenticate(r)
	if status == http.StatusOK && user.Admin {
		var suspendUser SuspendForm
		// Get the JSON body and decode into credentials
		err := json.NewDecoder(r.Body).Decode(&suspendUser)
		if err != nil {
			// If the structure of the body is wrong, return an HTTP error
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		rdb.Set(context.Background(), suspendUser.Id, suspendUser.Id, maxExperiation)
	}
}

// TODO
// Make this service read the gateway.conf file to get the list of auth services
// When user wishes to delete account, notify admin, admin can then delete user from all services by submitting the user id to this endpoint
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	// send delete signal to other services (each service will deal with delete it its own way)
	// add id to blacklist until token expires

	// submit a user deletion job in the gateway? The gateway will periodically delete user data for each auth service
	// Do this in the redis table?

	// publish to redis delete-user channel
	// gateway will listen to this channel and delete user data from all services
	// gateway will also delete user data from its own database
	status, user := authenticate(r)
	if status == http.StatusOK && user.Admin {
		var deleteUser DeleteForm
		// // Get the JSON body and decode into credentials
		err := json.NewDecoder(r.Body).Decode(&deleteUser)
		if err != nil {
			// If the structure of the body is wrong, return an HTTP error
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := rdb.Publish(context.Background(), "user-delete", deleteUser.Id).Err(); err != nil {
			fmt.Println(err)
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
}
