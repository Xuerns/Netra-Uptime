package websocket

import (
	"PulseWatch/internal/model"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool { return true },
}

type WsHandler struct {
	Hub *Hub
	db  *gorm.DB
}

func NewWsHandler(Hub *Hub, db *gorm.DB) *WsHandler {
	return &WsHandler{Hub: Hub, db: db}
}

func (h *WsHandler) ServeWS(c *gin.Context) {
	userId, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var members []model.OrganizationMember
	if err := h.db.Where("user_id = ?", userId.(string)).Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load organization"})
		return
	}

	var orgIDs []string
	for _, member := range members {
		orgIDs = append(orgIDs, member.OrganizationID.String())
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		log.Println("Upgrade Error: ", err)
		return
	}

	client := &Client{
		Hub:    h.Hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: userId.(string),
		OrgIDs: orgIDs,
	}
	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
