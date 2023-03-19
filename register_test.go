package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Test case for successful user registration
func TestRegisterNewUser(t *testing.T) {
	// Get MongoDB collection and Redis client
	users := getUserCollection()
	rdb := getCache()

	// Get handler object
	handler := createHandler(users, rdb)

	// Create http recorder to record response
	recorder := httptest.NewRecorder()

	// Generate a new unique user email
	gofakeit.Seed(0)
	newUserEmail := gofakeit.Email()

	// Create a new user json
	user := `{"email": "` + newUserEmail + `", "password": "somePassword", "firstName": "John", "lastName": "Smith"}`

	// Create a new request
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(user))

	// Send request to service
	handler.ServeHTTP(recorder, req)

	// Check if user is indeed inserted into the database
	filter := bson.D{{Key: "email", Value: newUserEmail}}
	mongoErr := users.FindOne(context.Background(), filter).Err()

	if mongoErr == mongo.ErrNoDocuments {
		t.Error("User was not added to database")
	}

	// Check if user creation response is sent to the client
	if recorder.Code != http.StatusCreated {
		t.Error("Endpoint did not return correct status code")
	}
}

// Test case for attempting to register a user that already exists
func TestRegisterPreexistingUser(t *testing.T) {
	// Get MongoDB collection and Redis client
	users := getUserCollection()
	rdb := getCache()

	// Get handler object
	handler := createHandler(users, rdb)

	// Create http recorder to record response
	recorder := httptest.NewRecorder()

	// Generate a new unique user email
	gofakeit.Seed(0)
	newUserEmail := gofakeit.Email()

	// Pre-insert a user into the database (so that we can attempt to register it again)
	users.InsertOne(context.Background(), bson.D{
		{Key: "email", Value: newUserEmail},
		{Key: "password", Value: "somePasswordHash"},
		{Key: "firstName", Value: "John"},
		{Key: "lastName", Value: "Smith"},
	})

	// New user json with same email as the pre-inserted user
	user := `{"email": "` + newUserEmail + `", "password": "somePassword", "firstName": "John", "lastName": "Smith"}`

	// Create a new request
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(user))

	// Send request to service
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Error("Endpoint did not return correct status code")
	}
}

// Test case for attempting to register a user with an invalid payload
func TestInvalidRegisterPayload(t *testing.T) {
	// Get MongoDB collection and Redis client
	users := getUserCollection()
	rdb := getCache()

	// Get handler object
	handler := createHandler(users, rdb)

	// Create http recorder to record response
	recorder := httptest.NewRecorder()

	// New user json with missing firstName field
	// BUG: This should be an invalid payload (missing firstName field), but it is not
	user := `{"email": "john@smith.com", "password": "somePassword", "lastName": "Smith"}`

	// New user json with misspelled email field
	// user = `{"emal": "john@smith.com", "password": "somePassword", "lastName": "Smith"}`

	// Create a new request
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(user))

	// Send request to service
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Error("Endpoint did not return correct status code")
	}
}
