package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
)

// SuspendForm describes the expected json payload when a user suspension request is made
type SuspendForm struct {
	Id string `json:"id"`
}

// suspendUser is an only admin accessible handler for suspending a user
func makeSuspendClient(users *mongo.Collection, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add ID to blacklist until token expires
		code, claim := processClaim(r)
		if code != http.StatusOK {
			w.WriteHeader(code)
			log.Println(w.Write([]byte(`{"message": "Unable to process JWT token"}`)))
			return
		}

		status, _ := authAndAuthorisedAdmin(users, claim)

		if status == http.StatusOK {
			var suspendUser SuspendForm
			// Get the JSON body and decode into credentials
			err := json.NewDecoder(r.Body).Decode(&suspendUser)
			if err != nil {
				// If the structure of the body is wrong, return an HTTP error
				w.WriteHeader(http.StatusBadRequest)
				log.Println(w.Write([]byte(`{"message": "Invalid request body"}`)))
				return
			}
			rdb.Set(context.Background(), suspendUser.Id, suspendUser.Id, maxJwtTokenExpiration)
		}
	}
}
