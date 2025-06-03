package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPCheckMiddleware(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name          string
		trustedSubnet string
		clientIP      string
		wantStatus    int
	}{
		{
			name:          "Valid IP in subnet",
			trustedSubnet: "192.168.1.0/24",
			clientIP:      "192.168.1.100",
			wantStatus:    http.StatusOK,
		},
		{
			name:          "Invalid IP outside subnet",
			trustedSubnet: "192.168.1.0/24",
			clientIP:      "10.0.0.1",
			wantStatus:    http.StatusForbidden,
		},
		{
			name:          "Empty trusted subnet",
			trustedSubnet: "",
			clientIP:      "192.168.1.100",
			wantStatus:    http.StatusForbidden,
		},
		{
			name:          "No X-Real-IP header",
			trustedSubnet: "192.168.1.0/24",
			clientIP:      "",
			wantStatus:    http.StatusForbidden,
		},
		{
			name:          "Invalid IP format",
			trustedSubnet: "192.168.1.0/24",
			clientIP:      "invalid-ip",
			wantStatus:    http.StatusBadRequest,
		},
		{
			name:          "Valid IP at subnet boundary",
			trustedSubnet: "10.0.0.0/8",
			clientIP:      "10.255.255.255",
			wantStatus:    http.StatusOK,
		},
		{
			name:          "IPv6 in IPv4 subnet",
			trustedSubnet: "192.168.1.0/24",
			clientIP:      "::1",
			wantStatus:    http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := IPCheckMiddleware(tt.trustedSubnet, mockHandler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.clientIP != "" {
				req.Header.Set("X-Real-IP", tt.clientIP)
			}

			rec := httptest.NewRecorder()
			middleware.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestIPCheckMiddleware_InvalidSubnet(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test with invalid CIDR notation
	middleware := IPCheckMiddleware("invalid-cidr", mockHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")

	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
