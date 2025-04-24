package user

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserService(t *testing.T) {
	secretKey := "test-secret-key"
	service := NewUserService(secretKey)

	t.Run("GenerateUserID", func(t *testing.T) {
		id1 := GenerateUserID()
		id2 := GenerateUserID()

		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2)
	})

	t.Run("GenerateUserToken", func(t *testing.T) {
		token, err := service.GenerateUserToken()
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token structure
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		require.NoError(t, err)
		require.True(t, parsedToken.Valid)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		require.True(t, ok)
		assert.NotEmpty(t, claims["user_id"])
		assert.NotEmpty(t, claims["exp"])
	})

	t.Run("VerifyUserToken", func(t *testing.T) {
		tests := []struct {
			name      string
			token     string
			wantError bool
		}{
			{
				name:      "Valid token",
				token:     createTestToken(t, secretKey, GenerateUserID(), time.Hour),
				wantError: false,
			},
			{
				name:      "Expired token",
				token:     createTestToken(t, secretKey, GenerateUserID(), -time.Hour),
				wantError: true,
			},
			{
				name:      "Invalid signature",
				token:     createTestToken(t, "wrong-secret", GenerateUserID(), time.Hour),
				wantError: true,
			},
			{
				name:      "Invalid token format",
				token:     "invalid-token",
				wantError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				userID, err := service.VerifyUserToken(tt.token)
				if tt.wantError {
					assert.Error(t, err)
					assert.Empty(t, userID)
				} else {
					assert.NoError(t, err)
					assert.NotEmpty(t, userID)
				}
			})
		}
	})

	t.Run("GetUserIDFromCookie", func(t *testing.T) {
		tests := []struct {
			name        string
			setupCookie func() *http.Request
			wantUserID  bool
		}{
			{
				name: "Valid cookie",
				setupCookie: func() *http.Request {
					req := &http.Request{Header: make(http.Header)}
					token := createTestToken(t, secretKey, "test-user", time.Hour)
					req.AddCookie(&http.Cookie{
						Name:  "auth_token",
						Value: token,
					})
					return req
				},
				wantUserID: true,
			},
			{
				name: "No cookie",
				setupCookie: func() *http.Request {
					return &http.Request{Header: make(http.Header)}
				},
				wantUserID: false,
			},
			{
				name: "Invalid cookie value",
				setupCookie: func() *http.Request {
					req := &http.Request{Header: make(http.Header)}
					req.AddCookie(&http.Cookie{
						Name:  "auth_token",
						Value: "invalid-token",
					})
					return req
				},
				wantUserID: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := tt.setupCookie()
				userID := service.GetUserIDFromCookie(req)
				if tt.wantUserID {
					assert.NotEmpty(t, userID)
				} else {
					assert.Empty(t, userID)
				}
			})
		}
	})
}

// Helper function to create test tokens
func createTestToken(t *testing.T, secret string, userID string, expiration time.Duration) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expiration).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return tokenString
}
