package main

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id             primitive.ObjectID `bson:"_id"`
	Email          string             `bson:"email"`
	FirstName      string             `bson:"first_name"`
	LastName       string             `bson:"last_name"`
	HashedPassword string             `bson:"hashed_password"`
	Suspended      bool               `bson:"suspended"`
	Admin          bool               `bson:"admin"`
}
