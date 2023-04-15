package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
)

// BUG: User suspension and testing logic not tested

// SuspendForm describes the expected json payload when a user suspension request is made
type SuspendForm struct {
	Id string `json:"id"`
}

// suspendUser is an only admin accessible handler for suspending a user
func makeSuspendClient(users *mongo.Collection, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var suspendUser SuspendForm

		// Add ID to blacklist until token expires
		// Get the JWT token from cookie
		token, _ := c.Cookie("token")
		code, claim := processClaim(token)
		if code != http.StatusOK {
			c.AbortWithStatusJSON(code, gin.H{"message": "Unable to process JWT token"})
			return
		}

		status, _ := authAndAuthorisedAdmin(users, claim)

		if status != http.StatusOK {
			// Get the JSON body and decode into credentials
			c.AbortWithStatusJSON(status, gin.H{"message": "Unable to authenticate and authorise user"})
		}

		rdb.Set(context.Background(), suspendUser.Id, suspendUser.Id, jwtTokenExpiration)
	}
}
