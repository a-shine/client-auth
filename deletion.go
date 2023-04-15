package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// BUG: User deletion logic and testing incomplete
// TODO: Document user deletion mechanism

// DO NOT DELETE USERS DIRECTLY FROM THE DATABASE. INSTEAD, USE THE DELETE USER HANDLER (NEEDS TO TRICKLE DOWN TO ALL SERVICES)
// deleteUser handler enables users to request for their data to be deleted. This communicates with the API Gateway
// through a pubsub 'user-delete' channel. The API Gateway will then communicate with each of the services that
// required authentication, so they can handle deletion of the user data they contain.
func makeDeleteUserHandler(users *mongo.Collection, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the JWT token from cookie
		token, _ := c.Cookie("token")
		code, claim := processClaim(token)
		if code != http.StatusOK {
			c.AbortWithStatusJSON(code, gin.H{"message": "Unable to process JWT token"})
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
			c.AbortWithStatusJSON(status, gin.H{"message": "Unable to authenticate and authorise user"})
			return
		}

		// Add ID to blacklist until token expires
		if err := rdb.Set(context.Background(), claim.Id, claim.Id, time.Until(claim.ExpiresAt.Time)).Err(); err != nil {
			log.Println("Unable to set Redis key: ", err)
		}

		objID, _ := primitive.ObjectIDFromHex(claim.Id)
		_, err := users.DeleteOne(context.Background(), bson.M{"_id": objID})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Unable to delete user"})
			return
		}
	}
}
