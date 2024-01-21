package router

import (
	"github.com/dexguitar/wskfkchat/internal/hub"
	"github.com/dexguitar/wskfkchat/internal/user"
	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func InitRouter(userHandler *user.Handler, hubHandler *hub.Handler) {
	r = gin.Default()

	r.POST("/signup", userHandler.CreateUser)
	r.POST("/login", userHandler.Login)
	r.GET("/logout", userHandler.Logout)

	r.POST("/ws/createRoom", hubHandler.CreateRoom)
	r.GET("/ws/joinRoom/:roomName", hubHandler.JoinRoom)
	r.GET("/ws/getRooms", hubHandler.GetRooms)
	r.GET("/ws/getClients/:roomId", hubHandler.GetClients)
}

func Start(addr string) error {
	return r.Run(addr)
}
