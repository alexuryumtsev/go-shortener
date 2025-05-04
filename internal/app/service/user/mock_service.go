package user

import "net/http"

// MockUserService - это структура, которая реализует интерфейс UserService
type MockUserService struct {
	userID string
}

// NewMockUserService создает новый экземпляр MockUserService с заданным userID.
func NewMockUserService(userID string) *MockUserService {
	return &MockUserService{userID: userID}
}

// GetUserIDFromCookie - это метод, который возвращает userID из cookie.
func (m *MockUserService) GetUserIDFromCookie(r *http.Request) string {
	return m.userID
}

// GenerateUserToken - это метод, который генерирует токен для пользователя.
func (m *MockUserService) GenerateUserToken() (string, error) {
	return "mock-token", nil
}

// VerifyUserToken - это метод, который проверяет токен пользователя и возвращает userID.
func (m *MockUserService) VerifyUserToken(token string) (string, error) {
	return m.userID, nil
}
