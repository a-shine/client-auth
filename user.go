package main

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id             primitive.ObjectID `bson:"_id" json:"-"`
	Email          string             `bson:"email" json:"email"`
	FirstName      string             `bson:"firstName" json:"firstName"`
	LastName       string             `bson:"lastName" json:"lastName"`
	HashedPassword string             `bson:"hashedPassword" json:"-"`
	Suspended      bool               `bson:"suspended" json:"-"`
	Groups         []string           `bson:"groups" json:"groups"`
}

type Service struct {
	Id        primitive.ObjectID `bson:"_id" json:"-"`
	Name      string             `bson:"name" json:"name"`
	Key       string             `bson:"key" json:"key"` // API key
	Suspended bool               `bson:"suspended" json:"suspended"`
	Groups    []string           `bson:"groups" json:"groups"`
}
