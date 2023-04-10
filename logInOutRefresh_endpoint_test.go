package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestSuccessfulUserLogin(t *testing.T) {
	// Create http recorder to record response
	recorder := httptest.NewRecorder()

	email := genRandomEmail()
	hashedPass, _ := hashAndSalt("somePassword")

	// Insert a new user into the database
	clients.InsertOne(context.Background(), bson.D{
		{Key: "email", Value: email},
		{Key: "hashedPassword", Value: hashedPass},
		{Key: "firstName", Value: "John"},
		{Key: "lastName", Value: "Smith"},
	})

	// Create a new user json
	login := `{"email": "` + email + `", "password": "somePassword"}`

	// Create a new request
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(login))

	// Send request to service
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Error("Endpoint did not return correct status code")
	}
}

func TestFailedLoginWithInvalidEmail(t *testing.T) {
	// Create http recorder to record response
	recorder := httptest.NewRecorder()

	email := genRandomEmail()
	// Create a new user json
	user := `{"email": "` + email + `", "password": "somePassword"}`

	// Create a new request
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(user))

	// Send request to service
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Error("Endpoint did not return correct status code")
	}
}

func TestLoginWithInvalidPassword(t *testing.T) {
	// Create http recorder to record response
	recorder := httptest.NewRecorder()

	email := genRandomEmail()
	hashedPass, _ := hashAndSalt("somePassword")

	// Insert a new user into the database
	clients.InsertOne(context.Background(), bson.D{
		{Key: "email", Value: email},
		{Key: "hashedPassword", Value: hashedPass},
		{Key: "firstName", Value: "John"},
		{Key: "lastName", Value: "Smith"},
	})

	// Create a new user json with the wrong password
	login := `{"email": "` + email + `", "password": "someOtherPassword"}`

	// Create a new request
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(login))

	// Send request to service
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Error("Endpoint did not return correct status code")
	}
}

func TestLoginSuspendedAccount(t *testing.T) {
	// Create http recorder to record response
	recorder := httptest.NewRecorder()

	email := genRandomEmail()
	hashedPass, _ := hashAndSalt("somePassword")

	// Insert a new user into the database but suspend the account
	clients.InsertOne(context.Background(), bson.D{
		{Key: "email", Value: email},
		{Key: "hashedPassword", Value: hashedPass},
		{Key: "firstName", Value: "John"},
		{Key: "lastName", Value: "Smith"},
		{Key: "suspended", Value: true},
	})

	// Create a new user json
	login := `{"email": "` + email + `", "password": "somePassword"}`

	// Create a new request
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(login))

	// Send request to service
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Error("Endpoint did not return correct status code")
	}
}

// func TestLoginWithInvalidPayload(t *testing.T)
// func TestCorrectCookieCreation(t *testing.T)

// BUG: This refresh is not working
// func TestRefresh(t *testing.T) {
// 	// Create http recorder to record response
// 	recorder := httptest.NewRecorder()

// 	email := genRandomEmail()
// 	hashedPass, _ := hashAndSalt("somePassword")

// 	// Insert a new user into the database
// 	clients.InsertOne(context.Background(), bson.D{
// 		{Key: "email", Value: email},
// 		{Key: "hashedPassword", Value: hashedPass},
// 		{Key: "firstName", Value: "John"},
// 		{Key: "lastName", Value: "Smith"},
// 	})

// 	// Create a new user json
// 	login := `{"email": "` + email + `", "password": "somePassword"}`

// 	// Create a new request
// 	req1, _ := http.NewRequest("POST", "/login", strings.NewReader(login))

// 	// Send request to service
// 	handler.ServeHTTP(recorder, req1)

// 	tokenCookies := recorder.Result().Cookies()

// 	fmt.Println(tokenCookies)

// 	// make a new request with the cookie
// 	req2, _ := http.NewRequest("POST", "/refresh", nil)

// 	for _, cookie := range tokenCookies {
// 		req2.AddCookie(cookie)
// 	}
// 	// req2.AddCookie(tokenCookies)

// 	fmt.Println(req2.Cookies())

// 	handler.ServeHTTP(recorder, req2)

// 	fmt.Println(recorder.Body)
// 	fmt.Println(recorder.Code)

// 	if recorder.Code != http.StatusOK {
// 		t.Error("Endpoint did not return correct status code")
// 	}
// }
