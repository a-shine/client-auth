package main

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id             primitive.ObjectID `bson:"_id" json:"-"`
	Email          string             `bson:"email" json:"email"`
	FirstName      string             `bson:"first_name" json:"firstName"`
	LastName       string             `bson:"last_name" json:"lastName"`
	HashedPassword string             `bson:"hashed_password" json:"-"`
	Suspended      bool               `bson:"suspended" json:"-"`
	Admin          bool               `bson:"admin" json:"-"`
}
