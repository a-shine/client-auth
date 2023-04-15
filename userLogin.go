package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

// LoginForm describes the expected json payload when a user logs in
type LoginForm struct {
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

// login handler for user login
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

// BUG: This refresh logic is not correct
// Refresh is called by frontend when 401 is received it is a way of authenticating user and extending their session without the need for them to input there credentials again but also checks that they are still authorised to use the system
// refresh handler enabling authenticated non-suspended users to apply for new token lengthening their session. // If
// unable to authenticate user or is not authorised (e.g. suspended) then do not refresh token
func makeRefreshHandler(users *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the JWT string from the cookie
		oldToken, err := c.Cookie("token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unable to refresh token"})
			return
		}

		code, claim := processClaim(oldToken)
		if code != http.StatusOK {
			c.AbortWithStatusJSON(code, gin.H{"message": "Unable to refresh token"})
		}

		code, _ = authAndAuthorised(users, claim)
		if code != http.StatusOK {
			c.AbortWithStatusJSON(code, gin.H{"message": "Unable to refresh token"})
			return
		}

		// We ensure that a new token is not issued until enough time has elapsed. In this case, a new token will only be
		// issued if the old token is within 30 seconds of expiry. Otherwise, return a bad request status.
		if time.Until(claim.ExpiresAt.Time) > 30*time.Second {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Token not expired"})
			return
		}

		// Now, create a new token for the current use, with a renewed expiration time
		expirationTime := time.Now().Add(5 * time.Minute)
		claim.ExpiresAt = jwt.NewNumericDate(expirationTime)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
		newToken, err := token.SignedString(jwtKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Unable to refresh token"})
			return
		}

		// Set the new token as the users `token` cookie
		c.SetCookie("token", newToken, int(expirationTime.Unix()), "/", "localhost", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Successfully refreshed token"})
	}
}

// logout handler for user logout
func makeLogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Immediately clear the token cookie by setting cookie expiry to now
		c.SetCookie("token", "", 0, "/", "localhost", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
	}
}
