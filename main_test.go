package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: Replace users and rdb with mock objects

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

	gofakeit.Seed(0)
	newUserEmail := gofakeit.Email()

	fmt.Println(newUserEmail)

	user := `{"email": "` + newUserEmail + `", "password": "somePassword", "firstName": "John", "lastName": "Smith"}`

	// Create a new request
	// Send request to service
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(user))

	// send request to service
	handler.ServeHTTP(recorder, req)

	fmt.Println(recorder.Body.String())
	fmt.Println(recorder.Code)

	// Check if user is in the database
	// Check if user already exists
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

func TestRegisterPreexistingUser(t *testing.T) {
	// Get handler object
	users := getUserCollection()

	rdb := getCache()

	gofakeit.Seed(0)
	newUserEmail := gofakeit.Email()

	_, err := users.InsertOne(context.Background(), bson.D{
		{Key: "email", Value: newUserEmail},
		{Key: "password", Value: "somePassword"},
		{Key: "firstName", Value: "John"},
		{Key: "lastName", Value: "Smith"},
	})

	if err != nil {
		t.Error(err)
	}

	handler := createHandler(users, rdb)

	user := `{"email": "` + newUserEmail + `", "password": "somePassword", "firstName": "John", "lastName": "Smith"}`

	// Create http recorder
	// Create http request
	recorder := httptest.NewRecorder()

	// Create a new request
	// Send request to service
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(user))

	// send request to service
	handler.ServeHTTP(recorder, req)

	fmt.Println(recorder.Body.String())
	fmt.Println(recorder.Code)

	if recorder.Code != http.StatusConflict {
		t.Error("Endpoint did not return correct status code")
	}

	// Create a new user json

}

func TestInvalidRegisterPayload(t *testing.T) {
	// Get handler object
	users := getUserCollection()

	rdb := getCache()

	handler := createHandler(users, rdb)

	// Create http recorder
	// Create http request
	recorder := httptest.NewRecorder()

	// Create a new request
	// Send request to service
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(""))

	// send request to service
	handler.ServeHTTP(recorder, req)

	fmt.Println(recorder.Body.String())
	fmt.Println(recorder.Code)

	if recorder.Code != http.StatusBadRequest {
		t.Error("Endpoint did not return correct status code")
	}
}

func TestRegister(t *testing.T) {
	handler := createHandler(getUserCollection(), getCache())
	data := map[string]string{
		"password":   "secret",
		"email":      "bob@myemail.com",
		"first_name": "Bob",
		"last_name":  "Smith",
	}
	buf, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	// convert map to json
	body := strings.NewReader(string(buf))
	req, err := http.NewRequest("POST", "/register", body)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf(
			"handler returned wrong status code: got %v want %v",
			status,
			http.StatusOK,
		)
	}
}
