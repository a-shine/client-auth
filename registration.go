package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
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
func makeUserRegistrationHandler(clients *mongo.Collection, validate *validator.Validate) http.HandlerFunc {
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
		if clientExists(clients, creds.Email) {
			w.WriteHeader(http.StatusConflict)
			log.Println(w.Write([]byte(`{"message":"An account with that email address is already registered"}`)))
			return
		}

		err = createNewUserClient(clients, creds.Email, creds.Password, creds.FirstName, creds.LastName, creds.Groups)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"Unable to register user"}`))
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
		if clientExists(clients, creds.Email) {
			w.WriteHeader(http.StatusConflict)
			log.Println(w.Write([]byte(`{"message":"An account with that email address is already registered"}`)))
			return
		}

		// Create new client
		err = createNewServiceClient(clients, creds.Email, creds.Name, creds.Groups)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"Unable to register service"}`))
		}

		// If client exists generate a new token
		// BUG: Potentially not returning the correct api token
		// apiToken := generateAPIClientToken(service)

		w.WriteHeader(http.StatusCreated)
		// w.Write([]byte(`"serviceToken": "` + apiToken.Raw + `"`))
	}
}
