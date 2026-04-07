package handlers

import (
	"taskflow/internal/models"
	ws "taskflow/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func (h *BoardHandler) HandleWebSocket(c *gin.Context) {
	userID := c.GetString("user_id")
	userName := c.GetString("user_name")
	boardID := c.Param("boardId")

	isMember, _ := h.boardRepo.IsMember(c.Request.Context(), boardID, userID)
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &ws.Client{
		ID:      uuid.New().String(),
		UserID:  userID,
		BoardID: boardID,
		Conn:    conn,
		Send:    make(chan []byte, 256),
	}

	h.hub.Register(client)

	// Broadcast join
	h.hub.Broadcast(&models.WSEvent{
		Type:    models.EventUserJoined,
		Payload: gin.H{"user_id": userID, "user_name": userName},
		BoardID: boardID,
		UserID:  userID,
	})
	h.broadcastPresence(boardID)

	go client.WritePump()

	// Block until disconnect
	client.ReadPump(h.hub)

	// Broadcast leave
	h.hub.Broadcast(&models.WSEvent{
		Type:    models.EventUserLeft,
		Payload: gin.H{"user_id": userID, "user_name": userName},
		BoardID: boardID,
		UserID:  userID,
	})
	h.broadcastPresence(boardID)
}
