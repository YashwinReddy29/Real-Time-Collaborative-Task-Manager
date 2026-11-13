package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPasswordRecoveryUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &AuthHandler{}
	router := gin.New()

	router.POST("/api/auth/forgot-password", handler.ForgotPassword)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/forgot-password",
		strings.NewReader(`{"email":"person@example.test"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"got HTTP %d, want 503: %s",
			rec.Code,
			rec.Body.String(),
		)
	}

	if strings.Contains(rec.Body.String(), "token") {
		t.Fatal("password reset response exposed token information")
	}
}
