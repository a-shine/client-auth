package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
)

// May be worth thinking more about how the suspend account logic should work.
// Maybe don't have it submitted by an admin but simply let the user suspend
// his own account?

// SuspendForm describes the expected json payload when a user suspension request is made
type SuspendForm struct {
	Id string `json:"id" validate:"required"`
}

// suspendUser is an only admin accessible handler for suspending a user
func makeSuspendClient(users *mongo.Collection, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var form SuspendForm

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
			c.AbortWithStatusJSON(status, gin.H{"message": "You are not authorised to perform this action"})
			return
		}

		// Now that we have validate that the user making the request is an admin, we can suspend the user he is requesting to suspend
		if err := c.ShouldBindJSON(&form); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Unable to parse JSON payload"})
			return
		}

		rdb.Set(context.Background(), form.Id, form.Id, jwtTokenExpiration)
		c.JSON(http.StatusOK, gin.H{"message": "Successfully suspended user"})
	}
}
