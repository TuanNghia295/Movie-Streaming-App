package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect() *mongo.Client {
	err := godotenv.Load(".env")

	if err != nil {
		log.Println("Warning: unable to find .env file")
	}

	MongoDB := os.Getenv("MONGODB_URI")
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://tuannghiait2905_db_user:h7AjpdlDthXkG1fO@cluster0.4amosxs.mongodb.net/?appName=Cluster0").SetServerAPIOptions(serverAPI)

	if MongoDB == "" {
		log.Fatal("MONGODB_URI not set!")
	}

	fmt.Println("MONGODB_URI: ", MongoDB)

	client, err := mongo.Connect(opts)

	if err != nil {
		return nil
	}

	return client
}

var Client *mongo.Client = Connect()

func OpenCollection(collectionName string) *mongo.Collection {
	err := godotenv.Load(".env")

	if err != nil {
		log.Println("Wanring: unable to find .env file")
	}

	databaseName := os.Getenv("DATABASE_NAME")

	fmt.Println("DATABASE_NAME: ", databaseName)

	collection := Client.Database(databaseName).Collection(collectionName)

	if collection == nil {
		return nil
	}

	return collection

}
