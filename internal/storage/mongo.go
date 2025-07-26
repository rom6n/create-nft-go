package storage

import (
	"log"
	"os"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewMongoClient() *mongo.Client {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("Error. Add MongoDB URI to env.")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Error. Connection to MongoDB unseccessfull")
	}
	return client
}
