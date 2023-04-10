package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

// LoginForm describes the expected json payload when a user logs in
type LoginForm struct {
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

// login handler for user login
func makeLoginHandler(users *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds LoginForm

		// Get the JSON body and decode into credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			// If the structure of the body is wrong, return an HTTP error
			w.WriteHeader(http.StatusBadRequest)
			log.Println(w.Write([]byte(`{"message":"Invalid request payload"}`)))
			return
		}

		// Get the user details from the database
		user, err := getClientByEmail(users, creds.Email)
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusUnauthorized)
			log.Println(w.Write([]byte(`{"message":"No account registered with this email"}`)))
			return
		}

		// Compare the provided password with the stored hashed password, if they do not match return an "Unauthorized"
		// status and an incorrect password message
		validPass := verifyPassword(user.HashedPassword, creds.Password)
		if !validPass {
			w.WriteHeader(http.StatusUnauthorized)
			log.Println(w.Write([]byte(`{"message":"Incorrect password"}`)))
			return
		}

		// Check if user is suspended. If suspended, do not issue a token
		if user.Suspended {
			w.WriteHeader(http.StatusUnauthorized)
			log.Println(w.Write([]byte(`{"message":"Account has been suspended"}`)))
			return
		}

		// Declare the expiration time of the token as determined by the maxJwtTokenExpiration variable
		expirationTime := time.Now().Add(maxJwtTokenExpiration)

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
			log.Println("Unable to sign token: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Finally, we set the client cookie for "token" as the JWT we just generated we also set an expiry time which is
		// the same as the token itself
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
			Path:    "/",
		})

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Login successful"}`))
	}
}

// Refresh is called by frontend when 401 is received it is a way of authenticating user and extending their session without the need for them to input there credentials again but also checks that they are still authorised to use the system
// refresh handler enabling authenticated non-suspended users to apply for new token lengthening their session. // If
// unable to authenticate user or is not authorised (e.g. suspended) then do not refresh token
func makeRefreshHandler(users *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code, claim := processClaim(r)
		if code != http.StatusOK {
			w.WriteHeader(code)
			log.Println(w.Write([]byte(`{"message": "Unable to process JWT token"}`)))
			return
		}

		code, _ = authAndAuthorised(users, claim)
		if code != http.StatusOK {
			w.WriteHeader(code)
			log.Println(w.Write([]byte(`{"message": "Unable to refresh token"}`)))
			return
		}

		// We ensure that a new token is not issued until enough time has elapsed. In this case, a new token will only be
		// issued if the old token is within 30 seconds of expiry. Otherwise, return a bad request status.
		if time.Until(claim.ExpiresAt.Time) > 30*time.Second {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(w.Write([]byte(`{"message": "Token still valid"}`)))
			return
		}

		// Now, create a new token for the current use, with a renewed expiration time
		expirationTime := time.Now().Add(5 * time.Minute)
		claim.ExpiresAt = jwt.NewNumericDate(expirationTime)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Set the new token as the users `token` cookie
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
			Path:    "/",
		})
	}
}

// logout handler for user logout
func makeLogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// Immediately clear the token cookie by setting cookie expiry to now
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Expires: time.Now(),
			Path:    "/",
		})
	}
}
