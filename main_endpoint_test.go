package main

import (
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
)

var clients *mongo.Collection
var rdb *redis.Client
var handler *gin.Engine

func genRandomEmail() string {
	gofakeit.Seed(0)
	return gofakeit.Email()
}

func TestMain(m *testing.M) {
	// Get MongoDB collection and Redis client
	clients = getClientCollection()
	rdb = getCache()

	// Get handler object
	handler = createHandler(clients, rdb)

	// Run tests
	m.Run()
}
