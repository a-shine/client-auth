package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: Document user deletion mechanism

// DO NOT DELETE USERS DIRECTLY FROM THE DATABASE. INSTEAD, USE THE DELETE USER HANDLER (NEEDS TO TRICKLE DOWN TO ALL SERVICES)
// deleteUser handler enables users to request for their data to be deleted. This communicates with the API Gateway
// through a pubsub 'user-delete' channel. The API Gateway will then communicate with each of the services that
// required authentication, so they can handle deletion of the user data they contain.
func makeDeleteUserHandler(users *mongo.Collection, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code, claim := processClaim(r)
		if code != http.StatusOK {
			w.WriteHeader(code)
			log.Println(w.Write([]byte(`{"message": "Unable to process JWT token"}`)))
			return
		}

		status, user := authAndAuthorised(users, claim)

		// TODO: Have some feedback system to deal with failed user deletion job requests
		// Make user deletion request to API Gateway requiring each authenticated service to delete user data
		if status == http.StatusOK {
			// Send delete signal to other Gateway (each authenticated service will deal with delete it its own way)
			// Gateway will listen to this channel and ask each service to delete user data
			if err := rdb.Publish(context.Background(), "user-delete", user.Id.Hex()).Err(); err != nil {
				log.Println("Unable to publish to Redis: ", err)
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			log.Println(w.Write([]byte(`{"message": "Unable to authenticate user"}`)))
			return
		}

		// Add ID to blacklist until token expires
		if err := rdb.Set(context.Background(), claim.Id, claim.Id, time.Until(claim.ExpiresAt.Time)).Err(); err != nil {
			log.Println("Unable to set blacklist: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		objID, _ := primitive.ObjectIDFromHex(claim.Id)
		_, err := users.DeleteOne(context.Background(), bson.M{"_id": objID})
		if err != nil {
			log.Println("Unable to delete user: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
