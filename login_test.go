package main

// import (
// 	"context"
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"

// 	"github.com/brianvoe/gofakeit"
// 	"go.mongodb.org/mongo-driver/bson"
// )

// func randomEmail() string {
// 	gofakeit.Seed(0)
// 	return gofakeit.Email()
// }

// func TestSuccessfulLogin(t *testing.T) {
// 	// Get MongoDB collection and Redis client
// 	users := getUserCollection()
// 	rdb := getCache()

// 	// Get handler object
// 	handler := createHandler(users, rdb)

// 	// Create http recorder to record response
// 	recorder := httptest.NewRecorder()

// 	email := randomEmail()
// 	hashedPass, _ := hashAndSalt("somePassword")

// 	// Insert a new user into the database
// 	users.InsertOne(context.Background(), bson.D{
// 		{Key: "email", Value: email},
// 		{Key: "hashedPassword", Value: hashedPass},
// 		{Key: "firstName", Value: "John"},
// 		{Key: "lastName", Value: "Smith"},
// 	})

// 	// Create a new user json
// 	login := `{"email": "` + email + `", "password": "somePassword"}`

// 	// Create a new request
// 	req, _ := http.NewRequest("POST", "/login", strings.NewReader(login))

// 	// Send request to service
// 	handler.ServeHTTP(recorder, req)

// 	if recorder.Code != http.StatusOK {
// 		t.Error("Endpoint did not return correct status code")
// 	}
// }

// func TestLoginWithInvalidEmail(t *testing.T) {
// 	// Get MongoDB collection and Redis client
// 	users := getUserCollection()
// 	rdb := getCache()

// 	// Get handler object
// 	handler := createHandler(users, rdb)

// 	// Create http recorder to record response
// 	recorder := httptest.NewRecorder()

// 	email := randomEmail()
// 	// Create a new user json
// 	user := `{"email": "` + email + `", "password": "somePassword"}`

// 	// Create a new request
// 	req, _ := http.NewRequest("POST", "/login", strings.NewReader(user))

// 	// Send request to service
// 	handler.ServeHTTP(recorder, req)

// 	if recorder.Code != http.StatusUnauthorized {
// 		t.Error("Endpoint did not return correct status code")
// 	}
// }

// func TestLoginWithInvalidPassword(t *testing.T) {
// 	// Get MongoDB collection and Redis client
// 	users := getUserCollection()
// 	rdb := getCache()

// 	// Get handler object
// 	handler := createHandler(users, rdb)

// 	// Create http recorder to record response
// 	recorder := httptest.NewRecorder()

// 	email := randomEmail()
// 	hashedPass, _ := hashAndSalt("somePassword")

// 	// Insert a new user into the database
// 	users.InsertOne(context.Background(), bson.D{
// 		{Key: "email", Value: email},
// 		{Key: "hashedPassword", Value: hashedPass},
// 		{Key: "firstName", Value: "John"},
// 		{Key: "lastName", Value: "Smith"},
// 	})

// 	// Create a new user json with the wrong password
// 	login := `{"email": "` + email + `", "password": "someOtherPassword"}`

// 	// Create a new request
// 	req, _ := http.NewRequest("POST", "/login", strings.NewReader(login))

// 	// Send request to service
// 	handler.ServeHTTP(recorder, req)

// 	if recorder.Code != http.StatusUnauthorized {
// 		t.Error("Endpoint did not return correct status code")
// 	}
// }

// func TestLoginSuspendedAccount(t *testing.T) {
// 	// Get MongoDB collection and Redis client
// 	users := getUserCollection()
// 	rdb := getCache()

// 	// Get handler object
// 	handler := createHandler(users, rdb)

// 	// Create http recorder to record response
// 	recorder := httptest.NewRecorder()

// 	email := randomEmail()
// 	hashedPass, _ := hashAndSalt("somePassword")

// 	// Insert a new user into the database but suspend the account
// 	users.InsertOne(context.Background(), bson.D{
// 		{Key: "email", Value: email},
// 		{Key: "hashedPassword", Value: hashedPass},
// 		{Key: "firstName", Value: "John"},
// 		{Key: "lastName", Value: "Smith"},
// 		{Key: "suspended", Value: true},
// 	})

// 	// Create a new user json
// 	login := `{"email": "` + email + `", "password": "somePassword"}`

// 	// Create a new request
// 	req, _ := http.NewRequest("POST", "/login", strings.NewReader(login))

// 	// Send request to service
// 	handler.ServeHTTP(recorder, req)

// 	if recorder.Code != http.StatusUnauthorized {
// 		t.Error("Endpoint did not return correct status code")
// 	}
// }

// // func TestLoginWithInvalidPayload(t *testing.T)
// // func TestCorrectCookieCreation(t *testing.T)
