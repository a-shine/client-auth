package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
func makeUserRegistrationHandler(clients *mongo.Collection, validate *validator.Validate) gin.HandlerFunc {
	return func(c *gin.Context) {
		var form UserRegistrationForm

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

		// Check if user already exists
		if clientExists(clients, form.Email) {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"message": "A user associated with the email address is already registered"})
			return
		}

		if err := createNewUserClient(clients, form.Email, form.Password, form.FirstName, form.LastName, form.Groups); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Unable to register user"})
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
	}
}

// makeUserRegistrationHandler handles the registration of a new service. A
// service is a type of client, unlike a user, that interacts programmatically
// with the backend. A service is not a user and therefore does not have a
// password, first name, last name, etc. The handler returns a JWT token that
// does not expire and is used to authenticate the service.
func makeServiceRegistrationHandler(clients *mongo.Collection, validate *validator.Validate) gin.HandlerFunc {
	return func(c *gin.Context) {
		var form ServiceRegistrationForm

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

		// Check if service already registered
		if clientExists(clients, form.Email) {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"message": "A service associated with the email address is already registered"})
			return
		}

		// Create new client
		if err := createNewServiceClient(clients, form.Email, form.Name, form.Groups); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Unable to register service"})
		}

		// TODO: Return a JWT token (that doesn't expire) to the client so that
		//  it can be used to authenticate JWTs
		// apiToken := generateAPIClientToken(service)

		c.JSON(http.StatusCreated, gin.H{"message": "Service registered successfully"})
	}
}
