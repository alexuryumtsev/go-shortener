package user

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// UserService интерфейс для работы с пользователями
type UserService interface {
	GetUserIDFromCookie(r *http.Request) string
	GenerateUserToken() (string, error)
	VerifyUserToken(token string) (string, error)
}

// userService реализация UserService
type userService struct {
	secretKey []byte
}

// NewUserService конструктор UserService
func NewUserService(secretKey string) UserService {
	return &userService{secretKey: []byte(secretKey)}
}

// GenerateUserID генерирует случайный ID пользователя
func GenerateUserID() string {
	id := uuid.New()
	return id.String()
}

// GenerateUserToken создает JWT-токен для пользователя
func (u *userService) GenerateUserToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": GenerateUserID(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(u.secretKey)
}

// VerifyUserToken проверяет и извлекает user_id из токена
func (u *userService) VerifyUserToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return u.secretKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["user_id"].(string); ok {
			return userID, nil
		}
	}
	return "", errors.New("invalid token claims")
}

// GetUserIDFromCookie получает user_id из куки
func (u *userService) GetUserIDFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return ""
	}
	userID, err := u.VerifyUserToken(cookie.Value)
	if err != nil {
		return ""
	}
	return userID
}
