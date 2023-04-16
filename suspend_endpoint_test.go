package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSuccessfulSuspend(t *testing.T) {
	// Create an admin user
	recorder := httptest.NewRecorder()

	adminUserEmail := genRandomEmail()
	adminHashedPass, _ := hashAndSalt("somePassword")

	// Insert a new user into the database
	result, _ := clients.InsertOne(context.Background(), bson.D{
		{Key: "email", Value: adminUserEmail},
		{Key: "hashedPassword", Value: adminHashedPass},
		{Key: "firstName", Value: "John"},
		{Key: "lastName", Value: "Smith"},
		{Key: "groups", Value: []string{"admin"}},
	})

	// Get admin user valid token to append to request
	adminId := result.InsertedID.(primitive.ObjectID).Hex()

	adminToken, _ := genToken(adminId, []string{"admin"})

	// Create a non-admin user

	email := genRandomEmail()
	hashedPass, _ := hashAndSalt("somePassword")

	// Insert a new user into the database
	id, _ := clients.InsertOne(context.Background(), bson.D{
		{Key: "email", Value: email},
		{Key: "hashedPassword", Value: hashedPass},
		{Key: "firstName", Value: "John"},
		{Key: "lastName", Value: "Smith"},
		{Key: "groups", Value: []string{"admin"}},
	})

	// Test that the admin user can suspend the non-admin user
	suspendPayload := `{"id": "` + id.InsertedID.(primitive.ObjectID).Hex() + `"}`

	// Create a new request
	req, _ := http.NewRequest("POST", "/suspend", strings.NewReader(suspendPayload))
	req.AddCookie(&http.Cookie{Name: "token", Value: adminToken})

	// Send request to service
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, `{"message":"Successfully suspended user"}`, recorder.Body.String())

	// Check redis to see if the user is in the blacklist
	x := rdb.Get(context.Background(), id.InsertedID.(primitive.ObjectID).Hex())
	// check if the user is in the blacklist
	assert.Equal(t, id.InsertedID.(primitive.ObjectID).Hex(), x.Val())

	// A succeful suspend is a correct response code and the user is in the blacklist in redis (the actual responsibility of suspending authorisation is at the gateway level)
}

func TestFailedNonAdminSuspend(t *testing.T) {
	recorder := httptest.NewRecorder()

	// Create two non-admin users
	email1 := genRandomEmail()
	email2 := genRandomEmail()

	hashedPass, _ := hashAndSalt("somePassword")

	// Insert a new user into the database
	result1, _ := clients.InsertOne(context.Background(), bson.D{
		{Key: "email", Value: email1},
		{Key: "hashedPassword", Value: hashedPass},
		{Key: "firstName", Value: "John"},
		{Key: "lastName", Value: "Smith"},
		{Key: "groups", Value: []string{""}},
	})

	result2, _ := clients.InsertOne(context.Background(), bson.D{
		{Key: "email", Value: email2},
		{Key: "hashedPassword", Value: hashedPass},
		{Key: "firstName", Value: "John"},
		{Key: "lastName", Value: "Smith"},
		{Key: "groups", Value: []string{""}},
	})

	user1Token, _ := genToken(result1.InsertedID.(primitive.ObjectID).Hex(), []string{""})

	// Test that one non-admin user cannot suspend the other
	suspendPayload := `{"id": "` + result2.InsertedID.(primitive.ObjectID).Hex() + `"}`

	// Create a new request
	req, _ := http.NewRequest("POST", "/suspend", strings.NewReader(suspendPayload))
	req.AddCookie(&http.Cookie{Name: "token", Value: user1Token})

	// Send request to service
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Equal(t, `{"message":"You are not authorised to perform this action"}`, recorder.Body.String())
}
