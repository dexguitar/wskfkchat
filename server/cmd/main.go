package main

import (
	"github.com/dexguitar/wskfkchat/internal/user"
	"github.com/dexguitar/wskfkchat/internal/ws"
	"github.com/dexguitar/wskfkchat/router"
	"github.com/segmentio/kafka-go"
	"log"

	"github.com/dexguitar/wskfkchat/db"
	_ "github.com/lib/pq"
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
