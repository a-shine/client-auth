package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (

	// collections variables
	usersCollection *mongo.Collection

	databaseName = ""
	mongoURI     = ""
	database     *mongo.Database
)

// ----------------------------
// 		TEST MAIN FUNCTION
// ----------------------------

// func TestMain(m *testing.M) {
// 	mongoServer, err := strikememongo.StartWithOptions(&strikememongo.Options{MongoVersion: "6.0.5", DownloadURL: "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-ubuntu2204-6.0.5.tgz", LogLevel: strikememongolog.LogLevelDebug})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	mongoURI = mongoServer.URIWithRandomDB()
// 	splitedDatabaseName := strings.Split(mongoURI, "/")
// 	databaseName = splitedDatabaseName[len(splitedDatabaseName)-1]

// 	defer mongoServer.Stop()

// 	setup()
// 	m.Run()
// }

// ----------------------------
//
//	SET UP FUNCTION
//
// ----------------------------
func setup() {
	startApplication()
	createCollections()
	cleanup()
}

// createCollections cretaes the necessary collections to be used during tests
func createCollections() {
	err := database.CreateCollection(context.Background(), "users")
	if err != nil {
		fmt.Printf("error creating collection: %s", err.Error())
	}

	usersCollection = database.Collection("users")
}

// startApplication initializes the engine and the necessary components for the (test) service to work
func startApplication() {
	// Initialize Database (memongodb)
	dbClient, ctx, err := initDB()
	if err != nil {
		log.Fatal("error connecting to database", err)
	}

	err = dbClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("error connecting to database", err)
	}

	database = dbClient.Database(databaseName)
}

func initDB() (client *mongo.Client, ctx context.Context, err error) {
	uri := fmt.Sprintf("%s%s", mongoURI, "?retryWrites=false")
	client, err = mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return
	}

	return
}

// ----------------------------
//
//	TEAR DOWN FUNCTION
//
// ----------------------------
func cleanup() {
	usersCollection.DeleteMany(context.Background(), bson.M{})
}
