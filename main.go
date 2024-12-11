package main

import (
	"context"
	"dannyroman2015/phoebe/internal/app"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongo:rzLmDKylubzBEngsuxZTvuqgfFxXFxVM@roundhouse.proxy.rlwy.net:49073"))
	if err != nil {
		panic(err)
	}
	mgdb := client.Database("phoebe")

	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	} else {
		port = ":" + port
	}

	server := app.NewServer(port, mgdb)
	server.Start()
}
