package handlers

import (
	"net/http"
	"taskflow/internal/models"
	"taskflow/internal/repository"
	ws "taskflow/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type BoardHandler struct {
	userRepo     *repository.UserRepo
	boardRepo    *repository.BoardRepo
	columnRepo   *repository.ColumnRepo
	taskRepo     *repository.TaskRepo
	activityRepo *repository.ActivityRepo
	hub          *ws.Hub
	upgrader     websocket.Upgrader
}

func NewBoardHandler(
	boardRepo *repository.BoardRepo,
	columnRepo *repository.ColumnRepo,
	taskRepo *repository.TaskRepo,
	activityRepo *repository.ActivityRepo,
	userRepo *repository.UserRepo,
	hub *ws.Hub,
	frontendURL string,
) *BoardHandler {
	return &BoardHandler{
		boardRepo:    boardRepo,
		columnRepo:   columnRepo,
		taskRepo:     taskRepo,
		activityRepo: activityRepo,
		userRepo:     userRepo,
		hub:          hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *BoardHandler) ListBoards(c *gin.Context) {
	userID := c.GetString("user_id")
	boards, err := h.boardRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if boards == nil {
		boards = []models.Board{}
	}
	c.JSON(http.StatusOK, boards)
}

func (h *BoardHandler) CreateBoard(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	board, err := h.boardRepo.Create(c.Request.Context(), req.Name, req.Description, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, board)
}

func (h *BoardHandler) GetBoard(c *gin.Context) {
	userID := c.GetString("user_id")
	boardID := c.Param("boardId")

	isMember, err := h.boardRepo.IsMember(c.Request.Context(), boardID, userID)
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	board, err := h.boardRepo.GetByID(c.Request.Context(), boardID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "board not found"})
		return
	}

	columns, err := h.columnRepo.GetByBoardID(c.Request.Context(), boardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if columns == nil {
		columns = []models.Column{}
	}

	tasks, err := h.taskRepo.GetByBoardID(c.Request.Context(), boardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tasks == nil {
		tasks = []models.Task{}
	}

	members, err := h.boardRepo.GetMembers(c.Request.Context(), boardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"board":   board,
		"columns": columns,
		"tasks":   tasks,
		"members": members,
	})
}

func (h *BoardHandler) CreateColumn(c *gin.Context) {
	userID := c.GetString("user_id")
	userName := c.GetString("user_name")
	boardID := c.Param("boardId")

	isMember, _ := h.boardRepo.IsMember(c.Request.Context(), boardID, userID)
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req struct {
		Name     string `json:"name" binding:"required"`
		Position int    `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	col, err := h.columnRepo.Create(c.Request.Context(), boardID, req.Name, req.Position)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.hub.Broadcast(&models.WSEvent{
		Type:    models.EventColumnCreated,
		Payload: col,
		BoardID: boardID,
		UserID:  userID,
	})

	h.logActivity(c, boardID, userID, userName, "created column: "+req.Name)
	c.JSON(http.StatusCreated, col)
}

func (h *BoardHandler) DeleteColumn(c *gin.Context) {
	userID := c.GetString("user_id")
	boardID := c.Param("boardId")
	columnID := c.Param("columnId")

	isMember, _ := h.boardRepo.IsMember(c.Request.Context(), boardID, userID)
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.columnRepo.Delete(c.Request.Context(), columnID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.hub.Broadcast(&models.WSEvent{
		Type:    models.EventColumnDeleted,
		Payload: gin.H{"column_id": columnID},
		BoardID: boardID,
		UserID:  userID,
	})

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BoardHandler) CreateTask(c *gin.Context) {
	userID := c.GetString("user_id")
	userName := c.GetString("user_name")
	boardID := c.Param("boardId")

	isMember, _ := h.boardRepo.IsMember(c.Request.Context(), boardID, userID)
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req struct {
		ColumnID    string `json:"column_id" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
		Position    int    `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Priority == "" {
		req.Priority = "medium"
	}

	task, err := h.taskRepo.Create(
		c.Request.Context(),
		req.ColumnID, boardID, req.Title, req.Description, req.Priority, req.Position,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.hub.Broadcast(&models.WSEvent{
		Type:    models.EventTaskCreated,
		Payload: task,
		BoardID: boardID,
		UserID:  userID,
	})

	h.logActivity(c, boardID, userID, userName, "created task: "+req.Title)
	c.JSON(http.StatusCreated, task)
}

func (h *BoardHandler) UpdateTask(c *gin.Context) {
	userID := c.GetString("user_id")
	userName := c.GetString("user_name")
	boardID := c.Param("boardId")
	taskID := c.Param("taskId")

	isMember, _ := h.boardRepo.IsMember(c.Request.Context(), boardID, userID)
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskRepo.Update(c.Request.Context(), taskID, req.Title, req.Description, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.hub.Broadcast(&models.WSEvent{
		Type:    models.EventTaskUpdated,
		Payload: task,
		BoardID: boardID,
		UserID:  userID,
	})

	h.logActivity(c, boardID, userID, userName, "updated task: "+req.Title)
	c.JSON(http.StatusOK, task)
}

func (h *BoardHandler) MoveTask(c *gin.Context) {
	userID := c.GetString("user_id")
	userName := c.GetString("user_name")
	boardID := c.Param("boardId")
	taskID := c.Param("taskId")

	isMember, _ := h.boardRepo.IsMember(c.Request.Context(), boardID, userID)
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req struct {
		ColumnID string `json:"column_id" binding:"required"`
		Position int    `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskRepo.Move(c.Request.Context(), taskID, req.ColumnID, req.Position)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.hub.Broadcast(&models.WSEvent{
		Type:    models.EventTaskMoved,
		Payload: task,
		BoardID: boardID,
		UserID:  userID,
	})

	h.logActivity(c, boardID, userID, userName, "moved a task")
	c.JSON(http.StatusOK, task)
}

func (h *BoardHandler) DeleteTask(c *gin.Context) {
	userID := c.GetString("user_id")
	userName := c.GetString("user_name")
	boardID := c.Param("boardId")
	taskID := c.Param("taskId")

	isMember, _ := h.boardRepo.IsMember(c.Request.Context(), boardID, userID)
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	task, _ := h.taskRepo.GetByID(c.Request.Context(), taskID)

	if err := h.taskRepo.Delete(c.Request.Context(), taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.hub.Broadcast(&models.WSEvent{
		Type:    models.EventTaskDeleted,
		Payload: gin.H{"task_id": taskID},
		BoardID: boardID,
		UserID:  userID,
	})

	if task != nil {
		h.logActivity(c, boardID, userID, userName, "deleted task: "+task.Title)
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BoardHandler) GetActivity(c *gin.Context) {
	userID := c.GetString("user_id")
	boardID := c.Param("boardId")

	isMember, _ := h.boardRepo.IsMember(c.Request.Context(), boardID, userID)
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	activities, err := h.activityRepo.GetByBoardID(c.Request.Context(), boardID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if activities == nil {
		activities = []models.Activity{}
	}
	c.JSON(http.StatusOK, activities)
}

func (h *BoardHandler) logActivity(c *gin.Context, boardID, userID, userName, action string) {
	activity, err := h.activityRepo.Create(c.Request.Context(), boardID, userID, userName, action)
	if err != nil {
		return
	}
	h.hub.Broadcast(&models.WSEvent{
		Type:    models.EventActivity,
		Payload: activity,
		BoardID: boardID,
		UserID:  userID,
	})
}

func (h *BoardHandler) broadcastPresence(boardID string) {
	presence := h.hub.GetPresence(boardID)
	h.hub.Broadcast(&models.WSEvent{
		Type:    "presence.update",
		Payload: presence,
		BoardID: boardID,
	})
}

func (h *BoardHandler) InviteMember(c *gin.Context) {
	userID := c.GetString("user_id")
	boardID := c.Param("boardId")

	// Only members can invite
	isMember, _ := h.boardRepo.IsMember(c.Request.Context(), boardID, userID)
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	invitedUser, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no user found with that email"})
		return
	}

	// Check if already a member
	alreadyMember, _ := h.boardRepo.IsMember(c.Request.Context(), boardID, invitedUser.ID)
	if alreadyMember {
		c.JSON(http.StatusConflict, gin.H{"error": "user is already a member"})
		return
	}

	// Add to board
	if err := h.boardRepo.AddMember(c.Request.Context(), boardID, invitedUser.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userName := c.GetString("user_name")
	h.logActivity(c, boardID, userID, userName, "invited "+invitedUser.Name+" to the board")

	c.JSON(http.StatusOK, gin.H{
		"user": invitedUser,
		"message": "member added successfully",
	})
}
