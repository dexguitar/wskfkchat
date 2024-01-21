package hub

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/http"
)

type Handler struct {
	hub *Hub
}

func NewHandler(h *Hub) *Handler {
	return &Handler{
		hub: h,
	}
}

type CreateRoomReq struct {
	Name string `json:"name" bson:"name"`
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var req CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please provide room name"})
		return
	}

	roomID := uuid.New().String()
	err := h.hub.MongoRepo.CreateRoom(c, roomID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.hub.Rooms[req.Name] = &Room{
		ID:      roomID,
		Name:    req.Name,
		Clients: make(map[string]*Client),
	}

	c.JSON(http.StatusOK, req)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  0,
	WriteBufferSize: 0,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) JoinRoom(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	roomName := c.Param("roomName")
	_, err = h.hub.MongoRepo.GetRoomByName(c, roomName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: addClient to room in mongo

	clientID := c.Query("userId")
	username := c.Query("username")

	cl := &Client{
		Conn:     conn,
		Message:  make(chan *Message, 10),
		ID:       clientID,
		RoomName: roomName,
		Username: username,
	}

	m := &Message{
		Content:  "👋 user joined the room",
		RoomName: roomName,
		Username: username,
	}

	//	Register a new client through the register channel
	h.hub.Register <- cl
	//	Broadcast client's message
	h.hub.Produce(c, m)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go cl.ReadSocket(h.hub)
	cl.Consume(c)
}

type RoomRes struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *Handler) GetRooms(c *gin.Context) {
	//rooms := make([]RoomRes, 0)
	rooms := h.hub.MongoRepo.GetRooms(c)

	//for _, r := range h.hub.Rooms {
	//	rooms = append(rooms, RoomRes{
	//		ID:   r.ID,
	//		Name: r.Name,
	//	})
	//}

	c.JSON(http.StatusOK, rooms)
}

type ClientRes struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (h *Handler) GetClients(c *gin.Context) {
	var clients []ClientRes
	roomId := c.Param("roomId")

	if _, ok := h.hub.Rooms[roomId]; !ok {
		clients := make([]ClientRes, 0)
		c.JSON(http.StatusOK, clients)
	}

	for _, c := range h.hub.Rooms[roomId].Clients {
		clients = append(clients, ClientRes{
			ID:       c.ID,
			Username: c.Username,
		})
	}

	c.JSON(http.StatusOK, clients)
}
