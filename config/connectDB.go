package config

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/erneap/go-models/converters"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() *mongo.Client {
	user := Config("MONGO_USER")
	passwd := strings.TrimSpace(Config("MONGO_PASSWD"))
	host := Config("MONGO_HOST")
	port := Config("MONGO_PORT")
	uri := Config("MONGO_URI")
	prefix := Config("MONGO_PREFIX")
	if user != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s/?%s", user, passwd, host, port, prefix)
	}
	log.Println(uri)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB")
	return client
}

func SetLogLevel() int {
	answer := 0
	level := Config("LOGLEVEL")
	if level != "" {
		answer = converters.ParseInt(level)
	}
	return answer
}

var DB *mongo.Client = ConnectDB()

var LogLevel int = SetLogLevel()

// get the requested database collection
func GetCollection(client *mongo.Client, dbName, collectionName string) *mongo.Collection {
	collection := client.Database(dbName).Collection(collectionName)
	return collection
}
