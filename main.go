package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get env variables and make available package wide
var dbHost = os.Getenv("DB_HOST")
var dbPort = os.Getenv("DB_PORT")
var dbUser = os.Getenv("DB_USER")
var dbPassword = os.Getenv("DB_PASSWORD")
var dbName = os.Getenv("DB_NAME")
var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

var jwtTokenExpiration time.Duration

func setTokenExpirationDuration() {
	// Get max expiration time of JWT token from env variable and convert to
	// integer
	mins, err := strconv.Atoi(os.Getenv("JWT_TOKEN_EXP_MIN"))
	if err != nil {
		log.Fatalln("Invalid JWT_TOKEN_EXP_MIN env variable value")
	}
	jwtTokenExpiration = time.Duration(mins) * time.Minute
}

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

func createHandler(clients *mongo.Collection, rdb *redis.Client) *gin.Engine {
	// create an http handler
	handler := gin.Default()
	validate := validator.New()

	// Add logging middleware
	handler.Use(gin.Logger())

	// Both Services and Users are clients
	// - Users have a temporary JWT token which is set as a browser cookie that needs to be refreshed (more secure for users but difficult to interact programmatically)
	// - Service have persistent JWT token (doesn't expire) that remains available to identify the service (decoded by the gateway)

	handler.POST("/register-user", makeUserRegistrationHandler(clients, validate))
	handler.POST("/register-service", makeServiceRegistrationHandler(clients, validate))

	// User browser login specific routes
	handler.POST("/login", makeLoginHandler(clients, validate)) // only to register users (generates temporary tokens while gen-api-token generates persistent tokens)
	handler.GET("/refresh-user-token", makeRefreshHandler(clients))
	handler.POST("/logout", makeLogoutHandler()) // https://stackoverflow.com/questions/3521290/logout-get-or-post

	handler.GET("/me", makeMeHandler(clients))
	handler.POST("/delete", makeDeleteUserHandler(clients, rdb))
	handler.POST("/suspend", makeSuspendClient(clients, rdb))

	return handler
}

func main() {
	setTokenExpirationDuration()

	log.Println("Connecting to user database...")
	users := getClientCollection()

	log.Println("Connecting to user cache...")
	rdb := getCache()

	handler := createHandler(users, rdb)
	log.Println("Starting server on port 8000...")
	// log.Fatalln(http.ListenAndServe(":8000", handler))
	handler.Run(":8000")
}
