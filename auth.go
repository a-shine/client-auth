package main

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Claim describes the structure of a JWT claim (this is the same payload as in the
// https://github.com/a-shine/api-gateway repo which is what makes them compatible)
type Claim struct {
	Id     string   `json:"id"`
	Groups []string `json:"groups"` // TODO: Use this to check if user is admin, allows groups to be verified by the API gateway
	jwt.RegisteredClaims
}

// verifyPassword compares a hashed password with the raw password
func verifyPassword(hashedPwd string, plainPwd string) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(plainPwd))
	return err == nil // true if the err is nil and false otherwise
}

func processClaim(token string) (int, *Claim) {
	claim := &Claim{}

	// c, err := r.Cookie("token")
	// if err != nil {
	// 	if err == http.ErrNoCookie {
	// 		return http.StatusUnauthorized, nil
	// 	}
	// 	return http.StatusBadRequest, nil
	// }
	// tknStr := c.Value

	tkn, err := jwt.ParseWithClaims(token, claim, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return http.StatusUnauthorized, nil
		}
		return http.StatusBadRequest, nil
	}
	if !tkn.Valid {
		return http.StatusUnauthorized, nil
	}

	return http.StatusOK, claim
}

func authenticate(users *mongo.Collection, claim *Claim) (int, *Client) {
	user := &Client{}

	// Get user by the ID in the token claim payload
	objID, _ := primitive.ObjectIDFromHex(claim.Id)
	value := users.FindOne(context.Background(), bson.M{"_id": objID}).Decode(user)

	// Check user if user can be authenticated
	if value == mongo.ErrNoDocuments {
		return http.StatusUnauthorized, user
	} else {
		return http.StatusOK, user
	}
}

func authAndAuthorised(users *mongo.Collection, claim *Claim) (int, *Client) {
	code, user := authenticate(users, claim)
	switch code {
	case http.StatusOK:
		if user.Suspended {
			return http.StatusUnauthorized, nil
		} else {
			return http.StatusOK, user
		}
	default:
		return code, nil
	}
}

func authAndAuthorisedAdmin(users *mongo.Collection, claim *Claim) (int, *Client) {
	code, user := authAndAuthorised(users, claim)
	switch code {
	case http.StatusOK:
		// If user group contains "admin" then they are authorised
		for _, group := range user.Groups {
			if group == "admin" {
				return http.StatusOK, user
			}
		}
		return http.StatusUnauthorized, nil
	default:
		return code, nil
	}
}

// Needs to be a jwt token so that the API gateway can verify it
func generateAPIClientToken(client *Client) (string, error) {
	// Create the JWT claims, which includes the user ID with no expiration time
	claims := &Claim{
		Id:               client.Id.Hex(),
		Groups:           client.Groups,
		RegisteredClaims: jwt.RegisteredClaims{},
	}

	// Create the JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtKey)
	return tokenStr, err
}
