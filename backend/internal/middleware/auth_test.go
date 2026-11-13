package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"taskflow/internal/auth"

	"github.com/gin-gonic/gin"
)

func TestAuthenticationTokenSources(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const secret = "test-only-secret-at-least-32-characters"

	token, err := auth.GenerateToken(
		"user-123",
		"user@example.test",
		"Test User",
		secret,
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		websocket  bool
		header     string
		queryToken string
		wantStatus int
	}{
		{
			name:       "REST accepts bearer token",
			header:     "Bearer " + token,
			wantStatus: http.StatusOK,
		},
		{
			name:       "REST rejects query token",
			queryToken: token,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "WebSocket accepts query token",
			websocket:  true,
			queryToken: token,
			wantStatus: http.StatusOK,
		},
		{
			name:       "WebSocket accepts bearer token",
			websocket:  true,
			header:     "Bearer " + token,
			wantStatus: http.StatusOK,
		},
		{
			name:       "REST rejects missing token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "REST rejects malformed bearer header",
			header:     "Basic " + token,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "WebSocket rejects invalid token",
			websocket:  true,
			queryToken: "invalid-token",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()

			var middleware gin.HandlerFunc
			if tc.websocket {
				middleware = WebSocketAuthMiddleware(secret)
			} else {
				middleware = AuthMiddleware(secret)
			}

			router.GET("/test", middleware, func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(
				http.MethodGet, "/test", nil,
			)

			if tc.queryToken != "" {
				q := req.URL.Query()
				q.Set("token", tc.queryToken)
				req.URL.RawQuery = q.Encode()
			}

			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf(
					"HTTP %d, want %d: %s",
					rec.Code,
					tc.wantStatus,
					rec.Body.String(),
				)
			}
		})
	}
}
