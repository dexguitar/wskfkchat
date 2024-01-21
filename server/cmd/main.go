package main

import (
	"context"
	"github.com/dexguitar/wskfkchat/db"
	"github.com/dexguitar/wskfkchat/internal/hub"
	"github.com/dexguitar/wskfkchat/internal/user"
	"github.com/dexguitar/wskfkchat/router"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func main() {
	dbConn, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("could not initialize database connection: %s", err)
	}

	userRep := user.NewRepository(dbConn.GetDB())
	userSvc := user.NewService(userRep)
	userHandler := user.NewHandler(userSvc)

	kw := &kafka.Writer{
		Addr:     kafka.TCP(hub.Broker1Address, hub.Broker2Address, hub.Broker3Address),
		Topic:    hub.Topic,
		Balancer: &kafka.LeastBytes{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	if err = mongoClient.Ping(ctx, nil); err != nil {
		panic(err)
	}

	mongoRepo := hub.NewMongoRepo(mongoClient)
	h := hub.NewHub(kw, mongoRepo)
	hubHandler := hub.NewHandler(h)

	go h.Run()

	router.InitRouter(userHandler, hubHandler)
	router.Start("0.0.0.0:8080")
}
