package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectDB() *mongo.Database {
	mongoURI := os.Getenv("MONGO_URL")

	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
		fmt.Println("⚠️  MONGO_URL not set — using localhost")
	}

	clientOptions := options.Client().ApplyURI(mongoURI)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("❌ MongoDB connection failed:", err)
	}

	DB = client.Database("event_planner_demo")
	fmt.Println("✅ Connected to MongoDB!")

	return DB
}
