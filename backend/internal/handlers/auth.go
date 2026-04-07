package handlers

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"time"
	"taskflow/internal/auth"
	"taskflow/internal/repository"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo  *repository.UserRepo
	jwtSecret string
}

func NewAuthHandler(userRepo *repository.UserRepo, jwtSecret string) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user, err := h.userRepo.Create(c.Request.Context(), req.Name, req.Email, string(hashed))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Name, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Name, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		// Don't reveal whether email exists
		c.JSON(http.StatusOK, gin.H{"message": "If that email exists, a reset link has been sent"})
		return
	}

	// Generate token
	token := generateResetToken()
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = h.userRepo.CreatePasswordReset(c.Request.Context(), token, user.ID, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create reset token"})
		return
	}

	// In production this would send an email
	// For dev/portfolio: print to terminal
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", "http://localhost:5173", token)
	fmt.Printf("\n🔑 PASSWORD RESET LINK for %s:\n%s\n\n", user.Email, resetURL)

	c.JSON(http.StatusOK, gin.H{"message": "If that email exists, a reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.userRepo.ValidateResetToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired reset token"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := h.userRepo.UpdatePassword(c.Request.Context(), userID, string(hashed)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	if err := h.userRepo.MarkResetTokenUsed(c.Request.Context(), req.Token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to invalidate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func generateResetToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
