package ws

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
	"log"
)

type Client struct {
	Conn     *websocket.Conn
	Message  chan *Message
	ID       string `json:"id"`
	RoomID   string `json:"roomId"`
	Username string `json:"username"`
}

type Message struct {
	Content  string `json:"content"`
	RoomID   string `json:"roomId"`
	Username string `json:"username"`
}

func (c *Client) readSocket(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, m, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		msg := &Message{
			Content:  string(m),
			RoomID:   c.RoomID,
			Username: c.Username,
		}

		hub.Produce(context.Background(), msg)
	}
}

func (c *Client) Consume(ctx context.Context) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{Broker1Address, Broker2Address, Broker3Address},
		Topic:   Topic,
	})

	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("could not read message: %v", err)
		}

		// separating messages by room
		if string(msg.Key) == c.RoomID {
			m := &Message{
				Content:  string(msg.Value),
				RoomID:   string(msg.Key),
				Username: string(msg.Headers[0].Value),
			}

			c.Conn.WriteJSON(m)
		}
	}
}
