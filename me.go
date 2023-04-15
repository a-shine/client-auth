package main

import (
	"encoding/json"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

// me handler for user details
// TODO: update this to return user details from database and not from JWT token
func makeMeHandler(users *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code, claim := processClaim(r)
		if code != http.StatusOK {
			w.WriteHeader(code)
			log.Println(w.Write([]byte(`{"message": "Unable to process JWT token"}`)))
			return
		}

		status, user := authAndAuthorised(users, claim)
		if status == http.StatusOK {
			err := json.NewEncoder(w).Encode(user)
			if err != nil {
				log.Println("Unable to encode user: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			// w.WriteHeader(http.StatusUnauthorized)
			w.WriteHeader(status)
			return
		}
	}
}
