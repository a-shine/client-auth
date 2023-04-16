package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSuccessfulMe(t *testing.T) {
	recorder := httptest.NewRecorder()

	// Create a user
	email := genRandomEmail()
	hashedPass, _ := hashAndSalt("somePassword")

	// Insert a new user into the database
	result, _ := clients.InsertOne(context.Background(), bson.D{
		{Key: "email", Value: email},
		{Key: "hashedPassword", Value: hashedPass},
		{Key: "firstName", Value: "John"},
		{Key: "lastName", Value: "Smith"},
		{Key: "groups", Value: []string{""}},
	})

	// Get user valid token to append to request
	id := result.InsertedID.(primitive.ObjectID).Hex()

	token, _ := genToken(id, []string{""})

	// Create a new request
	req, _ := http.NewRequest("GET", "/me", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})

	// Send request to service
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	fmt.Println(recorder.Body.String())

	// Check that the user details are returned
	assert.Contains(t, recorder.Body.String(), email)
}
