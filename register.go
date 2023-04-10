package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserRegistrationForm describes the expected json payload when a user registers
type UserRegistrationForm struct {
	Password  string   `json:"password" validate:"required,min=8,max=64"`
	Email     string   `json:"email" validate:"required,email"`
	FirstName string   `json:"firstName" validate:"required"`
	LastName  string   `json:"lastName" validate:"required"`
	Groups    []string `json:"groups" validate:"required"`
}

type ServiceRegistrationForm struct {
	Email  string   `json:"email" validate:"required,email"`
	Name   string   `json:"name" validate:"required"`
	Groups []string `json:"groups" validate:"required"`
}

// makeUserRegistrationHandler registers handler function for client registration endpoint
// If client is a user a password, first name and last name are required
// If a client is a service (IoT device) only an email is required (the email serves as the unique identifier)
func makeUserRegistrationHandler(users *mongo.Collection, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds UserRegistrationForm

		// Get the JSON body and decode into credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			// If the structure of the body is wrong, return an HTTP error
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"Invalid JSON payload"}`))
			return
		}

		// validate json schema
		err = validate.Struct(creds)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"Invalid JSON payload"}`))
			return
		}

		// Check if user already exists
		filter := bson.D{{Key: "email", Value: creds.Email}}
		mongoErr := users.FindOne(context.Background(), filter).Err()

		if mongoErr != mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusConflict)
			log.Println(w.Write([]byte(`{"message":"An account with that email address is already registered"}`)))
			return
		}

		// Hash user password before storing in database
		hashedPassword, err := hashAndSalt(creds.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Create user
		user := &Client{
			Id:             primitive.NewObjectID(),
			Email:          creds.Email,
			FirstName:      creds.FirstName,
			LastName:       creds.LastName,
			HashedPassword: hashedPassword,
			Suspended:      false,
			Groups:         creds.Groups,
		}

		// Insert user into database
		_, err = users.InsertOne(context.Background(), user)
		if err != nil {
			log.Println("Unable to insert user into database: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(w.Write([]byte(`{"message":"Unable to register user"}`)))
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message":"User registered successfully"}`))
	}
}

// makeUserRegistrationHandler handles the registration of a new service. A
// service is a type of client, unlike a user, that interacts programmatically
// with the backend. A service is not a user and therefore does not have a
// password, first name, last name, etc. The handler returns a JWT token that
// does not expire and is used to authenticate the service.
func makeServiceRegistrationHandler(clients *mongo.Collection, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds ServiceRegistrationForm

		// Get the JSON body and decode into credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			// If the structure of the body is wrong, return an HTTP error
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"Invalid JSON payload"}`))
			return
		}

		// Validate JSON schema
		err = validate.Struct(creds)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"Invalid payload fields"}`))
			return
		}

		// Check if service already registered
		filter := bson.D{{Key: "email", Value: creds.Email}}
		mongoErr := clients.FindOne(context.Background(), filter).Err()
		if mongoErr != mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusConflict)
			log.Println(w.Write([]byte(`{"message":"An account with that email address is already registered"}`)))
			return
		}

		// Create new client
		service := &Client{
			Id:        primitive.NewObjectID(),
			Email:     creds.Email,
			Name:      creds.Name,
			Groups:    creds.Groups,
			Suspended: false,
		}

		// Insert user into database
		_, err = clients.InsertOne(context.Background(), service)
		if err != nil {
			log.Println("Unable to insert service into database: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(w.Write([]byte(`{"message":"Unable to register user"}`)))
		}

		// If client exists generate a new token
		// BUG: Potentially not returning the correct api token
		apiToken := generateAPIClientToken(service)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`"serviceToken": "` + apiToken.Raw + `"`))
	}
}
