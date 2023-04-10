package main

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func createNewUserClient(clients *mongo.Collection, email string, password string, firstName string, lastName string, groups []string) error {
	// Hash user password before storing in database
	hashedPassword, err := hashAndSalt(password)
	if err != nil {
		return err
	}

	// Create user
	user := &Client{
		Id:             primitive.NewObjectID(),
		Email:          email,
		FirstName:      firstName,
		LastName:       lastName,
		HashedPassword: hashedPassword,
		Suspended:      false,
		Groups:         groups,
	}

	// Insert user into database
	_, err = clients.InsertOne(context.Background(), user)
	if err != nil {
		return err
	}
	return nil
}

func createNewServiceClient(clients *mongo.Collection, email string, name string, groups []string) error {
	// Create service
	service := &Client{
		Id:        primitive.NewObjectID(),
		Email:     email,
		Name:      name,
		Suspended: false,
		Groups:    groups,
	}

	// Insert service into database
	_, err := clients.InsertOne(context.Background(), service)
	if err != nil {
		return err
	}
	return nil
}

func getClientByEmail(clients *mongo.Collection, email string) (*Client, error) {
	filter := bson.D{{Key: "email", Value: email}}
	client := &Client{}
	err := clients.FindOne(context.Background(), filter).Decode(client)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func clientExists(clients *mongo.Collection, email string) bool {
	filter := bson.D{{Key: "email", Value: email}}
	err := clients.FindOne(context.Background(), filter).Err()
	return err == nil
}
