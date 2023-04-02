package main

import "go.mongodb.org/mongo-driver/bson/primitive"

// At a bare minimum any authenticated client (including people/users, services,
// and bots) must have an email address and a hashed password.

type Client struct {
	// Required fields for all clients
	Id        primitive.ObjectID `bson:"_id," json:"-"`
	Email     string             `bson:"email" json:"email"`
	Suspended bool               `bson:"suspended" json:"-"`
	Groups    []string           `bson:"groups" json:"groups"`

	// If the client is a person, then the following fields are required
	FirstName      string `bson:"firstName, omitempty" json:"firstName"`
	LastName       string `bson:"lastName, omitempty" json:"lastName"`
	HashedPassword string `bson:"hashedPassword, omitempty" json:"-"`

	// If the client is a service, then the following fields are required
	Name string `bson:"name, omitempty" json:"name"`
}
