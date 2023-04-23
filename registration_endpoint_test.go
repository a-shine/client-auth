package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// https://github.com/tryvium-travels/memongo
// https://medium.com/@victor.neuret/mocking-the-official-mongo-golang-driver-5aad5b226a78
// Good repo with examples on how to mock mongoDB

// Here we are testing the API endpoints as a form of integration testing. This
// requires other dependent services to be available such as the database and
// cache so docker compose is used to facilitate orchestration

// Test case for successful user registration
func TestSuccessfulUserRegistration(t *testing.T) {
	recorder := httptest.NewRecorder()

	// Generate a new unique user email
	newUserEmail := genRandomEmail()

	// Register new user client JSON
	user := `{"email": "` + newUserEmail + `", "password": "somePassword", 
				"firstName": "John", "lastName": "Smith", "groups": ["admin"]}`

	// Create a new request
	req, _ := http.NewRequest("POST", "/register-user", strings.NewReader(user))

	// Send request to service
	handler.ServeHTTP(recorder, req)

	// Check if user is, indeed, inserted into the database
	filter := bson.D{{Key: "email", Value: newUserEmail}}
	mongoErr := clients.FindOne(context.Background(), filter).Err()
	assert.NotErrorIs(t, mongo.ErrNoDocuments, mongoErr)

	// Check for another database error
	assert.Nil(t, mongoErr)

	// Check if user creation response is sent to the client
	assert.Equal(t, http.StatusCreated, recorder.Code)

	// assert that the response body is what we expect
	assert.Equal(t, `{"message":"User registered successfully"}`, recorder.Body.String())
}

func TestSuccessfulServiceRegistration(t *testing.T) {
	recorder := httptest.NewRecorder()

	// Generate a new unique service account email
	// get current timestamp
	time := time.Now().Unix()

	// Create a new service account email
	newServiceEmail := strconv.FormatInt(time, 10) + "@service.com"

	// Register service client JSON
	servicePayload := `{"email": "` + newServiceEmail + `", "name": "Service A",
						"groups": ["tempProbes"]}`

	// Create a new request
	req, _ := http.NewRequest("POST", "/register-service", strings.NewReader(servicePayload))

	// Send request to service
	handler.ServeHTTP(recorder, req)

	// Check if user is indeed inserted into the database
	filter := bson.D{{Key: "email", Value: newServiceEmail}}
	mongoErr := clients.FindOne(context.Background(), filter).Err()
	assert.NotErrorIs(t, mongo.ErrNoDocuments, mongoErr)

	// Check for another database error
	assert.Nil(t, mongoErr)

	// Check if user creation response is sent to the client
	assert.Equal(t, http.StatusCreated, recorder.Code)

	// Test that the body returns both the service API token and a success message
	assert.Contains(t, recorder.Body.String(), `"apiToken"`)
}

// Test case for attempting to register a user that already exists
func TestFailedPreExistingUserRegistration(t *testing.T) {
	recorder := httptest.NewRecorder()

	// Generate a new unique user email
	newUserEmail := genRandomEmail()

	// Pre-insert a user into the database (so that we can attempt to register it again)
	clients.InsertOne(context.Background(), bson.D{
		{Key: "email", Value: newUserEmail},
		{Key: "password", Value: "somePasswordHash"},
		{Key: "firstName", Value: "John"},
		{Key: "lastName", Value: "Smith"},
	})

	// New user json with same email as the pre-inserted user
	user := `{"email": "` + newUserEmail + `", "password": "somePassword",
				"firstName": "John", "lastName": "Smith", "groups": ["admin"]}`

	// Create a new request
	req, _ := http.NewRequest("POST", "/register-user", strings.NewReader(user))

	// Send request to service
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusConflict, recorder.Code)
	assert.Equal(t, `{"message":"A user associated with the email address is already registered"}`, recorder.Body.String())
}

// Test case for attempting to register a user with an invalid payload
func TestFailedMissingFirstNameRegistration(t *testing.T) {
	recorder := httptest.NewRecorder()

	// New user json with missing firstName field
	user := `{"email": "john@smith.com", "password": "somePassword", 
				"lastName": "Smith", "groups": ["admin"]}`

	// Create a new request
	req, _ := http.NewRequest("POST", "/register-user", strings.NewReader(user))

	// Send request to service
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, `{"message":"Key: 'UserRegistrationForm.FirstName' Error:Field validation for 'FirstName' failed on the 'required' tag"}`, recorder.Body.String())
}

func TestFailedInvalidJson(t *testing.T) {
	recorder := httptest.NewRecorder()

	user := `"email": "john@smith.com",, "password": "somePassword", 
				"firstName": "John", "lastName": "Smith", "groups": ["admin"]}`

	// Create a new request
	req, _ := http.NewRequest("POST", "/register-user", strings.NewReader(user))

	// Send request to service
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, `{"message":"Invalid JSON payload"}`, recorder.Body.String())
}
