package main

import (
	"github.com/segmentio/kafka-go"
	"go-next-ts-chat/internal/user"
	"go-next-ts-chat/internal/ws"
	"go-next-ts-chat/router"
	"log"

	_ "github.com/lib/pq"
	"go-next-ts-chat/db"
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
		Addr:     kafka.TCP(ws.Broker1Address, ws.Broker2Address, ws.Broker3Address),
		Topic:    ws.Topic,
		Balancer: &kafka.LeastBytes{},
	}
	hub := ws.NewHub(kw)
	wsHandler := ws.NewHandler(hub)

	go hub.Run()

	router.InitRouter(userHandler, wsHandler)
	router.Start("0.0.0.0:8080")
}
