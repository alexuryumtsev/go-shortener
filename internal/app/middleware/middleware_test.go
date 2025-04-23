package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mockUserService := user.NewMockUserService("test-user")
	middleware := AuthMiddleware(mockUserService, mockHandler)

	t.Run("No Cookie Present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Result().Cookies())
	})

	t.Run("With Valid Cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		cookie := &http.Cookie{
			Name:  "auth_token",
			Value: "mock-token",
		}
		req.AddCookie(cookie)
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
