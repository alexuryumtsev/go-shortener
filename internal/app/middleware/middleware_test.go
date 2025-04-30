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

		result := rec.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusOK, result.StatusCode)
		assert.NotEmpty(t, result.Cookies())
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

		result := rec.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusOK, result.StatusCode)
	})

	t.Run("With Invalid Cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		cookie := &http.Cookie{
			Name:  "auth_token",
			Value: "invalid-token",
		}
		req.AddCookie(cookie)
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		result := rec.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusOK, result.StatusCode)
	})

	t.Run("Token Generation Failure", func(t *testing.T) {
		// Create a failing mock service
		failingMockService := user.NewMockUserService("")
		failingMiddleware := AuthMiddleware(failingMockService, mockHandler)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		failingMiddleware.ServeHTTP(rec, req)

		result := rec.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusOK, result.StatusCode)
	})
}
