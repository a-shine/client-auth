package main

import "go.mongodb.org/mongo-driver/bson/primitive"

// At a bare minimum any authenticated client (including people/users, services,
// and bots) must have an email address and a hashed password.

type Client struct {
	Id             primitive.ObjectID `bson:"_id" json:"-"`
	Email          string             `bson:"email" json:"email" validate:"required,email"`
	HashedPassword string             `bson:"hashedPassword" json:"-" validate:"required"`

	Suspended bool     `bson:"suspended" json:"-"`
	Groups    []string `bson:"groups" json:"groups"`

	// If the client is a person, then the following fields are required.
	FirstName string `bson:"firstName" json:"firstName"`
	LastName  string `bson:"lastName" json:"lastName"`
}
