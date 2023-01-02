package main

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id             primitive.ObjectID `bson:"_id"`
	Email          string             `bson:"email"`
	FirstName      string             `bson:"first_name"`
	LastName       string             `bson:"last_name"`
	HashedPassword string             `bson:"hashed_password"` // do not return password
	Active         bool               `bson:"active"`          // do not return active status
}
