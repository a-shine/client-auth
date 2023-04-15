package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get and declare env variables package wide
var dbHost = os.Getenv("DB_HOST")
var dbPort = os.Getenv("DB_PORT")
var dbUser = os.Getenv("DB_USER")
var dbPassword = os.Getenv("DB_PASSWORD")
var dbName = os.Getenv("DB_NAME")
var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
var jwtTokenExpiration, _ = time.ParseDuration(os.Getenv("JWT_TOKEN_EXP_MIN") + "m")

// getClientCollection returns a MongoDB collection for the client collection.
// This is a blocking call that will retry every 5 seconds until a connection
// is established.
func getClientCollection() *mongo.Collection {
	// Construct a connection string to the database
	mongoUri := "mongodb://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort
	clientOptions := options.Client().ApplyURI(mongoUri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalln("Unable to create a database client: ", err)
	}

	// Retry every 5 seconds
	for {
		err = client.Ping(context.Background(), nil)
		if err != nil {
			log.Println("Warning could not connect to database: ", err, "\r\nRetrying...")
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	// Get or create users collection
	return client.Database(dbName).Collection("users")
}

// getCache returns a Redis client used to cache for blacklisted (suspended)
// clients and client deletion pub-sub. This is a blocking call that will retry
// every 5 seconds until a connection is established.
func getCache() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0, // use default DB
	})

	// Retry every 5 seconds
	for {
		_, err := rdb.Ping(context.Background()).Result()
		if err != nil {
			log.Println("Warning could not connect to Redis: ", err, "\r\nRetrying...")
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	return rdb
}

// createHandler creates a gin handler with all the service routes.
func createHandler(clients *mongo.Collection, rdb *redis.Client) *gin.Engine {
	// create an http handler
	handler := gin.Default()
	validate := validator.New()

	// Add logging middleware
	handler.Use(gin.Logger())

	// Both Services and Users are clients
	// - Users have a temporary JWT token which is set as a browser cookie that
	//   needs to be refreshed (more secure for users but difficult to interact
	//   programmatically)
	// - Services have persistent JWT token (don't expire) that remains
	//   available to identify the service (decoded by the gateway)
	handler.POST("/register-user", makeUserRegistrationHandler(clients, validate))
	handler.POST("/register-service", makeServiceRegistrationHandler(clients, validate))

	// User browser login specific routes
	handler.POST("/login", makeLoginHandler(clients, validate))
	handler.GET("/refresh-user-token", makeRefreshHandler(clients))
	handler.POST("/logout", makeLogoutHandler()) // https://stackoverflow.com/questions/3521290/logout-get-or-post

	// These routes have to authenticate and authorize the client
	handler.GET("/me", makeMeHandler(clients))
	handler.POST("/delete", makeDeleteUserHandler(clients, rdb))
	handler.POST("/suspend", makeSuspendClient(clients, rdb))

	return handler
}

func main() {
	log.Println("Connecting to user database...")
	users := getClientCollection()

	log.Println("Connecting to user cache...")
	rdb := getCache()

	handler := createHandler(users, rdb)

	log.Println("Starting server on port 8000...")
	handler.Run(":8000")
}
