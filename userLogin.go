package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

// LoginForm describes the expected JSON payload when a user logs in
type LoginForm struct {
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

func makeLoginHandler(users *mongo.Collection, validate *validator.Validate) gin.HandlerFunc {
	return func(c *gin.Context) {
		var form LoginForm

		// Bind the JSON payload to the form
		if err := c.ShouldBindJSON(&form); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON payload"})
			return
		}

		// Validate the form against the schema
		if err := validate.Struct(form); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// Get the user details from the database
		user, err := getClientByEmail(users, form.Email)
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "User not found"})
			return
		}

		// Compare the provided password with the stored hashed password, if they do not match return an "Unauthorized"
		// status and an incorrect password message
		validPass := verifyPassword(user.HashedPassword, form.Password)
		if !validPass {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Incorrect password"})
			return
		}

		// Check if user is suspended. If suspended, do not issue a token
		if user.Suspended {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "User account suspended"})
			return
		}

		// Declare the expiration time of the token as determined by the jwtTokenExpiration variable
		expirationTime := time.Now().Add(jwtTokenExpiration)

		// Create the JWT claims, which includes the authenticated user ID and expiry time
		claims := &Claim{
			Id:     user.Id.Hex(),
			Groups: user.Groups,
			RegisteredClaims: jwt.RegisteredClaims{
				// In JWT, the expiry time is expressed as unix milliseconds
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}

		// Create the token with the HS256 algorithm used for signing, and the created claim
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Create the JWT string
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			// If there is an error in creating the JWT return an internal server error
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Unable to create token"})
			return
		}

		// Finally, we set the client cookie for "token" as the JWT we just generated we also set an expiry time which is
		// the same as the token itself
		c.SetCookie("token", tokenString, int(expirationTime.Unix()), "/", "localhost", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Successfully logged in"})
	}
}

func makeLogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Clear the token cookie by setting cookie to expiry now
		c.SetCookie("token", "", 0, "/", "localhost", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
	}
}

// https://stackoverflow.com/questions/3487991/why-does-oauth-v2-have-both-access-and-refresh-tokens
func makeRefreshHandler(users *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement refresh handler
		// Try understand refresh token vs access token and how to implement refresh token
		c.JSON(http.StatusOK, gin.H{"message": "Token refresh not possible yet"})
	}
}
