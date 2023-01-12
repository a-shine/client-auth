package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Get env variables and make available package wide
var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
var maxJwtTokenExpiration time.Duration
var dbHost = os.Getenv("DB_HOST")
var dbPort = os.Getenv("DB_PORT")
var dbUser = os.Getenv("DB_USER")
var dbPassword = os.Getenv("DB_PASSWORD")
var dbName = os.Getenv("DB_NAME")

// Package wide variable
var users *mongo.Collection
var ctx = context.Background()
var rdb *redis.Client

func main() {
	// Get max expiration time of JWT token
	mins, err := strconv.Atoi(os.Getenv("JWT_TOKEN_EXP_MIN"))
	if err != nil {
		log.Fatalln("Invalid JWT_TOKEN_EXP_MIN env variable value")
	}
	maxJwtTokenExpiration = time.Duration(mins) * time.Minute

	log.Println("Connecting to user database...")

	// Construct a connection string to the database
	mongoUri := "mongodb://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort
	clientOptions := options.Client().ApplyURI(mongoUri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalln("Unable to create a database client: ", err)
	}

	// Exponential backoff retry for DB connection
	for {
		wait := time.Duration(2)
		err = client.Ping(ctx, nil)
		if err != nil {
			log.Println("Warning could not connect to database: ", err, "\r\nRetrying...")
			time.Sleep(wait * time.Second)
			wait = wait * 2
		}
		break
	}

	// Get or create users collection
	users = client.Database(dbName).Collection("users")

	log.Println("Connecting to user cache...")
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0, // use default DB
	})

	// Exponential backoff retry for Redis connection
	for {
		wait := time.Duration(2)
		_, err = rdb.Ping(ctx).Result()
		if err != nil {
			log.Println("Warning could not connect to Redis: ", err, "\r\nRetrying...")
			time.Sleep(wait * time.Second)
			wait = wait * 2
		}
		break
	}

	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/refresh", refresh)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/me", me)
	http.HandleFunc("/delete", deleteUser)
	http.HandleFunc("/suspend", suspendUser)

	log.Println("Starting server on port 8000...")
	log.Fatalln(http.ListenAndServe(":8000", nil))
}
