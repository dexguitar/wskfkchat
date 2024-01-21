package hub

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
)

type Room struct {
	ID      string             `json:"id"`
	Name    string             `json:"name"`
	Clients map[string]*Client `json:"clients"`
}

type Hub struct {
	Rooms      map[string]*Room
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *Message
	Writer     *kafka.Writer
	MongoRepo  *MongoRepo
}

func NewHub(writer *kafka.Writer, mongo *MongoRepo) *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *Message, 5),
		Writer:     writer,
		MongoRepo:  mongo,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case cl := <-h.Register:
			if _, ok := h.Rooms[cl.RoomName]; ok {
				r := h.Rooms[cl.RoomName]

				if _, ok := r.Clients[cl.ID]; !ok {
					r.Clients[cl.ID] = cl
				}
			}
		case cl := <-h.Unregister:
			if _, ok := h.Rooms[cl.RoomName]; ok {
				if _, ok := h.Rooms[cl.RoomName].Clients[cl.ID]; ok {
					// produce a message saying that the client has left the room
					if len(h.Rooms[cl.RoomName].Clients) != 0 {
						h.Produce(context.Background(), &Message{
							Content:  fmt.Sprintf("❌ user left the room"),
							RoomName: cl.RoomName,
							Username: cl.Username,
						})
					}

					delete(h.Rooms[cl.RoomName].Clients, cl.ID)
					close(cl.Message)
				}
			}
		}
	}
}

func (h *Hub) Produce(ctx context.Context, m *Message) {
	err := h.Writer.WriteMessages(ctx, kafka.Message{
		Key:     []byte(m.RoomName),
		Value:   []byte(m.Content),
		Headers: []kafka.Header{{Key: "username", Value: []byte(m.Username)}},
	})
	if err != nil {
		log.Printf("could not write message: %v", err)
	}

	//// save to mongo
	//err = h.MongoRepo.SaveMessage(ctx, m)
	//if err != nil {
	//	return err
	//}
	//
	//return nil
}
