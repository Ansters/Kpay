package main

import (
	"kpay/api"
	"context"
	"log"
	"os"
	"time"

	"github.com/mongodb/mongo-go-driver/mongo"
)

func main() {

	client, err := mongo.NewClient("mongodb") //add your mongodb url
	if err != nil {
		log.Fatal(err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}

	api.StartServer(":"+os.Getenv("PORT"), client)
}