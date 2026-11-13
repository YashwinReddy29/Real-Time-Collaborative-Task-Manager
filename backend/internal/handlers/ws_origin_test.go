package handlers

import (
	"net/http"
	"testing"
)

func TestWebSocketOriginPolicy(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		allowed bool
	}{
		{"same origin", "http://localhost:3000", true},
		{"different port", "http://localhost:5173", false},
		{"different host", "http://evil.example", false},
		{"different scheme", "https://localhost:3000", false},
		{"missing origin", "", false},
		{"lookalike hostname", "http://localhost:3000.evil.example", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(
				http.MethodGet,
				"http://localhost:8080/ws/board1",
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			got := isAllowedWebSocketOrigin(
				req,
				"http://localhost:3000",
			)

			if got != tt.allowed {
				t.Fatalf(
					"origin=%q: got %v, want %v",
					tt.origin, got, tt.allowed,
				)
			}
		})
	}
}
