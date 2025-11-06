package database

import (
	"context"
	"os"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var Client *mongo.Client

func Connect() error {
	mongoUri := os.Getenv("MONGO_URI")
	//mongoUri := "mongodb://localhost:27017/"
	connectionString := options.Client().ApplyURI(mongoUri)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(connectionString)
	if err != nil {
		log.Println("Mongo Connect error:", err)
		return err
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Println("Mongo Ping error:", err)
		return err
	}

	Client = client
	log.Println("MongoDB connected successfully")
	return nil
}
