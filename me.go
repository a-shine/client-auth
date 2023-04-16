package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// me handler for user details
func makeMeHandler(users *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the JWT token from cookie
		token, err := c.Cookie("token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unable to get JWT token from cookie"})
			return
		}

		code, claim := processClaim(token)
		if code != http.StatusOK {
			c.AbortWithStatusJSON(code, gin.H{"message": "Unable to process JWT token"})
			return
		}

		status, user := authAndAuthorised(users, claim)
		if status != http.StatusOK {
			c.AbortWithStatusJSON(status, gin.H{"message": "Unable to authenticate and authorise user"})
			return
		}

		// Return the user details
		c.JSON(http.StatusOK, gin.H{"user": user})
	}
}
