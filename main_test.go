package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMain(t *testing.T) {
	// Get handler object
	// service := initService()
}

func TestRegisterNewUser(t *testing.T) {
	// Get handler object
	users := getUserCollection()

	rdb := getCache()

	handler := createHandler(users, rdb)

	// Create http recorder
	// Create http request
	recorder := httptest.NewRecorder()

	// Create a new user json
	// TODO: Make email random
	user := `{"email": "johnsmith@mydomain.com", "password": "somePassword", "firstName": "John", "lastName": "Smith"}`

	// Create a new request
	// Send request to service
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(user))

	// send request to service
	handler.ServeHTTP(recorder, req)

	// Check if user is in the database
	// Check if user already exists
	filter := bson.D{{"email", "johnsmith@mydomain.com"}}
	mongoErr := users.FindOne(context.Background(), filter).Err()

	if mongoErr == mongo.ErrNoDocuments {
		t.Error("User was not added to database")
	}

	// Check if user creation response is sent to the client
	if recorder.Code != http.StatusCreated {
		t.Error("Endpoint did not return correct status code")
	}
}

func TestRegisterPreexistingUser(t *testing.T) {

}
